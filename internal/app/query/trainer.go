package query

import (
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
)

// Trainer Queries

// GetTrainerByIDQuery gets a trainer by ID
type GetTrainerByIDQuery struct {
	BaseQuery
	TrainerID string `json:"trainer_id"`
}

// NewGetTrainerByIDQuery creates a new get trainer by ID query
func NewGetTrainerByIDQuery(trainerID string) GetTrainerByIDQuery {
	return GetTrainerByIDQuery{
		BaseQuery: NewBaseQuery("GetTrainerByID"),
		TrainerID: trainerID,
	}
}

// GetTrainerByNicknameQuery gets a trainer by nickname
type GetTrainerByNicknameQuery struct {
	BaseQuery
	Nickname string `json:"nickname"`
}

// NewGetTrainerByNicknameQuery creates a new get trainer by nickname query
func NewGetTrainerByNicknameQuery(nickname string) GetTrainerByNicknameQuery {
	return GetTrainerByNicknameQuery{
		BaseQuery: NewBaseQuery("GetTrainerByNickname"),
		Nickname:  nickname,
	}
}

// GetTrainersByPositionQuery gets trainers at a position
type GetTrainersByPositionQuery struct {
	BaseQuery
	Position shared.Position `json:"position"`
}

// NewGetTrainersByPositionQuery creates a new get trainers by position query
func NewGetTrainersByPositionQuery(position shared.Position) GetTrainersByPositionQuery {
	return GetTrainersByPositionQuery{
		BaseQuery: NewBaseQuery("GetTrainersByPosition"),
		Position:  position,
	}
}

// GetTrainerInventoryQuery gets a trainer's inventory
type GetTrainerInventoryQuery struct {
	BaseQuery
	TrainerID string `json:"trainer_id"`
}

// NewGetTrainerInventoryQuery creates a new get trainer inventory query
func NewGetTrainerInventoryQuery(trainerID string) GetTrainerInventoryQuery {
	return GetTrainerInventoryQuery{
		BaseQuery: NewBaseQuery("GetTrainerInventory"),
		TrainerID: trainerID,
	}
}

// GetTrainerInventoryByTypeQuery gets trainer's inventory items by type
type GetTrainerInventoryByTypeQuery struct {
	BaseQuery
	TrainerID string           `json:"trainer_id"`
	ItemType  trainer.ItemType `json:"item_type"`
}

// NewGetTrainerInventoryByTypeQuery creates a new get trainer inventory by type query
func NewGetTrainerInventoryByTypeQuery(trainerID string, itemType trainer.ItemType) GetTrainerInventoryByTypeQuery {
	return GetTrainerInventoryByTypeQuery{
		BaseQuery: NewBaseQuery("GetTrainerInventoryByType"),
		TrainerID: trainerID,
		ItemType:  itemType,
	}
}

// GetTrainerPartyQuery gets a trainer's animal party
type GetTrainerPartyQuery struct {
	BaseQuery
	TrainerID string `json:"trainer_id"`
}

// NewGetTrainerPartyQuery creates a new get trainer party query
func NewGetTrainerPartyQuery(trainerID string) GetTrainerPartyQuery {
	return GetTrainerPartyQuery{
		BaseQuery: NewBaseQuery("GetTrainerParty"),
		TrainerID: trainerID,
	}
}

// GetTrainerStatsQuery gets a trainer's stats
type GetTrainerStatsQuery struct {
	BaseQuery
	TrainerID string `json:"trainer_id"`
}

// NewGetTrainerStatsQuery creates a new get trainer stats query
func NewGetTrainerStatsQuery(trainerID string) GetTrainerStatsQuery {
	return GetTrainerStatsQuery{
		BaseQuery: NewBaseQuery("GetTrainerStats"),
		TrainerID: trainerID,
	}
}

// ListTrainersQuery lists trainers with pagination
type ListTrainersQuery struct {
	BaseQuery
	Pagination Pagination `json:"pagination"`
}

// NewListTrainersQuery creates a new list trainers query
func NewListTrainersQuery(page, pageSize int) ListTrainersQuery {
	return ListTrainersQuery{
		BaseQuery:  NewBaseQuery("ListTrainers"),
		Pagination: NewPagination(page, pageSize),
	}
}
