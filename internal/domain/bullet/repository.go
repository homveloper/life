package bullet

import (
	"context"
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// Repository represents the bullet repository interface
type Repository interface {
	// Save saves a bullet to the repository
	Save(ctx context.Context, bullet *Bullet) error
	
	// Load loads a bullet by ID
	Load(ctx context.Context, bulletID BulletID) (*Bullet, error)
	
	// LoadActive loads all active bullets
	LoadActive(ctx context.Context) ([]*Bullet, error)
	
	// LoadActiveForPlayer loads all active bullets for a specific player
	LoadActiveForPlayer(ctx context.Context, playerID PlayerID) ([]*Bullet, error)
	
	// LoadInArea loads all active bullets within a specific area
	LoadInArea(ctx context.Context, topLeft, bottomRight shared.Position) ([]*Bullet, error)
	
	// Delete removes a bullet from the repository
	Delete(ctx context.Context, bulletID BulletID) error
	
	// DeleteExpired removes all expired bullets (older than specified time)
	DeleteExpired(ctx context.Context, expiredBefore time.Time) (int, error)
	
	// Exists checks if a bullet exists
	Exists(ctx context.Context, bulletID BulletID) (bool, error)
	
	// Count returns the total number of bullets
	Count(ctx context.Context) (int64, error)
	
	// CountActiveForPlayer returns the number of active bullets for a player
	CountActiveForPlayer(ctx context.Context, playerID PlayerID) (int64, error)
	
	// Batch operations for performance optimization
	
	// SaveBatch saves multiple bullets in a single transaction
	SaveBatch(ctx context.Context, bullets []*Bullet) error
	
	// DeleteBatch deletes multiple bullets in a single transaction
	DeleteBatch(ctx context.Context, bulletIDs []BulletID) error
	
	// LoadBatch loads multiple bullets by their IDs
	LoadBatch(ctx context.Context, bulletIDs []BulletID) ([]*Bullet, error)
}

// PlayerStats represents player firing statistics and state
type PlayerStats struct {
	PlayerID            PlayerID   `json:"player_id"`
	AmmoCount          int        `json:"ammo_count"`
	WeaponType         WeaponType `json:"weapon_type"`
	LastFireTime       time.Time  `json:"last_fire_time"`
	FireSessionStarted *time.Time `json:"fire_session_started,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// NewPlayerStats creates new player stats
func NewPlayerStats(playerID PlayerID, ammoCount int, weaponType WeaponType) *PlayerStats {
	return &PlayerStats{
		PlayerID:   playerID,
		AmmoCount:  ammoCount,
		WeaponType: weaponType,
		UpdatedAt:  time.Now(),
	}
}

// CanFire checks if player can fire based on ammo and cooldown
func (ps *PlayerStats) CanFire(currentTime time.Time) bool {
	// Check ammo
	if ps.AmmoCount <= 0 {
		return false
	}
	
	// Check cooldown
	cooldownMs := ps.WeaponType.GetCooldownMs()
	cooldownDuration := time.Duration(cooldownMs) * time.Millisecond
	timeSinceLastFire := currentTime.Sub(ps.LastFireTime)
	
	return timeSinceLastFire >= cooldownDuration
}

// Fire decrements ammo and updates last fire time
func (ps *PlayerStats) Fire(fireTime time.Time) error {
	if ps.AmmoCount <= 0 {
		return shared.NewDomainError(shared.ErrCodeInsufficientFunds, "No ammo remaining")
	}
	
	ps.AmmoCount--
	ps.LastFireTime = fireTime
	ps.UpdatedAt = time.Now()
	
	return nil
}

// StartFireSession starts a continuous fire session
func (ps *PlayerStats) StartFireSession(startTime time.Time) {
	ps.FireSessionStarted = &startTime
	ps.UpdatedAt = time.Now()
}

// StopFireSession stops the continuous fire session
func (ps *PlayerStats) StopFireSession() {
	ps.FireSessionStarted = nil
	ps.UpdatedAt = time.Now()
}

// IsFiring checks if player is currently in a fire session
func (ps *PlayerStats) IsFiring() bool {
	return ps.FireSessionStarted != nil
}

// ReloadAmmo reloads ammo to full capacity
func (ps *PlayerStats) ReloadAmmo(maxAmmo int) {
	ps.AmmoCount = maxAmmo
	ps.UpdatedAt = time.Now()
}


// PlayerStatsRepository represents the player stats repository interface
type PlayerStatsRepository interface {
	// SaveStats saves player firing stats
	SaveStats(ctx context.Context, stats *PlayerStats) error
	
	// LoadStats loads player firing stats
	LoadStats(ctx context.Context, playerID PlayerID) (*PlayerStats, error)
	
	// LoadStatsWithDefaults loads stats or creates default if not found
	LoadStatsWithDefaults(ctx context.Context, playerID PlayerID, defaultAmmo int, defaultWeapon WeaponType) (*PlayerStats, error)
	
	// DeleteStats removes player stats
	DeleteStats(ctx context.Context, playerID PlayerID) error
	
	// StatsExists checks if player stats exist
	StatsExists(ctx context.Context, playerID PlayerID) (bool, error)
	
	// Batch operations for performance optimization
	
	// SaveStatsBatch saves multiple player stats in a single transaction
	SaveStatsBatch(ctx context.Context, stats []*PlayerStats) error
	
	// LoadStatsBatch loads multiple player stats by their IDs
	LoadStatsBatch(ctx context.Context, playerIDs []PlayerID) ([]*PlayerStats, error)
	
	// ValidateFirePermissionsBatch validates fire permissions for multiple players
	ValidateFirePermissionsBatch(ctx context.Context, playerIDs []PlayerID, currentTime time.Time) (map[PlayerID]bool, error)
}

// FireSession represents a continuous firing session
type FireSession struct {
	SessionID   shared.ID  `json:"session_id"`
	PlayerID    PlayerID   `json:"player_id"`
	WeaponType  WeaponType `json:"weapon_type"`
	FireRate    int        `json:"fire_rate"`    // Rounds per minute
	StartedAt   time.Time  `json:"started_at"`
	LastFiredAt *time.Time `json:"last_fired_at,omitempty"`
	BulletsShot int        `json:"bullets_shot"`
	IsActive    bool       `json:"is_active"`
}

// NewFireSession creates a new fire session
func NewFireSession(playerID PlayerID, weaponType WeaponType) *FireSession {
	return &FireSession{
		SessionID:   shared.NewID(),
		PlayerID:    playerID,
		WeaponType:  weaponType,
		FireRate:    weaponType.GetFireRate(),
		StartedAt:   time.Now(),
		BulletsShot: 0,
		IsActive:    true,
	}
}

// RecordShot records a bullet shot in this session
func (fs *FireSession) RecordShot(shotTime time.Time) {
	fs.BulletsShot++
	fs.LastFiredAt = &shotTime
}

// Stop stops the fire session
func (fs *FireSession) Stop() {
	fs.IsActive = false
}

// FireSessionRepository represents the fire session repository interface
type FireSessionRepository interface {
	// SaveSession saves a fire session
	SaveSession(ctx context.Context, session *FireSession) error
	
	// LoadSession loads a fire session by ID
	LoadSession(ctx context.Context, sessionID shared.ID) (*FireSession, error)
	
	// LoadActiveSessionForPlayer loads active session for a player
	LoadActiveSessionForPlayer(ctx context.Context, playerID PlayerID) (*FireSession, error)
	
	// DeleteSession removes a fire session
	DeleteSession(ctx context.Context, sessionID shared.ID) error
	
	// DeleteExpiredSessions removes expired sessions (older than specified time)
	DeleteExpiredSessions(ctx context.Context, expiredBefore time.Time) (int, error)
}