package world

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// World event types
const (
	WorldCreatedEventType = "world.created"
	EntityMovedEventType  = "world.entity_moved"
)

// WorldCreatedEventData represents the data for world created event
type WorldCreatedEventData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// WorldCreatedEvent represents world created event
type WorldCreatedEvent struct {
	shared.BaseEvent
}

// NewWorldCreatedEvent creates a new world created event
func NewWorldCreatedEvent(worldID, name string, width, height int) (shared.Event, error) {
	data := WorldCreatedEventData{
		ID:     worldID,
		Name:   name,
		Width:  width,
		Height: height,
	}

	baseEvent, err := shared.NewBaseEvent(
		WorldCreatedEventType,
		worldID,
		"world",
		data,
	)
	if err != nil {
		return nil, err
	}

	return WorldCreatedEvent{BaseEvent: baseEvent}, nil
}

// EntityMovedEventData represents the data for entity moved event
type EntityMovedEventData struct {
	WorldID  string          `json:"world_id"`
	EntityID string          `json:"entity_id"`
	FromPos  shared.Position `json:"from_position"`
	ToPos    shared.Position `json:"to_position"`
}

// EntityMovedEvent represents entity moved event
type EntityMovedEvent struct {
	shared.BaseEvent
}

// NewEntityMovedEvent creates a new entity moved event
func NewEntityMovedEvent(worldID, entityID string, fromPos, toPos shared.Position) (shared.Event, error) {
	data := EntityMovedEventData{
		WorldID:  worldID,
		EntityID: entityID,
		FromPos:  fromPos,
		ToPos:    toPos,
	}

	baseEvent, err := shared.NewBaseEvent(
		EntityMovedEventType,
		worldID,
		"world",
		data,
	)
	if err != nil {
		return nil, err
	}

	return EntityMovedEvent{BaseEvent: baseEvent}, nil
}
