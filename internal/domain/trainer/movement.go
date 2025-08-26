package trainer

import (
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// MovementDirection represents movement direction
type MovementDirection struct {
	X, Y float64 // Direction vector (-1, 0, 1)
}

// MovementState represents trainer's current movement state
type MovementState struct {
	Direction   MovementDirection `json:"direction"`    // Current direction
	Speed       float64           `json:"speed"`        // Units per second
	StartTime   time.Time         `json:"start_time"`   // When movement started
	StartPos    shared.Position   `json:"start_pos"`    // Position when movement started
	IsMoving    bool              `json:"is_moving"`    // Whether currently moving
}

// NewMovementState creates a new movement state
func NewMovementState() MovementState {
	return MovementState{
		Direction: MovementDirection{X: 0, Y: 0},
		Speed:     2.0, // Default 2 units per second
		IsMoving:  false,
	}
}

// StartMovement starts movement in given direction
func (ms *MovementState) StartMovement(direction MovementDirection, currentPos shared.Position) {
	ms.Direction = direction
	ms.StartTime = time.Now()
	ms.StartPos = currentPos
	ms.IsMoving = true
}

// StopMovement stops current movement
func (ms *MovementState) StopMovement(currentPos shared.Position) {
	ms.IsMoving = false
	ms.StartPos = currentPos
	ms.Direction = MovementDirection{X: 0, Y: 0}
}

// CalculateCurrentPosition calculates current position based on time elapsed
func (ms *MovementState) CalculateCurrentPosition() shared.Position {
	if !ms.IsMoving {
		return ms.StartPos
	}

	elapsed := time.Since(ms.StartTime).Seconds()
	distance := ms.Speed * elapsed

	newX := ms.StartPos.X + (ms.Direction.X * distance)
	newY := ms.StartPos.Y + (ms.Direction.Y * distance)

	return shared.Position{X: newX, Y: newY}
}