package command

import (
	"github.com/danghamo/life/internal/domain/equipment"
	"github.com/danghamo/life/internal/domain/shared"
)

// Equipment Commands

// CreateEquipmentCommand creates a new equipment
type CreateEquipmentCommand struct {
	BaseCommand
	Name          string                  `json:"name"`
	EquipmentType equipment.EquipmentType `json:"equipment_type"`
	Rarity        equipment.Rarity        `json:"rarity"`
	BaseStats     shared.Stats            `json:"base_stats"`
}

// NewCreateEquipmentCommand creates a new create equipment command
func NewCreateEquipmentCommand(name string, equipmentType equipment.EquipmentType, rarity equipment.Rarity, baseStats shared.Stats) CreateEquipmentCommand {
	return CreateEquipmentCommand{
		BaseCommand:   NewBaseCommand("CreateEquipment", ""),
		Name:          name,
		EquipmentType: equipmentType,
		Rarity:        rarity,
		BaseStats:     baseStats,
	}
}

// EquipToAnimalCommand equips equipment to an animal
type EquipToAnimalCommand struct {
	BaseCommand
	EquipmentID string `json:"equipment_id"`
	AnimalID    string `json:"animal_id"`
}

// NewEquipToAnimalCommand creates a new equip to animal command
func NewEquipToAnimalCommand(equipmentID, animalID string) EquipToAnimalCommand {
	return EquipToAnimalCommand{
		BaseCommand: NewBaseCommand("EquipToAnimal", equipmentID),
		EquipmentID: equipmentID,
		AnimalID:    animalID,
	}
}

// UnequipFromAnimalCommand unequips equipment from an animal
type UnequipFromAnimalCommand struct {
	BaseCommand
	EquipmentID string `json:"equipment_id"`
}

// NewUnequipFromAnimalCommand creates a new unequip from animal command
func NewUnequipFromAnimalCommand(equipmentID string) UnequipFromAnimalCommand {
	return UnequipFromAnimalCommand{
		BaseCommand: NewBaseCommand("UnequipFromAnimal", equipmentID),
		EquipmentID: equipmentID,
	}
}
