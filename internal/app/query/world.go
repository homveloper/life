package query

import (
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/world"
)

// World Queries

// GetWorldByIDQuery gets a world by ID
type GetWorldByIDQuery struct {
	BaseQuery
	WorldID string `json:"world_id"`
}

// NewGetWorldByIDQuery creates a new get world by ID query
func NewGetWorldByIDQuery(worldID string) GetWorldByIDQuery {
	return GetWorldByIDQuery{
		BaseQuery: NewBaseQuery("GetWorldByID"),
		WorldID:   worldID,
	}
}

// GetTileAtPositionQuery gets a tile at specific position
type GetTileAtPositionQuery struct {
	BaseQuery
	WorldID  string          `json:"world_id"`
	Position shared.Position `json:"position"`
}

// NewGetTileAtPositionQuery creates a new get tile at position query
func NewGetTileAtPositionQuery(worldID string, position shared.Position) GetTileAtPositionQuery {
	return GetTileAtPositionQuery{
		BaseQuery: NewBaseQuery("GetTileAtPosition"),
		WorldID:   worldID,
		Position:  position,
	}
}

// GetEntitiesAtPositionQuery gets entities at specific position
type GetEntitiesAtPositionQuery struct {
	BaseQuery
	WorldID  string          `json:"world_id"`
	Position shared.Position `json:"position"`
}

// NewGetEntitiesAtPositionQuery creates a new get entities at position query
func NewGetEntitiesAtPositionQuery(worldID string, position shared.Position) GetEntitiesAtPositionQuery {
	return GetEntitiesAtPositionQuery{
		BaseQuery: NewBaseQuery("GetEntitiesAtPosition"),
		WorldID:   worldID,
		Position:  position,
	}
}

// GetNeighboringTilesQuery gets neighboring walkable tiles
type GetNeighboringTilesQuery struct {
	BaseQuery
	WorldID  string          `json:"world_id"`
	Position shared.Position `json:"position"`
}

// NewGetNeighboringTilesQuery creates a new get neighboring tiles query
func NewGetNeighboringTilesQuery(worldID string, position shared.Position) GetNeighboringTilesQuery {
	return GetNeighboringTilesQuery{
		BaseQuery: NewBaseQuery("GetNeighboringTiles"),
		WorldID:   worldID,
		Position:  position,
	}
}

// GetTilesByTerrainQuery gets tiles by terrain type
type GetTilesByTerrainQuery struct {
	BaseQuery
	WorldID     string            `json:"world_id"`
	TerrainType world.TerrainType `json:"terrain_type"`
}

// NewGetTilesByTerrainQuery creates a new get tiles by terrain query
func NewGetTilesByTerrainQuery(worldID string, terrainType world.TerrainType) GetTilesByTerrainQuery {
	return GetTilesByTerrainQuery{
		BaseQuery:   NewBaseQuery("GetTilesByTerrain"),
		WorldID:     worldID,
		TerrainType: terrainType,
	}
}

// GetWorldAreaQuery gets a rectangular area of tiles
type GetWorldAreaQuery struct {
	BaseQuery
	WorldID     string          `json:"world_id"`
	TopLeft     shared.Position `json:"top_left"`
	BottomRight shared.Position `json:"bottom_right"`
}

// NewGetWorldAreaQuery creates a new get world area query
func NewGetWorldAreaQuery(worldID string, topLeft, bottomRight shared.Position) GetWorldAreaQuery {
	return GetWorldAreaQuery{
		BaseQuery:   NewBaseQuery("GetWorldArea"),
		WorldID:     worldID,
		TopLeft:     topLeft,
		BottomRight: bottomRight,
	}
}

// IsPositionWalkableQuery checks if a position is walkable
type IsPositionWalkableQuery struct {
	BaseQuery
	WorldID  string          `json:"world_id"`
	Position shared.Position `json:"position"`
}

// NewIsPositionWalkableQuery creates a new is position walkable query
func NewIsPositionWalkableQuery(worldID string, position shared.Position) IsPositionWalkableQuery {
	return IsPositionWalkableQuery{
		BaseQuery: NewBaseQuery("IsPositionWalkable"),
		WorldID:   worldID,
		Position:  position,
	}
}

// ListWorldsQuery lists worlds with pagination
type ListWorldsQuery struct {
	BaseQuery
	Pagination Pagination `json:"pagination"`
}

// NewListWorldsQuery creates a new list worlds query
func NewListWorldsQuery(page, pageSize int) ListWorldsQuery {
	return ListWorldsQuery{
		BaseQuery:  NewBaseQuery("ListWorlds"),
		Pagination: NewPagination(page, pageSize),
	}
}
