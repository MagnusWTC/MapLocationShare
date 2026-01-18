package storage

import (
	"maplocationshare/backend/models"
)

// Storage 存储接口
type Storage interface {
	// Close 关闭存储
	Close() error
	// CreateSession 创建会话
	CreateSession(sessionID string, initialLocation models.Location) error
	// GetSession 获取会话
	GetSession(sessionID string) (*models.Session, error)
	// UpdateLocation 更新位置
	UpdateLocation(sessionID string, location models.Location) error
	// GetAllLocations 获取所有位置
	GetAllLocations(sessionID string) ([]models.Location, error)
	// SessionExists 检查会话是否存在
	SessionExists(sessionID string) bool
}
