package animal

import (
	"fmt"
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// AnimalID represents a unique animal identifier
type AnimalID shared.ID

// NewAnimalID creates a new animal ID
func NewAnimalID() AnimalID {
	return AnimalID(shared.NewID())
}

// String returns string representation
func (id AnimalID) String() string {
	return string(id)
}

// AnimalType represents the type of animal
type AnimalType string

const (
	Lion     AnimalType = "lion"
	Elephant AnimalType = "elephant"
	Cheetah  AnimalType = "cheetah"
)

// String returns string representation
func (at AnimalType) String() string {
	return string(at)
}

// IsValid checks if animal type is valid
func (at AnimalType) IsValid() bool {
	return at == Lion || at == Elephant || at == Cheetah
}

// GetBaseStats returns base stats for animal type
func (at AnimalType) GetBaseStats() shared.Stats {
	switch at {
	case Lion:
		return shared.NewStats(120, 25, 15, 12, 18) // High ATK, fast AS
	case Elephant:
		return shared.NewStats(200, 18, 25, 8, 10) // High HP, DEF, slow
	case Cheetah:
		return shared.NewStats(90, 20, 10, 25, 15) // High SPD, medium ATK
	default:
		return shared.NewStats(100, 15, 10, 10, 10) // Default stats
	}
}

// AnimalState represents the current state of an animal
type AnimalState string

const (
	Wild      AnimalState = "wild"       // Wild animal in the field
	Captured  AnimalState = "captured"   // Captured by trainer
	InParty   AnimalState = "in_party"   // Active in trainer's party
	InStorage AnimalState = "in_storage" // Stored (not in active party)
)

// String returns string representation
func (as AnimalState) String() string {
	return string(as)
}

// EquipmentSlot represents an equipment slot for animals
type EquipmentSlot struct {
	EquipmentID shared.ID `json:"equipment_id"`
	Equipped    bool      `json:"equipped"`
}

// NewEquipmentSlot creates a new equipment slot
func NewEquipmentSlot() EquipmentSlot {
	return EquipmentSlot{Equipped: false}
}

// Equip equips an equipment
func (es *EquipmentSlot) Equip(equipmentID shared.ID) {
	es.EquipmentID = equipmentID
	es.Equipped = true
}

// Unequip unequips the equipment
func (es *EquipmentSlot) Unequip() shared.ID {
	oldEquipmentID := es.EquipmentID
	es.EquipmentID = shared.ID("")
	es.Equipped = false
	return oldEquipmentID
}

// IsEquipped checks if slot has equipment
func (es *EquipmentSlot) IsEquipped() bool {
	return es.Equipped
}

// GetEquipmentID returns equipped equipment ID
func (es *EquipmentSlot) GetEquipmentID() shared.ID {
	return es.EquipmentID
}

// Animal represents an animal aggregate
type Animal struct {
	ID           AnimalID          `json:"id"`
	AnimalType   AnimalType        `json:"animal_type"`
	Level        shared.Level      `json:"level"`
	Experience   shared.Experience `json:"experience"`
	BaseStats    shared.Stats      `json:"base_stats"`
	CurrentStats shared.Stats      `json:"current_stats"`
	CurrentHP    int               `json:"current_hp"`
	MaxHP        int               `json:"max_hp"`
	State        AnimalState       `json:"state"`
	OwnerID      shared.ID         `json:"owner_id"` // TrainerID when captured
	Position     shared.Position   `json:"position"`
	Equipment    EquipmentSlot     `json:"equipment"` // Single necklace slot
	LastActionAt shared.Timestamp  `json:"last_action_at"`
	CreatedAt    shared.Timestamp  `json:"created_at"`
	UpdatedAt    shared.Timestamp  `json:"updated_at"`
}

// NewWildAnimal creates a new wild animal
func NewWildAnimal(animalType AnimalType, level int, position shared.Position) (*Animal, error) {
	if !animalType.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidAnimalType, fmt.Sprintf("Invalid animal type: %s", animalType))
	}

	animalLevel, err := shared.NewLevel(level)
	if err != nil {
		return nil, err
	}

	id := NewAnimalID()
	baseStats := animalType.GetBaseStats()

	// Scale stats by level
	levelMultiplier := float64(level)*0.1 + 1.0 // Level 1 = 1.1x, Level 10 = 2.0x
	scaledStats := shared.NewStats(
		int(float64(baseStats.HP)*levelMultiplier),
		int(float64(baseStats.ATK)*levelMultiplier),
		int(float64(baseStats.DEF)*levelMultiplier),
		int(float64(baseStats.SPD)*levelMultiplier),
		int(float64(baseStats.AS)*levelMultiplier),
	)

	experience, _ := shared.NewExperience(0, 0)
	timestamp := shared.NewTimestamp()

	animal := &Animal{
		ID:           id,
		AnimalType:   animalType,
		Level:        animalLevel,
		Experience:   experience,
		BaseStats:    baseStats,
		CurrentStats: scaledStats,
		CurrentHP:    scaledStats.HP,
		MaxHP:        scaledStats.HP,
		State:        Wild,
		Position:     position,
		Equipment:    NewEquipmentSlot(),
		LastActionAt: timestamp,
		CreatedAt:    timestamp,
		UpdatedAt:    timestamp,
	}

	return animal, nil
}

// NewCapturedAnimal creates a captured animal (from wild)
func NewCapturedAnimal(wild *Animal, ownerID shared.ID) (*Animal, error) {
	if wild.State != Wild {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidState, "Can only capture wild animals")
	}

	captured := *wild // Copy the wild animal
	captured.State = Captured
	captured.OwnerID = ownerID
	captured.UpdatedAt = shared.NewTimestamp()

	return &captured, nil
}

// IsWild checks if animal is wild
func (a *Animal) IsWild() bool {
	return a.State == Wild
}

// IsCaptured checks if animal is captured
func (a *Animal) IsCaptured() bool {
	return a.State == Captured || a.State == InParty || a.State == InStorage
}

// IsAlive checks if animal is alive (HP > 0)
func (a *Animal) IsAlive() bool {
	return a.CurrentHP > 0
}

// IsFainted checks if animal is fainted (HP = 0)
func (a *Animal) IsFainted() bool {
	return a.CurrentHP == 0
}

// MoveTo moves the animal to a new position
func (a *Animal) MoveTo(newPosition shared.Position) error {
	if !a.IsWild() {
		return shared.NewDomainError(shared.ErrCodeInvalidOperation, "Only wild animals can move freely")
	}

	a.Position = newPosition
	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// TakeDamage reduces HP by damage amount
func (a *Animal) TakeDamage(damage int) error {
	if damage < 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidDamage, "Damage cannot be negative")
	}

	if a.IsFainted() {
		return shared.NewDomainError(shared.ErrCodeAlreadyFainted, "Animal is already fainted")
	}

	// Apply defense
	actualDamage := damage - a.CurrentStats.DEF/2
	if actualDamage < 1 {
		actualDamage = 1 // Minimum 1 damage
	}

	a.CurrentHP -= actualDamage
	if a.CurrentHP < 0 {
		a.CurrentHP = 0
	}

	a.LastActionAt = shared.NewTimestamp()
	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// Heal restores HP
func (a *Animal) Heal(amount int) error {
	if amount < 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidHeal, "Heal amount cannot be negative")
	}

	a.CurrentHP += amount
	if a.CurrentHP > a.MaxHP {
		a.CurrentHP = a.MaxHP
	}

	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// GainExperience adds experience points
func (a *Animal) GainExperience(points int) error {
	if points <= 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidExp, "Experience points must be positive")
	}

	a.Experience = a.Experience.Add(points)
	a.UpdatedAt = shared.NewTimestamp()

	// Check for level up
	requiredExp := a.calculateRequiredExperience()
	if a.Experience.CanLevelUp(requiredExp) {
		return a.levelUp(requiredExp)
	}

	return nil
}

// levelUp levels up the animal
func (a *Animal) levelUp(requiredExp int) error {
	if !a.Level.CanLevelUp() {
		return shared.NewDomainError(shared.ErrCodeMaxLevel, "Animal is already at max level")
	}

	// Consume experience
	newExp, err := a.Experience.ConsumeForLevelUp(requiredExp)
	if err != nil {
		return err
	}

	// Level up
	newLevel, err := a.Level.LevelUp()
	if err != nil {
		return err
	}

	// Update stats based on animal type
	statGrowth := a.calculateStatGrowth()

	a.Experience = newExp
	a.Level = newLevel
	a.CurrentStats = a.CurrentStats.Add(statGrowth)
	a.MaxHP = a.CurrentStats.HP

	// Heal to full when leveling up
	a.CurrentHP = a.MaxHP

	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// calculateRequiredExperience calculates experience needed for next level
func (a *Animal) calculateRequiredExperience() int {
	// Simple exponential formula: level * level * 80
	return a.Level.Value() * a.Level.Value() * 80
}

// calculateStatGrowth calculates stat growth on level up
func (a *Animal) calculateStatGrowth() shared.Stats {
	switch a.AnimalType {
	case Lion:
		return shared.NewStats(15, 4, 2, 1, 3) // ATK and AS focused
	case Elephant:
		return shared.NewStats(25, 2, 4, 1, 1) // HP and DEF focused
	case Cheetah:
		return shared.NewStats(10, 3, 1, 4, 2) // SPD focused
	default:
		return shared.NewStats(12, 2, 2, 2, 2) // Balanced growth
	}
}

// EquipItem equips an item to the necklace slot
func (a *Animal) EquipItem(itemID shared.ID) error {
	if !a.IsCaptured() {
		return shared.NewDomainError(shared.ErrCodeNotCaptured, "Only captured animals can equip items")
	}

	if a.Equipment.IsEquipped() {
		a.Equipment.Unequip()
	}

	a.Equipment.Equip(itemID)
	a.UpdatedAt = shared.NewTimestamp()

	// TODO: Apply equipment stat bonuses when equipment domain is implemented

	return nil
}

// UnequipItem unequips the current item
func (a *Animal) UnequipItem() (shared.ID, error) {
	if !a.Equipment.IsEquipped() {
		return shared.ID(""), shared.NewDomainError(shared.ErrCodeNoEquipment, "No item equipped")
	}

	itemID := a.Equipment.Unequip()
	a.UpdatedAt = shared.NewTimestamp()

	// TODO: Remove equipment stat bonuses when equipment domain is implemented

	return itemID, nil
}

// ChangeState changes the animal state
func (a *Animal) ChangeState(newState AnimalState) error {
	if a.State == newState {
		return nil // No change
	}

	// Validate state transitions
	if !a.isValidStateTransition(newState) {
		return shared.NewDomainError(shared.ErrCodeInvalidStateTransition,
			fmt.Sprintf("Cannot transition from %s to %s", a.State, newState))
	}

	a.State = newState
	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// isValidStateTransition checks if state transition is valid
func (a *Animal) isValidStateTransition(newState AnimalState) bool {
	switch a.State {
	case Wild:
		return newState == Captured
	case Captured:
		return newState == InParty || newState == InStorage
	case InParty:
		return newState == InStorage
	case InStorage:
		return newState == InParty
	default:
		return false
	}
}

// CanBeCaptured checks if animal can be captured
func (a *Animal) CanBeCaptured() bool {
	return a.IsWild() && a.IsAlive()
}

// GetCaptureChance calculates capture chance based on current HP
func (a *Animal) GetCaptureChance(captureToolEffectiveness float64) float64 {
	if !a.CanBeCaptured() {
		return 0.0
	}

	// Base formula: (1 - currentHP/maxHP) * tool_effectiveness * random_factor
	hpRatio := 1.0 - (float64(a.CurrentHP) / float64(a.MaxHP))
	baseChance := hpRatio * captureToolEffectiveness

	// Cap at 95% max chance
	if baseChance > 0.95 {
		baseChance = 0.95
	}

	return baseChance
}

// CanPerformAction checks if animal can perform actions (not fainted, recently acted)
func (a *Animal) CanPerformAction(cooldownDuration time.Duration) bool {
	if a.IsFainted() {
		return false
	}

	return a.LastActionAt.IsExpired(cooldownDuration)
}

// GetEquippedItemID returns equipped item ID
func (a *Animal) GetEquippedItemID() (shared.ID, bool) {
	if a.Equipment.IsEquipped() {
		return a.Equipment.GetEquipmentID(), true
	}
	return shared.ID(""), false
}
