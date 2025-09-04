package cqrs

import (
	"time"

	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
)

// TrainerMovedEvent represents a domain event when a trainer moves
type TrainerMovedEvent struct {
	UserID    string                 `json:"user_id"`
	Nickname  string                 `json:"nickname"`
	Color     string                 `json:"color"`
	Position  shared.Position        `json:"position"`
	Movement  trainer.MovementState  `json:"movement"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
}

// TrainerStoppedEvent represents a domain event when a trainer stops moving
type TrainerStoppedEvent struct {
	UserID    string                 `json:"user_id"`
	Nickname  string                 `json:"nickname"`
	Color     string                 `json:"color"`
	Position  shared.Position        `json:"position"`
	Movement  trainer.MovementState  `json:"movement"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
}

// TrainerCreatedEvent represents a domain event when a trainer is created
type TrainerCreatedEvent struct {
	UserID    string           `json:"user_id"`
	Trainer   *trainer.Trainer `json:"trainer"`
	Timestamp time.Time        `json:"timestamp"`
	RequestID string           `json:"request_id"`
}

// SSENotificationEvent represents an event to send SSE notifications
type SSENotificationEvent struct {
	Type        string      `json:"type"`
	TargetUsers []string    `json:"target_users,omitempty"` // UserIDs for array targeting (empty for broadcast)
	Method      string      `json:"method"`
	Params      interface{} `json:"params"`
	Timestamp   time.Time   `json:"timestamp"`
	RequestID   string      `json:"request_id"`
}

// Event types for different notification patterns
const (
	SSENotificationTypeBroadcast = "broadcast" // Send to all users
	SSENotificationTypeUsers     = "users"     // Send to specific list of users
)
