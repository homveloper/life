package command

import (
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
)

// Trainer Commands

// CreateTrainerCommand creates a new trainer
type CreateTrainerCommand struct {
	BaseCommand
	Nickname string `json:"nickname"`
}

// NewCreateTrainerCommand creates a new create trainer command
func NewCreateTrainerCommand(nickname string) CreateTrainerCommand {
	return CreateTrainerCommand{
		BaseCommand: NewBaseCommand("CreateTrainer", ""),
		Nickname:    nickname,
	}
}

// MoveTrainerCommand moves a trainer
type MoveTrainerCommand struct {
	BaseCommand
	TrainerID   string          `json:"trainer_id"`
	NewPosition shared.Position `json:"new_position"`
}

// NewMoveTrainerCommand creates a new move trainer command
func NewMoveTrainerCommand(trainerID string, newPosition shared.Position) MoveTrainerCommand {
	return MoveTrainerCommand{
		BaseCommand: NewBaseCommand("MoveTrainer", trainerID),
		TrainerID:   trainerID,
		NewPosition: newPosition,
	}
}

// AddAnimalToPartyCommand adds an animal to trainer's party
type AddAnimalToPartyCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	AnimalID  string `json:"animal_id"`
}

// NewAddAnimalToPartyCommand creates a new add animal to party command
func NewAddAnimalToPartyCommand(trainerID, animalID string) AddAnimalToPartyCommand {
	return AddAnimalToPartyCommand{
		BaseCommand: NewBaseCommand("AddAnimalToParty", trainerID),
		TrainerID:   trainerID,
		AnimalID:    animalID,
	}
}

// RemoveAnimalFromPartyCommand removes an animal from trainer's party
type RemoveAnimalFromPartyCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	AnimalID  string `json:"animal_id"`
}

// NewRemoveAnimalFromPartyCommand creates a new remove animal from party command
func NewRemoveAnimalFromPartyCommand(trainerID, animalID string) RemoveAnimalFromPartyCommand {
	return RemoveAnimalFromPartyCommand{
		BaseCommand: NewBaseCommand("RemoveAnimalFromParty", trainerID),
		TrainerID:   trainerID,
		AnimalID:    animalID,
	}
}

// AddItemToInventoryCommand adds an item to trainer's inventory
type AddItemToInventoryCommand struct {
	BaseCommand
	TrainerID string           `json:"trainer_id"`
	ItemType  trainer.ItemType `json:"item_type"`
	ItemName  string           `json:"item_name"`
}

// NewAddItemToInventoryCommand creates a new add item to inventory command
func NewAddItemToInventoryCommand(trainerID string, itemType trainer.ItemType, itemName string) AddItemToInventoryCommand {
	return AddItemToInventoryCommand{
		BaseCommand: NewBaseCommand("AddItemToInventory", trainerID),
		TrainerID:   trainerID,
		ItemType:    itemType,
		ItemName:    itemName,
	}
}

// RemoveItemFromInventoryCommand removes an item from trainer's inventory
type RemoveItemFromInventoryCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	ItemID    string `json:"item_id"`
}

// NewRemoveItemFromInventoryCommand creates a new remove item from inventory command
func NewRemoveItemFromInventoryCommand(trainerID, itemID string) RemoveItemFromInventoryCommand {
	return RemoveItemFromInventoryCommand{
		BaseCommand: NewBaseCommand("RemoveItemFromInventory", trainerID),
		TrainerID:   trainerID,
		ItemID:      itemID,
	}
}

// GainExperienceCommand adds experience to trainer
type GainExperienceCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	Points    int    `json:"points"`
}

// NewGainExperienceCommand creates a new gain experience command
func NewGainExperienceCommand(trainerID string, points int) GainExperienceCommand {
	return GainExperienceCommand{
		BaseCommand: NewBaseCommand("GainExperience", trainerID),
		TrainerID:   trainerID,
		Points:      points,
	}
}

// SpendMoneyCommand spends trainer's money
type SpendMoneyCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	Amount    int    `json:"amount"`
}

// NewSpendMoneyCommand creates a new spend money command
func NewSpendMoneyCommand(trainerID string, amount int) SpendMoneyCommand {
	return SpendMoneyCommand{
		BaseCommand: NewBaseCommand("SpendMoney", trainerID),
		TrainerID:   trainerID,
		Amount:      amount,
	}
}

// EarnMoneyCommand earns money for trainer
type EarnMoneyCommand struct {
	BaseCommand
	TrainerID string `json:"trainer_id"`
	Amount    int    `json:"amount"`
}

// NewEarnMoneyCommand creates a new earn money command
func NewEarnMoneyCommand(trainerID string, amount int) EarnMoneyCommand {
	return EarnMoneyCommand{
		BaseCommand: NewBaseCommand("EarnMoney", trainerID),
		TrainerID:   trainerID,
		Amount:      amount,
	}
}
