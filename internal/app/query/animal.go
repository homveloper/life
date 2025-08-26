package query

import (
	"github.com/danghamo/life/internal/domain/animal"
	"github.com/danghamo/life/internal/domain/shared"
)

// Animal Queries

// GetAnimalByIDQuery gets an animal by ID
type GetAnimalByIDQuery struct {
	BaseQuery
	AnimalID string `json:"animal_id"`
}

// NewGetAnimalByIDQuery creates a new get animal by ID query
func NewGetAnimalByIDQuery(animalID string) GetAnimalByIDQuery {
	return GetAnimalByIDQuery{
		BaseQuery: NewBaseQuery("GetAnimalByID"),
		AnimalID:  animalID,
	}
}

// GetWildAnimalsByPositionQuery gets wild animals at a position
type GetWildAnimalsByPositionQuery struct {
	BaseQuery
	Position shared.Position `json:"position"`
}

// NewGetWildAnimalsByPositionQuery creates a new get wild animals by position query
func NewGetWildAnimalsByPositionQuery(position shared.Position) GetWildAnimalsByPositionQuery {
	return GetWildAnimalsByPositionQuery{
		BaseQuery: NewBaseQuery("GetWildAnimalsByPosition"),
		Position:  position,
	}
}

// GetAnimalsByOwnerQuery gets animals owned by a trainer
type GetAnimalsByOwnerQuery struct {
	BaseQuery
	OwnerID string `json:"owner_id"`
}

// NewGetAnimalsByOwnerQuery creates a new get animals by owner query
func NewGetAnimalsByOwnerQuery(ownerID string) GetAnimalsByOwnerQuery {
	return GetAnimalsByOwnerQuery{
		BaseQuery: NewBaseQuery("GetAnimalsByOwner"),
		OwnerID:   ownerID,
	}
}

// GetAnimalsByStateQuery gets animals by state
type GetAnimalsByStateQuery struct {
	BaseQuery
	State animal.AnimalState `json:"state"`
}

// NewGetAnimalsByStateQuery creates a new get animals by state query
func NewGetAnimalsByStateQuery(state animal.AnimalState) GetAnimalsByStateQuery {
	return GetAnimalsByStateQuery{
		BaseQuery: NewBaseQuery("GetAnimalsByState"),
		State:     state,
	}
}

// GetAnimalsByTypeQuery gets animals by type
type GetAnimalsByTypeQuery struct {
	BaseQuery
	AnimalType animal.AnimalType `json:"animal_type"`
}

// NewGetAnimalsByTypeQuery creates a new get animals by type query
func NewGetAnimalsByTypeQuery(animalType animal.AnimalType) GetAnimalsByTypeQuery {
	return GetAnimalsByTypeQuery{
		BaseQuery:  NewBaseQuery("GetAnimalsByType"),
		AnimalType: animalType,
	}
}

// GetAnimalsByLevelRangeQuery gets animals within level range
type GetAnimalsByLevelRangeQuery struct {
	BaseQuery
	MinLevel int `json:"min_level"`
	MaxLevel int `json:"max_level"`
}

// NewGetAnimalsByLevelRangeQuery creates a new get animals by level range query
func NewGetAnimalsByLevelRangeQuery(minLevel, maxLevel int) GetAnimalsByLevelRangeQuery {
	return GetAnimalsByLevelRangeQuery{
		BaseQuery: NewBaseQuery("GetAnimalsByLevelRange"),
		MinLevel:  minLevel,
		MaxLevel:  maxLevel,
	}
}

// GetAnimalStatsQuery gets an animal's stats
type GetAnimalStatsQuery struct {
	BaseQuery
	AnimalID string `json:"animal_id"`
}

// NewGetAnimalStatsQuery creates a new get animal stats query
func NewGetAnimalStatsQuery(animalID string) GetAnimalStatsQuery {
	return GetAnimalStatsQuery{
		BaseQuery: NewBaseQuery("GetAnimalStats"),
		AnimalID:  animalID,
	}
}

// ListAnimalsQuery lists animals with pagination
type ListAnimalsQuery struct {
	BaseQuery
	Pagination Pagination `json:"pagination"`
}

// NewListAnimalsQuery creates a new list animals query
func NewListAnimalsQuery(page, pageSize int) ListAnimalsQuery {
	return ListAnimalsQuery{
		BaseQuery:  NewBaseQuery("ListAnimals"),
		Pagination: NewPagination(page, pageSize),
	}
}

// GetNearbyAnimalsQuery gets animals near a position within radius
type GetNearbyAnimalsQuery struct {
	BaseQuery
	Position shared.Position `json:"position"`
	Radius   int             `json:"radius"`
}

// NewGetNearbyAnimalsQuery creates a new get nearby animals query
func NewGetNearbyAnimalsQuery(position shared.Position, radius int) GetNearbyAnimalsQuery {
	return GetNearbyAnimalsQuery{
		BaseQuery: NewBaseQuery("GetNearbyAnimals"),
		Position:  position,
		Radius:    radius,
	}
}
