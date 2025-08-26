package equipment

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// Equipment event types
const (
	EquipmentCreatedEventType    = "equipment.created"
	EquipmentEquippedEventType   = "equipment.equipped"
	EquipmentUnequippedEventType = "equipment.unequipped"
)

// EquipmentCreatedEventData represents the data for equipment created event
type EquipmentCreatedEventData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	EquipmentType string `json:"equipment_type"`
	Rarity        string `json:"rarity"`
}

// EquipmentCreatedEvent represents equipment created event
type EquipmentCreatedEvent struct {
	shared.BaseEvent
}

// NewEquipmentCreatedEvent creates a new equipment created event
func NewEquipmentCreatedEvent(equipmentID, name, equipmentType, rarity string) (shared.Event, error) {
	data := EquipmentCreatedEventData{
		ID:            equipmentID,
		Name:          name,
		EquipmentType: equipmentType,
		Rarity:        rarity,
	}

	baseEvent, err := shared.NewBaseEvent(
		EquipmentCreatedEventType,
		equipmentID,
		"equipment",
		data,
	)
	if err != nil {
		return nil, err
	}

	return EquipmentCreatedEvent{BaseEvent: baseEvent}, nil
}

// EquipmentEquippedEventData represents the data for equipment equipped event
type EquipmentEquippedEventData struct {
	EquipmentID     string `json:"equipment_id"`
	NewOwnerID      string `json:"new_owner_id"`
	PreviousOwnerID string `json:"previous_owner_id"`
}

// EquipmentEquippedEvent represents equipment equipped event
type EquipmentEquippedEvent struct {
	shared.BaseEvent
}

// NewEquipmentEquippedEvent creates a new equipment equipped event
func NewEquipmentEquippedEvent(equipmentID, newOwnerID, previousOwnerID string) (shared.Event, error) {
	data := EquipmentEquippedEventData{
		EquipmentID:     equipmentID,
		NewOwnerID:      newOwnerID,
		PreviousOwnerID: previousOwnerID,
	}

	baseEvent, err := shared.NewBaseEvent(
		EquipmentEquippedEventType,
		equipmentID,
		"equipment",
		data,
	)
	if err != nil {
		return nil, err
	}

	return EquipmentEquippedEvent{BaseEvent: baseEvent}, nil
}

// EquipmentUnequippedEventData represents the data for equipment unequipped event
type EquipmentUnequippedEventData struct {
	EquipmentID     string `json:"equipment_id"`
	PreviousOwnerID string `json:"previous_owner_id"`
}

// EquipmentUnequippedEvent represents equipment unequipped event
type EquipmentUnequippedEvent struct {
	shared.BaseEvent
}

// NewEquipmentUnequippedEvent creates a new equipment unequipped event
func NewEquipmentUnequippedEvent(equipmentID, previousOwnerID string) (shared.Event, error) {
	data := EquipmentUnequippedEventData{
		EquipmentID:     equipmentID,
		PreviousOwnerID: previousOwnerID,
	}

	baseEvent, err := shared.NewBaseEvent(
		EquipmentUnequippedEventType,
		equipmentID,
		"equipment",
		data,
	)
	if err != nil {
		return nil, err
	}

	return EquipmentUnequippedEvent{BaseEvent: baseEvent}, nil
}
