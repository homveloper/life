package animal

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// Event types
const (
	AnimalSpawnedEventType        = "animal.spawned"
	AnimalCapturedEventType       = "animal.captured"
	AnimalMovedEventType          = "animal.moved"
	AnimalTookDamageEventType     = "animal.took_damage"
	AnimalHealedEventType         = "animal.healed"
	AnimalFaintedEventType        = "animal.fainted"
	AnimalLeveledUpEventType      = "animal.leveled_up"
	AnimalEquippedItemEventType   = "animal.equipped_item"
	AnimalUnequippedItemEventType = "animal.unequipped_item"
	AnimalStateChangedEventType   = "animal.state_changed"
)

// AnimalSpawnedEvent represents animal spawning in the wild
type AnimalSpawnedEvent struct {
	shared.BaseEvent
}

// AnimalSpawnedEventData holds the event data
type AnimalSpawnedEventData struct {
	AnimalID   string          `json:"animal_id"`
	AnimalType string          `json:"animal_type"`
	Level      int             `json:"level"`
	Position   shared.Position `json:"position"`
}

// NewAnimalSpawnedEvent creates a new animal spawned event
func NewAnimalSpawnedEvent(animalID, animalType string, level int, position shared.Position) (AnimalSpawnedEvent, error) {
	data := AnimalSpawnedEventData{
		AnimalID:   animalID,
		AnimalType: animalType,
		Level:      level,
		Position:   position,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalSpawnedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalSpawnedEvent{}, err
	}

	return AnimalSpawnedEvent{BaseEvent: baseEvent}, nil
}

// AnimalCapturedEvent represents animal capture
type AnimalCapturedEvent struct {
	shared.BaseEvent
}

// AnimalCapturedEventData holds the event data
type AnimalCapturedEventData struct {
	AnimalID  string `json:"animal_id"`
	TrainerID string `json:"trainer_id"`
}

// NewAnimalCapturedEvent creates a new animal captured event
func NewAnimalCapturedEvent(animalID, trainerID string) (AnimalCapturedEvent, error) {
	data := AnimalCapturedEventData{
		AnimalID:  animalID,
		TrainerID: trainerID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalCapturedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalCapturedEvent{}, err
	}

	return AnimalCapturedEvent{BaseEvent: baseEvent}, nil
}

// AnimalMovedEvent represents animal movement
type AnimalMovedEvent struct {
	shared.BaseEvent
}

// AnimalMovedEventData holds the event data
type AnimalMovedEventData struct {
	AnimalID    string          `json:"animal_id"`
	OldPosition shared.Position `json:"old_position"`
	NewPosition shared.Position `json:"new_position"`
}

// NewAnimalMovedEvent creates a new animal moved event
func NewAnimalMovedEvent(animalID string, oldPos, newPos shared.Position) (AnimalMovedEvent, error) {
	data := AnimalMovedEventData{
		AnimalID:    animalID,
		OldPosition: oldPos,
		NewPosition: newPos,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalMovedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalMovedEvent{}, err
	}

	return AnimalMovedEvent{BaseEvent: baseEvent}, nil
}

// AnimalTookDamageEvent represents animal taking damage
type AnimalTookDamageEvent struct {
	shared.BaseEvent
}

// AnimalTookDamageEventData holds the event data
type AnimalTookDamageEventData struct {
	AnimalID  string `json:"animal_id"`
	Damage    int    `json:"damage"`
	CurrentHP int    `json:"current_hp"`
}

// NewAnimalTookDamageEvent creates a new animal took damage event
func NewAnimalTookDamageEvent(animalID string, damage, currentHP int) (AnimalTookDamageEvent, error) {
	data := AnimalTookDamageEventData{
		AnimalID:  animalID,
		Damage:    damage,
		CurrentHP: currentHP,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalTookDamageEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalTookDamageEvent{}, err
	}

	return AnimalTookDamageEvent{BaseEvent: baseEvent}, nil
}

// AnimalHealedEvent represents animal healing
type AnimalHealedEvent struct {
	shared.BaseEvent
}

// AnimalHealedEventData holds the event data
type AnimalHealedEventData struct {
	AnimalID   string `json:"animal_id"`
	HealAmount int    `json:"heal_amount"`
	CurrentHP  int    `json:"current_hp"`
}

// NewAnimalHealedEvent creates a new animal healed event
func NewAnimalHealedEvent(animalID string, healAmount, currentHP int) (AnimalHealedEvent, error) {
	data := AnimalHealedEventData{
		AnimalID:   animalID,
		HealAmount: healAmount,
		CurrentHP:  currentHP,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalHealedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalHealedEvent{}, err
	}

	return AnimalHealedEvent{BaseEvent: baseEvent}, nil
}

// AnimalFaintedEvent represents animal fainting
type AnimalFaintedEvent struct {
	shared.BaseEvent
}

// AnimalFaintedEventData holds the event data
type AnimalFaintedEventData struct {
	AnimalID string `json:"animal_id"`
}

// NewAnimalFaintedEvent creates a new animal fainted event
func NewAnimalFaintedEvent(animalID string) (AnimalFaintedEvent, error) {
	data := AnimalFaintedEventData{
		AnimalID: animalID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalFaintedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalFaintedEvent{}, err
	}

	return AnimalFaintedEvent{BaseEvent: baseEvent}, nil
}

// AnimalLeveledUpEvent represents animal leveling up
type AnimalLeveledUpEvent struct {
	shared.BaseEvent
}

// AnimalLeveledUpEventData holds the event data
type AnimalLeveledUpEventData struct {
	AnimalID string `json:"animal_id"`
	NewLevel int    `json:"new_level"`
}

// NewAnimalLeveledUpEvent creates a new animal leveled up event
func NewAnimalLeveledUpEvent(animalID string, newLevel int) (AnimalLeveledUpEvent, error) {
	data := AnimalLeveledUpEventData{
		AnimalID: animalID,
		NewLevel: newLevel,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalLeveledUpEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalLeveledUpEvent{}, err
	}

	return AnimalLeveledUpEvent{BaseEvent: baseEvent}, nil
}

// AnimalEquippedItemEvent represents animal equipping an item
type AnimalEquippedItemEvent struct {
	shared.BaseEvent
}

// AnimalEquippedItemEventData holds the event data
type AnimalEquippedItemEventData struct {
	AnimalID  string `json:"animal_id"`
	NewItemID string `json:"new_item_id"`
	OldItemID string `json:"old_item_id,omitempty"`
}

// NewAnimalEquippedItemEvent creates a new animal equipped item event
func NewAnimalEquippedItemEvent(animalID, newItemID, oldItemID string) (AnimalEquippedItemEvent, error) {
	data := AnimalEquippedItemEventData{
		AnimalID:  animalID,
		NewItemID: newItemID,
		OldItemID: oldItemID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalEquippedItemEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalEquippedItemEvent{}, err
	}

	return AnimalEquippedItemEvent{BaseEvent: baseEvent}, nil
}

// AnimalUnequippedItemEvent represents animal unequipping an item
type AnimalUnequippedItemEvent struct {
	shared.BaseEvent
}

// AnimalUnequippedItemEventData holds the event data
type AnimalUnequippedItemEventData struct {
	AnimalID string `json:"animal_id"`
	ItemID   string `json:"item_id"`
}

// NewAnimalUnequippedItemEvent creates a new animal unequipped item event
func NewAnimalUnequippedItemEvent(animalID, itemID string) (AnimalUnequippedItemEvent, error) {
	data := AnimalUnequippedItemEventData{
		AnimalID: animalID,
		ItemID:   itemID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalUnequippedItemEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalUnequippedItemEvent{}, err
	}

	return AnimalUnequippedItemEvent{BaseEvent: baseEvent}, nil
}

// AnimalStateChangedEvent represents animal state change
type AnimalStateChangedEvent struct {
	shared.BaseEvent
}

// AnimalStateChangedEventData holds the event data
type AnimalStateChangedEventData struct {
	AnimalID string `json:"animal_id"`
	OldState string `json:"old_state"`
	NewState string `json:"new_state"`
}

// NewAnimalStateChangedEvent creates a new animal state changed event
func NewAnimalStateChangedEvent(animalID, oldState, newState string) (AnimalStateChangedEvent, error) {
	data := AnimalStateChangedEventData{
		AnimalID: animalID,
		OldState: oldState,
		NewState: newState,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalStateChangedEventType,
		animalID,
		"animal",
		data,
	)
	if err != nil {
		return AnimalStateChangedEvent{}, err
	}

	return AnimalStateChangedEvent{BaseEvent: baseEvent}, nil
}
