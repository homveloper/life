package command

import (
	"github.com/danghamo/life/internal/domain/animal"
	"github.com/danghamo/life/internal/domain/shared"
)

// Animal Commands

// SpawnAnimalCommand spawns a new wild animal
type SpawnAnimalCommand struct {
	BaseCommand
	AnimalType animal.AnimalType `json:"animal_type"`
	Level      int               `json:"level"`
	Position   shared.Position   `json:"position"`
}

// NewSpawnAnimalCommand creates a new spawn animal command
func NewSpawnAnimalCommand(animalType animal.AnimalType, level int, position shared.Position) SpawnAnimalCommand {
	return SpawnAnimalCommand{
		BaseCommand: NewBaseCommand("SpawnAnimal", ""),
		AnimalType:  animalType,
		Level:       level,
		Position:    position,
	}
}

// CaptureAnimalCommand captures a wild animal
type CaptureAnimalCommand struct {
	BaseCommand
	AnimalID  string `json:"animal_id"`
	TrainerID string `json:"trainer_id"`
}

// NewCaptureAnimalCommand creates a new capture animal command
func NewCaptureAnimalCommand(animalID, trainerID string) CaptureAnimalCommand {
	return CaptureAnimalCommand{
		BaseCommand: NewBaseCommand("CaptureAnimal", animalID),
		AnimalID:    animalID,
		TrainerID:   trainerID,
	}
}

// MoveAnimalCommand moves a wild animal
type MoveAnimalCommand struct {
	BaseCommand
	AnimalID    string          `json:"animal_id"`
	NewPosition shared.Position `json:"new_position"`
}

// NewMoveAnimalCommand creates a new move animal command
func NewMoveAnimalCommand(animalID string, newPosition shared.Position) MoveAnimalCommand {
	return MoveAnimalCommand{
		BaseCommand: NewBaseCommand("MoveAnimal", animalID),
		AnimalID:    animalID,
		NewPosition: newPosition,
	}
}

// AnimalTakeDamageCommand makes animal take damage
type AnimalTakeDamageCommand struct {
	BaseCommand
	AnimalID string `json:"animal_id"`
	Damage   int    `json:"damage"`
}

// NewAnimalTakeDamageCommand creates a new animal take damage command
func NewAnimalTakeDamageCommand(animalID string, damage int) AnimalTakeDamageCommand {
	return AnimalTakeDamageCommand{
		BaseCommand: NewBaseCommand("AnimalTakeDamage", animalID),
		AnimalID:    animalID,
		Damage:      damage,
	}
}

// HealAnimalCommand heals an animal
type HealAnimalCommand struct {
	BaseCommand
	AnimalID string `json:"animal_id"`
	Amount   int    `json:"amount"`
}

// NewHealAnimalCommand creates a new heal animal command
func NewHealAnimalCommand(animalID string, amount int) HealAnimalCommand {
	return HealAnimalCommand{
		BaseCommand: NewBaseCommand("HealAnimal", animalID),
		AnimalID:    animalID,
		Amount:      amount,
	}
}

// AnimalGainExperienceCommand adds experience to animal
type AnimalGainExperienceCommand struct {
	BaseCommand
	AnimalID string `json:"animal_id"`
	Points   int    `json:"points"`
}

// NewAnimalGainExperienceCommand creates a new animal gain experience command
func NewAnimalGainExperienceCommand(animalID string, points int) AnimalGainExperienceCommand {
	return AnimalGainExperienceCommand{
		BaseCommand: NewBaseCommand("AnimalGainExperience", animalID),
		AnimalID:    animalID,
		Points:      points,
	}
}

// EquipItemOnAnimalCommand equips an item on animal
type EquipItemOnAnimalCommand struct {
	BaseCommand
	AnimalID string `json:"animal_id"`
	ItemID   string `json:"item_id"`
}

// NewEquipItemOnAnimalCommand creates a new equip item on animal command
func NewEquipItemOnAnimalCommand(animalID, itemID string) EquipItemOnAnimalCommand {
	return EquipItemOnAnimalCommand{
		BaseCommand: NewBaseCommand("EquipItemOnAnimal", animalID),
		AnimalID:    animalID,
		ItemID:      itemID,
	}
}

// UnequipItemFromAnimalCommand unequips an item from animal
type UnequipItemFromAnimalCommand struct {
	BaseCommand
	AnimalID string `json:"animal_id"`
}

// NewUnequipItemFromAnimalCommand creates a new unequip item from animal command
func NewUnequipItemFromAnimalCommand(animalID string) UnequipItemFromAnimalCommand {
	return UnequipItemFromAnimalCommand{
		BaseCommand: NewBaseCommand("UnequipItemFromAnimal", animalID),
		AnimalID:    animalID,
	}
}

// ChangeAnimalStateCommand changes animal state
type ChangeAnimalStateCommand struct {
	BaseCommand
	AnimalID string             `json:"animal_id"`
	NewState animal.AnimalState `json:"new_state"`
}

// NewChangeAnimalStateCommand creates a new change animal state command
func NewChangeAnimalStateCommand(animalID string, newState animal.AnimalState) ChangeAnimalStateCommand {
	return ChangeAnimalStateCommand{
		BaseCommand: NewBaseCommand("ChangeAnimalState", animalID),
		AnimalID:    animalID,
		NewState:    newState,
	}
}
