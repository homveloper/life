package bullet

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// Event types
const (
	BulletFiredEventType   = "bullet.fired"
	BulletHitEventType     = "bullet.hit"
	BulletExpiredEventType = "bullet.expired"
)

// BulletFiredEvent represents bullet firing
type BulletFiredEvent struct {
	shared.BaseEvent
}

// BulletFiredEventData holds the event data
type BulletFiredEventData struct {
	BulletID   string          `json:"bullet_id"`
	PlayerID   string          `json:"player_id"`
	WeaponType string          `json:"weapon_type"`
	StartPos   shared.Position `json:"start_position"`
	Direction  Direction       `json:"direction"`
	Velocity   Velocity        `json:"velocity"`
	MaxRange   float64         `json:"max_range"`
	FiredAt    int64           `json:"fired_at"` // Unix timestamp
}

// NewBulletFiredEvent creates a new bullet fired event
func NewBulletFiredEvent(bullet *Bullet) (BulletFiredEvent, error) {
	data := BulletFiredEventData{
		BulletID:   bullet.ID.String(),
		PlayerID:   bullet.PlayerID.String(),
		WeaponType: bullet.WeaponType.String(),
		StartPos:   bullet.StartPos,
		Direction:  bullet.Velocity.Direction,
		Velocity:   bullet.Velocity,
		MaxRange:   bullet.MaxRange,
		FiredAt:    bullet.FiredAt.Unix(),
	}

	baseEvent, err := shared.NewBaseEvent(
		BulletFiredEventType,
		bullet.ID.String(),
		"bullet",
		data,
	)
	if err != nil {
		return BulletFiredEvent{}, err
	}

	return BulletFiredEvent{BaseEvent: baseEvent}, nil
}

// BulletHitEvent represents bullet hitting a target
type BulletHitEvent struct {
	shared.BaseEvent
}

// BulletHitEventData holds the event data
type BulletHitEventData struct {
	BulletID    string          `json:"bullet_id"`
	PlayerID    string          `json:"player_id"`
	HitPosition shared.Position `json:"hit_position"`
	TargetType  string          `json:"target_type,omitempty"` // "player", "animal", "obstacle"
	TargetID    string          `json:"target_id,omitempty"`   // ID of target if applicable
	HitAt       int64           `json:"hit_at"`                // Unix timestamp
}

// NewBulletHitEvent creates a new bullet hit event
func NewBulletHitEvent(bullet *Bullet, targetType, targetID string) (BulletHitEvent, error) {
	data := BulletHitEventData{
		BulletID:    bullet.ID.String(),
		PlayerID:    bullet.PlayerID.String(),
		HitPosition: bullet.CurrentPos,
		TargetType:  targetType,
		TargetID:    targetID,
		HitAt:       bullet.UpdatedAt.Unix(),
	}

	baseEvent, err := shared.NewBaseEvent(
		BulletHitEventType,
		bullet.ID.String(),
		"bullet",
		data,
	)
	if err != nil {
		return BulletHitEvent{}, err
	}

	return BulletHitEvent{BaseEvent: baseEvent}, nil
}

// BulletExpiredEvent represents bullet expiring (max range or time)
type BulletExpiredEvent struct {
	shared.BaseEvent
}

// BulletExpiredEventData holds the event data
type BulletExpiredEventData struct {
	BulletID         string          `json:"bullet_id"`
	PlayerID         string          `json:"player_id"`
	FinalPosition    shared.Position `json:"final_position"`
	DistanceTraveled float64         `json:"distance_traveled"`
	ExpiredBy        string          `json:"expired_by"` // "range" or "time"
	ExpiredAt        int64           `json:"expired_at"` // Unix timestamp
}

// NewBulletExpiredEvent creates a new bullet expired event
func NewBulletExpiredEvent(bullet *Bullet, expiredBy string) (BulletExpiredEvent, error) {
	data := BulletExpiredEventData{
		BulletID:         bullet.ID.String(),
		PlayerID:         bullet.PlayerID.String(),
		FinalPosition:    bullet.CurrentPos,
		DistanceTraveled: bullet.GetDistanceTraveled(),
		ExpiredBy:        expiredBy,
		ExpiredAt:        bullet.UpdatedAt.Unix(),
	}

	baseEvent, err := shared.NewBaseEvent(
		BulletExpiredEventType,
		bullet.ID.String(),
		"bullet",
		data,
	)
	if err != nil {
		return BulletExpiredEvent{}, err
	}

	return BulletExpiredEvent{BaseEvent: baseEvent}, nil
}
