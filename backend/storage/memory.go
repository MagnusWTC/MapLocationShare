package storage

import (
	"sync"
	"time"

	"maplocationshare/backend/models"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	sessions    map[string]*SessionData
	locations   map[string]map[string]*LocationData
	mu          sync.RWMutex
	expiration  time.Duration
	cleanupTick *time.Ticker
}

// SessionData 会话数据
type SessionData struct {
	Session   models.Session
	LastAccess time.Time
}

// LocationData 位置数据
type LocationData struct {
	Location   models.Location
	LastAccess time.Time
}

// NewMemoryStorage 创建新的内存存储实例
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		sessions:    make(map[string]*SessionData),
		locations:   make(map[string]map[string]*LocationData),
		expiration:  24 * time.Hour,
		cleanupTick: time.NewTicker(1 * time.Hour), // 每小时清理一次过期数据
	}

	// 启动定期清理协程
	go storage.cleanupExpired()

	return storage
}

// Close 关闭存储，停止清理协程
func (m *MemoryStorage) Close() error {
	m.cleanupTick.Stop()
	return nil
}

// cleanupExpired 清理过期数据
func (m *MemoryStorage) cleanupExpired() {
	for range m.cleanupTick.C {
		now := time.Now()
		m.mu.Lock()

		// 清理过期会话
		for sessionID, sessionData := range m.sessions {
			if now.Sub(sessionData.LastAccess) > m.expiration {
				delete(m.sessions, sessionID)
				// 同时删除对应的位置数据
				delete(m.locations, sessionID)
			}
		}

		// 清理过期位置数据
		for sessionID, locationMap := range m.locations {
			for userID, locationData := range locationMap {
				if now.Sub(locationData.LastAccess) > m.expiration {
					delete(locationMap, userID)
				}
			}
			// 如果位置数据为空，删除整个会话的位置映射
			if len(locationMap) == 0 {
				delete(m.locations, sessionID)
				// 同时删除对应的会话数据
				delete(m.sessions, sessionID)
			}
		}

		m.mu.Unlock()
	}
}

// CreateSession 创建会话
func (m *MemoryStorage) CreateSession(sessionID string, initialLocation models.Location) error {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建会话
	m.sessions[sessionID] = &SessionData{
		Session: models.Session{
			SessionID: sessionID,
			CreatedAt: now,
			ExpiresAt: now.Add(m.expiration),
			UserCount: 1,
		},
		LastAccess: now,
	}

	// 初始化位置映射
	if _, ok := m.locations[sessionID]; !ok {
		m.locations[sessionID] = make(map[string]*LocationData)
	}

	// 添加初始位置
	m.locations[sessionID][initialLocation.UserID] = &LocationData{
		Location:   initialLocation,
		LastAccess: now,
	}

	return nil
}

// GetSession 获取会话
func (m *MemoryStorage) GetSession(sessionID string) (*models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionData, ok := m.sessions[sessionID]
	if !ok {
		return nil, nil // 会话不存在
	}

	// 更新最后访问时间（滑动过期）
	now := time.Now()
	sessionData.LastAccess = now
	sessionData.Session.ExpiresAt = now.Add(m.expiration)

	return &sessionData.Session, nil
}

// UpdateLocation 更新位置
func (m *MemoryStorage) UpdateLocation(sessionID string, location models.Location) error {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查会话是否存在
	if sessionData, ok := m.sessions[sessionID]; ok {
		// 更新会话最后访问时间
		sessionData.LastAccess = now
		sessionData.Session.ExpiresAt = now.Add(m.expiration)
	} else {
		// 会话不存在，创建新会话
		m.sessions[sessionID] = &SessionData{
			Session: models.Session{
				SessionID: sessionID,
				CreatedAt: now,
				ExpiresAt: now.Add(m.expiration),
				UserCount: 1,
			},
			LastAccess: now,
		}
	}

	// 确保位置映射存在
	if _, ok := m.locations[sessionID]; !ok {
		m.locations[sessionID] = make(map[string]*LocationData)
	}

	// 更新位置
	m.locations[sessionID][location.UserID] = &LocationData{
		Location:   location,
		LastAccess: now,
	}

	return nil
}

// GetAllLocations 获取所有位置
func (m *MemoryStorage) GetAllLocations(sessionID string) ([]models.Location, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查会话是否存在
	if sessionData, ok := m.sessions[sessionID]; ok {
		// 更新会话最后访问时间
		now := time.Now()
		sessionData.LastAccess = now
		sessionData.Session.ExpiresAt = now.Add(m.expiration)
	} else {
		return nil, nil // 会话不存在
	}

	// 收集所有位置
	var locations []models.Location
	if locationMap, ok := m.locations[sessionID]; ok {
		// 更新所有位置的最后访问时间
		now := time.Now()
		for _, locationData := range locationMap {
			locationData.LastAccess = now
			locations = append(locations, locationData.Location)
		}
	}

	return locations, nil
}

// SessionExists 检查会话是否存在
func (m *MemoryStorage) SessionExists(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.sessions[sessionID]
	return ok
}
