package trainer

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// Event types
const (
	TrainerCreatedEventType         = "trainer.created"
	TrainerMovedEventType           = "trainer.moved"
	TrainerLeveledUpEventType       = "trainer.leveled_up"
	AnimalAddedToPartyEventType     = "trainer.animal_added_to_party"
	AnimalRemovedFromPartyEventType = "trainer.animal_removed_from_party"
)

// TrainerCreatedEvent represents trainer creation
type TrainerCreatedEvent struct {
	shared.BaseEvent
}

// TrainerCreatedEventData holds the event data
type TrainerCreatedEventData struct {
	TrainerID string `json:"trainer_id"`
	Nickname  string `json:"nickname"`
}

// NewTrainerCreatedEvent creates a new trainer created event
func NewTrainerCreatedEvent(trainerID, nickname string) (TrainerCreatedEvent, error) {
	data := TrainerCreatedEventData{
		TrainerID: trainerID,
		Nickname:  nickname,
	}

	baseEvent, err := shared.NewBaseEvent(
		TrainerCreatedEventType,
		trainerID,
		"trainer",
		data,
	)
	if err != nil {
		return TrainerCreatedEvent{}, err
	}

	return TrainerCreatedEvent{BaseEvent: baseEvent}, nil
}

// TrainerMovedEvent represents trainer movement
type TrainerMovedEvent struct {
	shared.BaseEvent
}

// TrainerMovedEventData holds the event data
type TrainerMovedEventData struct {
	TrainerID   string          `json:"trainer_id"`
	OldPosition shared.Position `json:"old_position"`
	NewPosition shared.Position `json:"new_position"`
}

// NewTrainerMovedEvent creates a new trainer moved event
func NewTrainerMovedEvent(trainerID string, oldPos, newPos shared.Position) (TrainerMovedEvent, error) {
	data := TrainerMovedEventData{
		TrainerID:   trainerID,
		OldPosition: oldPos,
		NewPosition: newPos,
	}

	baseEvent, err := shared.NewBaseEvent(
		TrainerMovedEventType,
		trainerID,
		"trainer",
		data,
	)
	if err != nil {
		return TrainerMovedEvent{}, err
	}

	return TrainerMovedEvent{BaseEvent: baseEvent}, nil
}

// TrainerLeveledUpEvent represents trainer leveling up
type TrainerLeveledUpEvent struct {
	shared.BaseEvent
}

// TrainerLeveledUpEventData holds the event data
type TrainerLeveledUpEventData struct {
	TrainerID string `json:"trainer_id"`
	NewLevel  int    `json:"new_level"`
}

// NewTrainerLeveledUpEvent creates a new trainer leveled up event
func NewTrainerLeveledUpEvent(trainerID string, newLevel int) (TrainerLeveledUpEvent, error) {
	data := TrainerLeveledUpEventData{
		TrainerID: trainerID,
		NewLevel:  newLevel,
	}

	baseEvent, err := shared.NewBaseEvent(
		TrainerLeveledUpEventType,
		trainerID,
		"trainer",
		data,
	)
	if err != nil {
		return TrainerLeveledUpEvent{}, err
	}

	return TrainerLeveledUpEvent{BaseEvent: baseEvent}, nil
}

// AnimalAddedToPartyEvent represents adding animal to party
type AnimalAddedToPartyEvent struct {
	shared.BaseEvent
}

// AnimalAddedToPartyEventData holds the event data
type AnimalAddedToPartyEventData struct {
	TrainerID string `json:"trainer_id"`
	AnimalID  string `json:"animal_id"`
}

// NewAnimalAddedToPartyEvent creates a new animal added to party event
func NewAnimalAddedToPartyEvent(trainerID, animalID string) (AnimalAddedToPartyEvent, error) {
	data := AnimalAddedToPartyEventData{
		TrainerID: trainerID,
		AnimalID:  animalID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalAddedToPartyEventType,
		trainerID,
		"trainer",
		data,
	)
	if err != nil {
		return AnimalAddedToPartyEvent{}, err
	}

	return AnimalAddedToPartyEvent{BaseEvent: baseEvent}, nil
}

// AnimalRemovedFromPartyEvent represents removing animal from party
type AnimalRemovedFromPartyEvent struct {
	shared.BaseEvent
}

// AnimalRemovedFromPartyEventData holds the event data
type AnimalRemovedFromPartyEventData struct {
	TrainerID string `json:"trainer_id"`
	AnimalID  string `json:"animal_id"`
}

// NewAnimalRemovedFromPartyEvent creates a new animal removed from party event
func NewAnimalRemovedFromPartyEvent(trainerID, animalID string) (AnimalRemovedFromPartyEvent, error) {
	data := AnimalRemovedFromPartyEventData{
		TrainerID: trainerID,
		AnimalID:  animalID,
	}

	baseEvent, err := shared.NewBaseEvent(
		AnimalRemovedFromPartyEventType,
		trainerID,
		"trainer",
		data,
	)
	if err != nil {
		return AnimalRemovedFromPartyEvent{}, err
	}

	return AnimalRemovedFromPartyEvent{BaseEvent: baseEvent}, nil
}
