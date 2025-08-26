package command

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// World Commands

// CreateWorldCommand creates a new world
type CreateWorldCommand struct {
	BaseCommand
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// NewCreateWorldCommand creates a new create world command
func NewCreateWorldCommand(name string, width, height int) CreateWorldCommand {
	return CreateWorldCommand{
		BaseCommand: NewBaseCommand("CreateWorld", ""),
		Name:        name,
		Width:       width,
		Height:      height,
	}
}

// MoveEntityInWorldCommand moves an entity in the world
type MoveEntityInWorldCommand struct {
	BaseCommand
	WorldID  string          `json:"world_id"`
	EntityID string          `json:"entity_id"`
	FromPos  shared.Position `json:"from_position"`
	ToPos    shared.Position `json:"to_position"`
}

// NewMoveEntityInWorldCommand creates a new move entity in world command
func NewMoveEntityInWorldCommand(worldID, entityID string, fromPos, toPos shared.Position) MoveEntityInWorldCommand {
	return MoveEntityInWorldCommand{
		BaseCommand: NewBaseCommand("MoveEntityInWorld", worldID),
		WorldID:     worldID,
		EntityID:    entityID,
		FromPos:     fromPos,
		ToPos:       toPos,
	}
}
