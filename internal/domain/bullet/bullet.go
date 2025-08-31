package bullet

import (
	"math"
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// BulletID represents a unique bullet identifier
type BulletID shared.ID

// NewBulletID creates a new bullet ID
func NewBulletID() BulletID {
	return BulletID(shared.NewID())
}

// String returns string representation
func (id BulletID) String() string {
	return string(id)
}

// PlayerID represents a unique player identifier
type PlayerID shared.ID

// String returns string representation
func (id PlayerID) String() string {
	return string(id)
}

// WeaponType represents different types of weapons
type WeaponType string

const (
	// Pistol types
	BasicPistol    WeaponType = "basic_pistol"
	AdvancedPistol WeaponType = "advanced_pistol"
	
	// Rifle types
	AssaultRifle WeaponType = "assault_rifle"
	SniperRifle  WeaponType = "sniper_rifle"
	
	// Shotgun types
	PumpShotgun  WeaponType = "pump_shotgun"
	AutoShotgun  WeaponType = "auto_shotgun"
)

// String returns string representation
func (wt WeaponType) String() string {
	return string(wt)
}

// IsValid checks if weapon type is valid
func (wt WeaponType) IsValid() bool {
	switch wt {
	case BasicPistol, AdvancedPistol, AssaultRifle, SniperRifle, PumpShotgun, AutoShotgun:
		return true
	default:
		return false
	}
}

// GetFireRate returns the fire rate in rounds per minute
func (wt WeaponType) GetFireRate() int {
	switch wt {
	case BasicPistol:
		return 300 // 300 RPM
	case AdvancedPistol:
		return 450 // 450 RPM
	case AssaultRifle:
		return 600 // 600 RPM
	case SniperRifle:
		return 60  // 60 RPM (slow, precision)
	case PumpShotgun:
		return 90  // 90 RPM
	case AutoShotgun:
		return 300 // 300 RPM
	default:
		return 300 // Default fire rate
	}
}

// GetCooldownMs returns the cooldown in milliseconds between shots
func (wt WeaponType) GetCooldownMs() int {
	rpm := wt.GetFireRate()
	// Convert RPM to milliseconds between shots: 60000ms/RPM
	return 60000 / rpm
}

// GetMaxRange returns the maximum range of the weapon
func (wt WeaponType) GetMaxRange() float64 {
	switch wt {
	case BasicPistol:
		return 50.0
	case AdvancedPistol:
		return 75.0
	case AssaultRifle:
		return 150.0
	case SniperRifle:
		return 300.0
	case PumpShotgun:
		return 30.0
	case AutoShotgun:
		return 40.0
	default:
		return 50.0 // Default range
	}
}

// Direction represents a 2D direction vector
type Direction struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewDirection creates a new normalized direction
func NewDirection(x, y float64) Direction {
	// Normalize the direction vector
	length := math.Sqrt(x*x + y*y)
	if length == 0 {
		return Direction{X: 0, Y: 0}
	}
	return Direction{
		X: x / length,
		Y: y / length,
	}
}

// IsZero checks if direction is zero vector
func (d Direction) IsZero() bool {
	return d.X == 0 && d.Y == 0
}

// Velocity represents bullet velocity with speed and direction
type Velocity struct {
	Speed     float64   `json:"speed"`      // Units per second
	Direction Direction `json:"direction"`
}

// NewVelocity creates a new velocity
func NewVelocity(speed float64, direction Direction) Velocity {
	return Velocity{
		Speed:     speed,
		Direction: direction,
	}
}

// GetVelocityComponents returns X and Y velocity components
func (v Velocity) GetVelocityComponents() (float64, float64) {
	return v.Speed * v.Direction.X, v.Speed * v.Direction.Y
}

// BulletState represents the current state of a bullet
type BulletState string

const (
	BulletStateActive  BulletState = "active"   // Bullet is flying
	BulletStateHit     BulletState = "hit"      // Bullet hit a target
	BulletStateExpired BulletState = "expired"  // Bullet expired (max range/time)
)

// String returns string representation
func (bs BulletState) String() string {
	return string(bs)
}

// IsValid checks if bullet state is valid
func (bs BulletState) IsValid() bool {
	return bs == BulletStateActive || bs == BulletStateHit || bs == BulletStateExpired
}

// CanTransitionTo checks if state transition is valid
func (bs BulletState) CanTransitionTo(newState BulletState) bool {
	switch bs {
	case BulletStateActive:
		// Active bullets can transition to hit or expired
		return newState == BulletStateHit || newState == BulletStateExpired
	case BulletStateHit, BulletStateExpired:
		// Terminal states - no further transitions
		return false
	default:
		return false
	}
}

// Bullet represents a bullet aggregate
type Bullet struct {
	ID           BulletID          `json:"id"`
	PlayerID     PlayerID          `json:"player_id"`
	WeaponType   WeaponType        `json:"weapon_type"`
	State        BulletState       `json:"state"`
	StartPos     shared.Position   `json:"start_position"`
	CurrentPos   shared.Position   `json:"current_position"`
	Velocity     Velocity          `json:"velocity"`
	MaxRange     float64           `json:"max_range"`
	FiredAt      time.Time         `json:"fired_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewBullet creates a new bullet
func NewBullet(playerID PlayerID, weaponType WeaponType, startPos shared.Position, direction Direction) (*Bullet, error) {
	if !weaponType.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Invalid weapon type")
	}

	if direction.IsZero() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Direction cannot be zero vector")
	}

	// Set bullet speed based on weapon type
	var speed float64
	switch weaponType {
	case BasicPistol, AdvancedPistol:
		speed = 200.0 // 200 units/second
	case AssaultRifle:
		speed = 300.0 // 300 units/second
	case SniperRifle:
		speed = 400.0 // 400 units/second (fastest)
	case PumpShotgun, AutoShotgun:
		speed = 150.0 // 150 units/second (slower)
	default:
		speed = 200.0 // Default speed
	}

	velocity := NewVelocity(speed, direction)
	maxRange := weaponType.GetMaxRange()
	now := time.Now()
	
	// Calculate expiry time based on max range and speed
	timeToExpiry := time.Duration(maxRange/speed) * time.Second
	expiresAt := now.Add(timeToExpiry)

	bullet := &Bullet{
		ID:         NewBulletID(),
		PlayerID:   playerID,
		WeaponType: weaponType,
		State:      BulletStateActive,
		StartPos:   startPos,
		CurrentPos: startPos,
		Velocity:   velocity,
		MaxRange:   maxRange,
		FiredAt:    now,
		ExpiresAt:  expiresAt,
		UpdatedAt:  now,
	}

	return bullet, nil
}

// UpdatePosition updates the bullet's current position based on elapsed time
func (b *Bullet) UpdatePosition(currentTime time.Time) error {
	if b.State != BulletStateActive {
		return shared.NewDomainError(shared.ErrCodeInvalidOperation, "Cannot update position of non-active bullet")
	}

	// Calculate elapsed time since fired
	elapsedSeconds := currentTime.Sub(b.FiredAt).Seconds()
	if elapsedSeconds < 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Current time cannot be before fired time")
	}

	// Calculate new position
	vx, vy := b.Velocity.GetVelocityComponents()
	newX := b.StartPos.X + (vx * elapsedSeconds)
	newY := b.StartPos.Y + (vy * elapsedSeconds)
	
	b.CurrentPos = shared.NewPosition(newX, newY)
	b.UpdatedAt = currentTime

	// Check if bullet has exceeded max range
	distanceTraveled := b.StartPos.DistanceTo(b.CurrentPos)
	if math.Sqrt(distanceTraveled) >= b.MaxRange {
		return b.Expire()
	}

	// Check if bullet has expired by time
	if currentTime.After(b.ExpiresAt) {
		return b.Expire()
	}

	return nil
}

// Hit marks the bullet as hit and sets final position
func (b *Bullet) Hit(hitPosition shared.Position) error {
	if !b.State.CanTransitionTo(BulletStateHit) {
		return shared.NewDomainError(shared.ErrCodeInvalidOperation, "Cannot mark bullet as hit from current state")
	}

	b.State = BulletStateHit
	b.CurrentPos = hitPosition
	b.UpdatedAt = time.Now()

	return nil
}

// Expire marks the bullet as expired
func (b *Bullet) Expire() error {
	if !b.State.CanTransitionTo(BulletStateExpired) {
		return shared.NewDomainError(shared.ErrCodeInvalidOperation, "Cannot expire bullet from current state")
	}

	b.State = BulletStateExpired
	b.UpdatedAt = time.Now()

	return nil
}

// IsActive checks if bullet is still active
func (b *Bullet) IsActive() bool {
	return b.State == BulletStateActive
}

// IsTerminal checks if bullet is in a terminal state
func (b *Bullet) IsTerminal() bool {
	return b.State == BulletStateHit || b.State == BulletStateExpired
}

// GetDistanceTraveled returns the distance traveled from start position
func (b *Bullet) GetDistanceTraveled() float64 {
	return math.Sqrt(b.StartPos.DistanceTo(b.CurrentPos))
}

// GetRemainingRange returns the remaining range before expiry
func (b *Bullet) GetRemainingRange() float64 {
	traveled := b.GetDistanceTraveled()
	remaining := b.MaxRange - traveled
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsExpiredByTime checks if bullet is expired by time
func (b *Bullet) IsExpiredByTime(currentTime time.Time) bool {
	return currentTime.After(b.ExpiresAt)
}

// IsExpiredByRange checks if bullet is expired by range
func (b *Bullet) IsExpiredByRange() bool {
	return b.GetDistanceTraveled() >= b.MaxRange
}

// ShouldExpire checks if bullet should be expired (by time or range)
func (b *Bullet) ShouldExpire(currentTime time.Time) bool {
	return b.IsExpiredByTime(currentTime) || b.IsExpiredByRange()
}

