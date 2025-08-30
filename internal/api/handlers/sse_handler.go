package handlers

import (
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
	"time"
)

// TrainerPositionUpdate represents a position update event (legacy, for compatibility)
type TrainerPositionUpdate struct {
	Type      string                `json:"type"`
	TrainerID string                `json:"trainer_id"`
	Nickname  string                `json:"nickname,omitempty"`
	Position  shared.Position       `json:"position"`
	Movement  trainer.MovementState `json:"movement"`
	Timestamp time.Time             `json:"timestamp"`
}