package models

import "time"
type Location struct {
	UserID    string  `json:"userId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type Session struct {
	SessionID  string    `json:"sessionId"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	UserCount  int       `json:"userCount"`
}

type LocationUpdate struct {
	Type string   `json:"type"`
	Data Location `json:"data"`
}

type AllLocations struct {
	Type      string     `json:"type"`
	Data      []Location `json:"data"`
	SessionID string     `json:"sessionId"`
}

type CreateSessionRequest struct {
	UserID    string  `json:"userId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type UpdateLocationRequest struct {
	SessionID string  `json:"sessionId"`
	UserID    string  `json:"userId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
