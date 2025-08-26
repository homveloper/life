package query

import (
	"github.com/danghamo/life/internal/domain/equipment"
)

// Equipment Queries

// GetEquipmentByIDQuery gets equipment by ID
type GetEquipmentByIDQuery struct {
	BaseQuery
	EquipmentID string `json:"equipment_id"`
}

// NewGetEquipmentByIDQuery creates a new get equipment by ID query
func NewGetEquipmentByIDQuery(equipmentID string) GetEquipmentByIDQuery {
	return GetEquipmentByIDQuery{
		BaseQuery:   NewBaseQuery("GetEquipmentByID"),
		EquipmentID: equipmentID,
	}
}

// GetEquipmentByOwnerQuery gets equipment owned by an animal
type GetEquipmentByOwnerQuery struct {
	BaseQuery
	OwnerID string `json:"owner_id"`
}

// NewGetEquipmentByOwnerQuery creates a new get equipment by owner query
func NewGetEquipmentByOwnerQuery(ownerID string) GetEquipmentByOwnerQuery {
	return GetEquipmentByOwnerQuery{
		BaseQuery: NewBaseQuery("GetEquipmentByOwner"),
		OwnerID:   ownerID,
	}
}

// GetEquipmentByTypeQuery gets equipment by type
type GetEquipmentByTypeQuery struct {
	BaseQuery
	EquipmentType equipment.EquipmentType `json:"equipment_type"`
}

// NewGetEquipmentByTypeQuery creates a new get equipment by type query
func NewGetEquipmentByTypeQuery(equipmentType equipment.EquipmentType) GetEquipmentByTypeQuery {
	return GetEquipmentByTypeQuery{
		BaseQuery:     NewBaseQuery("GetEquipmentByType"),
		EquipmentType: equipmentType,
	}
}

// GetEquipmentByRarityQuery gets equipment by rarity
type GetEquipmentByRarityQuery struct {
	BaseQuery
	Rarity equipment.Rarity `json:"rarity"`
}

// NewGetEquipmentByRarityQuery creates a new get equipment by rarity query
func NewGetEquipmentByRarityQuery(rarity equipment.Rarity) GetEquipmentByRarityQuery {
	return GetEquipmentByRarityQuery{
		BaseQuery: NewBaseQuery("GetEquipmentByRarity"),
		Rarity:    rarity,
	}
}

// GetUnequippedEquipmentQuery gets all unequipped equipment
type GetUnequippedEquipmentQuery struct {
	BaseQuery
}

// NewGetUnequippedEquipmentQuery creates a new get unequipped equipment query
func NewGetUnequippedEquipmentQuery() GetUnequippedEquipmentQuery {
	return GetUnequippedEquipmentQuery{
		BaseQuery: NewBaseQuery("GetUnequippedEquipment"),
	}
}

// ListEquipmentQuery lists equipment with pagination
type ListEquipmentQuery struct {
	BaseQuery
	Pagination Pagination `json:"pagination"`
}

// NewListEquipmentQuery creates a new list equipment query
func NewListEquipmentQuery(page, pageSize int) ListEquipmentQuery {
	return ListEquipmentQuery{
		BaseQuery:  NewBaseQuery("ListEquipment"),
		Pagination: NewPagination(page, pageSize),
	}
}
