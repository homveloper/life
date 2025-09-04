package shared

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ID represents a unique identifier
type ID string

// NewID generates a new unique ID
func NewID() ID {
	return ID(uuid.New().String())
}

// String returns the string representation of ID
func (id ID) String() string {
	return string(id)
}

// IsEmpty checks if ID is empty
func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// Position represents a 2D coordinate
// DESIGN DECISION: Uses float64 for free-form movement (not grid-based)
// - 1 unit = abstract game unit, independent of pixels or tiles
// - Allows smooth movement like Position{X: 245.7, Y: 182.3}
// - Tiles are separate concept with position and size for collision/terrain
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewPosition creates a new position
func NewPosition(x, y float64) Position {
	return Position{X: x, Y: y}
}

// DistanceTo calculates the distance to another position
func (p Position) DistanceTo(other Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return dx*dx + dy*dy // Using squared distance for performance
}

// IsAdjacent checks if this position is adjacent to another (within 1 unit)
func (p Position) IsAdjacent(other Position) bool {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return (dx >= -1.0 && dx <= 1.0) && (dy >= -1.0 && dy <= 1.0) && !(dx == 0 && dy == 0)
}

// String returns string representation of position
func (p Position) String() string {
	return fmt.Sprintf("(%.1f,%.1f)", p.X, p.Y)
}

// Key returns a string key for map indexing
func (p Position) Key() string {
	return fmt.Sprintf("%.1f,%.1f", p.X, p.Y)
}

// Stats represents game statistics
type Stats struct {
	HP  int `json:"hp"`  // Health Points
	ATK int `json:"atk"` // Attack
	DEF int `json:"def"` // Defense
	SPD int `json:"spd"` // Speed (movement)
	AS  int `json:"as"`  // Attack Speed
}

// NewStats creates new stats
func NewStats(hp, atk, def, spd, as int) Stats {
	return Stats{
		HP:  hp,
		ATK: atk,
		DEF: def,
		SPD: spd,
		AS:  as,
	}
}

// Add adds another stats to this stats
func (s Stats) Add(other Stats) Stats {
	return Stats{
		HP:  s.HP + other.HP,
		ATK: s.ATK + other.ATK,
		DEF: s.DEF + other.DEF,
		SPD: s.SPD + other.SPD,
		AS:  s.AS + other.AS,
	}
}

// IsValid checks if stats are valid (all positive)
func (s Stats) IsValid() bool {
	return s.HP > 0 && s.ATK >= 0 && s.DEF >= 0 && s.SPD >= 0 && s.AS >= 0
}

// Level represents a level value
type Level struct {
	value int
}

// NewLevel creates a new level (1-100)
func NewLevel(value int) (Level, error) {
	if value < 1 || value > 100 {
		return Level{}, fmt.Errorf("level must be between 1 and 100, got %d", value)
	}
	return Level{value: value}, nil
}

// Value returns the level value
func (l Level) Value() int {
	return l.value
}

// CanLevelUp checks if can level up (not max level)
func (l Level) CanLevelUp() bool {
	return l.value < 100
}

// LevelUp increases level by 1
func (l Level) LevelUp() (Level, error) {
	if !l.CanLevelUp() {
		return l, fmt.Errorf("already at max level")
	}
	return Level{value: l.value + 1}, nil
}

// Experience represents experience points
type Experience struct {
	current int
	total   int
}

// NewExperience creates new experience
func NewExperience(current, total int) (Experience, error) {
	if current < 0 || total < 0 || current > total {
		return Experience{}, fmt.Errorf("invalid experience values: current=%d, total=%d", current, total)
	}
	return Experience{current: current, total: total}, nil
}

// Current returns current experience
func (e Experience) Current() int {
	return e.current
}

// Total returns total experience
func (e Experience) Total() int {
	return e.total
}

// Add adds experience points
func (e Experience) Add(points int) Experience {
	return Experience{
		current: e.current + points,
		total:   e.total + points,
	}
}

// CanLevelUp checks if has enough experience to level up
func (e Experience) CanLevelUp(requiredExp int) bool {
	return e.current >= requiredExp
}

// ConsumeForLevelUp consumes experience for level up
func (e Experience) ConsumeForLevelUp(requiredExp int) (Experience, error) {
	if !e.CanLevelUp(requiredExp) {
		return e, fmt.Errorf("not enough experience: have %d, need %d", e.current, requiredExp)
	}
	return Experience{
		current: e.current - requiredExp,
		total:   e.total,
	}, nil
}

// Money represents currency
type Money struct {
	amount int
}

// NewMoney creates new money
func NewMoney(amount int) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("money amount cannot be negative: %d", amount)
	}
	return Money{amount: amount}, nil
}

// Amount returns the money amount
func (m Money) Amount() int {
	return m.amount
}

// Add adds money
func (m Money) Add(amount int) (Money, error) {
	newAmount := m.amount + amount
	if newAmount < 0 {
		return Money{}, fmt.Errorf("insufficient funds: have %d, trying to subtract %d", m.amount, -amount)
	}
	return Money{amount: newAmount}, nil
}

// CanAfford checks if can afford the cost
func (m Money) CanAfford(cost int) bool {
	return m.amount >= cost
}

// Timestamp represents a point in time
type Timestamp struct {
	value time.Time
}

// NewTimestamp creates a new timestamp
func NewTimestamp() Timestamp {
	return Timestamp{value: time.Now()}
}

// NewTimestampFromTime creates timestamp from time.Time
func NewTimestampFromTime(t time.Time) Timestamp {
	return Timestamp{value: t}
}

// Value returns the time value
func (t Timestamp) Value() time.Time {
	return t.value
}

// IsAfter checks if this timestamp is after another
func (t Timestamp) IsAfter(other Timestamp) bool {
	return t.value.After(other.value)
}

// IsExpired checks if timestamp is expired (older than duration)
func (t Timestamp) IsExpired(duration time.Duration) bool {
	return time.Since(t.value) > duration
}

// DurationSince returns duration since this timestamp
func (t Timestamp) DurationSince() time.Duration {
	return time.Since(t.value)
}
