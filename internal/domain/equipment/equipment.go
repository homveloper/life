package equipment

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// EquipmentID represents a unique equipment identifier
type EquipmentID shared.ID

// NewEquipmentID creates a new equipment ID
func NewEquipmentID() EquipmentID {
	return EquipmentID(shared.NewID())
}

// String returns string representation
func (id EquipmentID) String() string {
	return string(id)
}

// EquipmentType represents the type of equipment
type EquipmentType string

const (
	Necklace EquipmentType = "necklace" // Only equipment type for MVP
)

// String returns string representation
func (et EquipmentType) String() string {
	return string(et)
}

// IsValid checks if equipment type is valid
func (et EquipmentType) IsValid() bool {
	return et == Necklace
}

// Rarity represents equipment rarity
type Rarity string

const (
	Common    Rarity = "common"
	Rare      Rarity = "rare"
	Epic      Rarity = "epic"
	Legendary Rarity = "legendary"
)

// String returns string representation
func (r Rarity) String() string {
	return string(r)
}

// IsValid checks if rarity is valid
func (r Rarity) IsValid() bool {
	return r == Common || r == Rare || r == Epic || r == Legendary
}

// GetStatMultiplier returns stat multiplier for rarity
func (r Rarity) GetStatMultiplier() float64 {
	switch r {
	case Common:
		return 1.0
	case Rare:
		return 1.5
	case Epic:
		return 2.0
	case Legendary:
		return 3.0
	default:
		return 1.0
	}
}

// Equipment represents an equipment aggregate
type Equipment struct {
	ID            EquipmentID      `json:"id"`
	Name          string           `json:"name"`
	EquipmentType EquipmentType    `json:"equipment_type"`
	Rarity        Rarity           `json:"rarity"`
	BaseStats     shared.Stats     `json:"base_stats"`
	OwnerID       shared.ID        `json:"owner_id"` // AnimalID when equipped, empty when not equipped
	CreatedAt     shared.Timestamp `json:"created_at"`
	UpdatedAt     shared.Timestamp `json:"updated_at"`
}

// NewEquipment creates new equipment
func NewEquipment(name string, equipmentType EquipmentType, rarity Rarity, baseStats shared.Stats) (*Equipment, error) {
	if !equipmentType.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidEquipmentType, "Invalid equipment type")
	}

	if !rarity.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidRarity, "Invalid rarity")
	}

	if len(name) < 1 || len(name) > 50 {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Equipment name must be between 1 and 50 characters")
	}

	id := NewEquipmentID()
	timestamp := shared.NewTimestamp()

	equipment := &Equipment{
		ID:            id,
		Name:          name,
		EquipmentType: equipmentType,
		Rarity:        rarity,
		BaseStats:     baseStats,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
	}

	return equipment, nil
}

// IsEquipped checks if equipment is currently equipped
func (e *Equipment) IsEquipped() bool {
	return e.OwnerID != ""
}

// GetEffectiveStats returns stats with rarity multiplier applied
func (e *Equipment) GetEffectiveStats() shared.Stats {
	multiplier := e.Rarity.GetStatMultiplier()

	return shared.NewStats(
		int(float64(e.BaseStats.HP)*multiplier),
		int(float64(e.BaseStats.ATK)*multiplier),
		int(float64(e.BaseStats.DEF)*multiplier),
		int(float64(e.BaseStats.SPD)*multiplier),
		int(float64(e.BaseStats.AS)*multiplier),
	)
}

// EquipTo equips this equipment to an animal
func (e *Equipment) EquipTo(animalID shared.ID) error {
	if e.IsEquipped() {
		return shared.NewDomainError(shared.ErrCodeAlreadyEquipped, "Equipment is already equipped to another animal")
	}

	if animalID == "" {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Animal ID cannot be empty")
	}

	e.OwnerID = animalID
	e.UpdatedAt = shared.NewTimestamp()

	return nil
}

// Unequip unequips this equipment
func (e *Equipment) Unequip() error {
	if !e.IsEquipped() {
		return shared.NewDomainError(shared.ErrCodeNotEquipped, "Equipment is not currently equipped")
	}

	e.OwnerID = ""
	e.UpdatedAt = shared.NewTimestamp()

	return nil
}
