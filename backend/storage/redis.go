package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"maplocationshare/backend/models"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStorage(addr string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &RedisStorage{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}

func (r *RedisStorage) CreateSession(sessionID string, initialLocation models.Location) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionData := models.Session{
		SessionID: sessionID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserCount: 1,
	}

	sessionJSON, _ := json.Marshal(sessionData)
	pipe := r.client.Pipeline()

	pipe.Set(r.ctx, sessionKey, sessionJSON, 24*time.Hour)

	locationKey := fmt.Sprintf("session:%s:locations", sessionID)
	locationJSON, _ := json.Marshal(initialLocation)
	pipe.HSet(r.ctx, locationKey, initialLocation.UserID, locationJSON)
	pipe.Expire(r.ctx, locationKey, 24*time.Hour)

	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisStorage) GetSession(sessionID string) (*models.Session, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	data, err := r.client.Get(r.ctx, sessionKey).Result()
	if err != nil {
		return nil, err
	}

	var session models.Session
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RedisStorage) UpdateLocation(sessionID string, location models.Location) error {
	locationKey := fmt.Sprintf("session:%s:locations", sessionID)
	locationJSON, err := json.Marshal(location)
	if err != nil {
		return err
	}

	return r.client.HSet(r.ctx, locationKey, location.UserID, locationJSON).Err()
}

func (r *RedisStorage) GetAllLocations(sessionID string) ([]models.Location, error) {
	locationKey := fmt.Sprintf("session:%s:locations", sessionID)
	data, err := r.client.HGetAll(r.ctx, locationKey).Result()
	if err != nil {
		return nil, err
	}

	var locations []models.Location
	for _, value := range data {
		var location models.Location
		if err := json.Unmarshal([]byte(value), &location); err == nil {
			locations = append(locations, location)
		}
	}

	return locations, nil
}

func (r *RedisStorage) SessionExists(sessionID string) bool {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	exists, _ := r.client.Exists(r.ctx, sessionKey).Result()
	return exists > 0
}
