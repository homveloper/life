package shared

import (
	"encoding/json"
	"time"

	"github.com/samber/oops"
)

// Event represents a domain event
type Event interface {
	// EventID returns unique event identifier
	EventID() string
	// EventType returns the type of event
	EventType() string
	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() string
	// AggregateType returns the type of aggregate
	AggregateType() string
	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
	// Version returns the event version
	Version() int
	// Data returns the event data as JSON
	Data() ([]byte, error)
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	AggrID       string    `json:"aggregate_id"`
	AggrType     string    `json:"aggregate_type"`
	Timestamp    time.Time `json:"occurred_at"`
	EventVersion int       `json:"version"`
	EventData    []byte    `json:"data"`
}

// EventID returns unique event identifier
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the type of event
func (e BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the ID of the aggregate
func (e BaseEvent) AggregateID() string {
	return e.AggrID
}

// AggregateType returns the type of aggregate
func (e BaseEvent) AggregateType() string {
	return e.AggrType
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Version returns the event version
func (e BaseEvent) Version() int {
	return e.EventVersion
}

// Data returns the event data as JSON
func (e BaseEvent) Data() ([]byte, error) {
	return e.EventData, nil
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, aggregateID, aggregateType string, data interface{}) (BaseEvent, error) {
	eventData, err := json.Marshal(data)
	if err != nil {
		return BaseEvent{}, err
	}

	return BaseEvent{
		ID:           NewID().String(),
		Type:         eventType,
		AggrID:       aggregateID,
		AggrType:     aggregateType,
		Timestamp:    time.Now(),
		EventVersion: 1,
		EventData:    eventData,
	}, nil
}

// EventStore represents an event store interface
type EventStore interface {
	// SaveEvents saves events for an aggregate
	SaveEvents(aggregateID string, events []Event, expectedVersion int) error
	// LoadEvents loads all events for an aggregate
	LoadEvents(aggregateID string) ([]Event, error)
	// LoadEventsFromVersion loads events from a specific version
	LoadEventsFromVersion(aggregateID string, fromVersion int) ([]Event, error)
}

// EventBus represents an event bus interface
type EventBus interface {
	// Publish publishes an event
	Publish(event Event) error
	// Subscribe subscribes to events of a specific type
	Subscribe(eventType string, handler EventHandler) error
}

// EventHandler handles events
type EventHandler func(event Event) error

// AggregateRoot represents the base for all aggregates
type AggregateRoot struct {
	id                ID
	version           int
	uncommittedEvents []Event
}

// NewAggregateRoot creates a new aggregate root
func NewAggregateRoot(id ID) AggregateRoot {
	return AggregateRoot{
		id:                id,
		version:           0,
		uncommittedEvents: make([]Event, 0),
	}
}

// ID returns the aggregate ID
func (ar *AggregateRoot) ID() ID {
	return ar.id
}

// Version returns the current version
func (ar *AggregateRoot) Version() int {
	return ar.version
}

// UncommittedEvents returns uncommitted events
func (ar *AggregateRoot) UncommittedEvents() []Event {
	return ar.uncommittedEvents
}

// MarkEventsAsCommitted marks events as committed
func (ar *AggregateRoot) MarkEventsAsCommitted() {
	ar.uncommittedEvents = make([]Event, 0)
}

// ApplyEvent applies an event to the aggregate
func (ar *AggregateRoot) ApplyEvent(event Event) {
	ar.uncommittedEvents = append(ar.uncommittedEvents, event)
	ar.version++
}

// LoadFromHistory loads aggregate from event history
func (ar *AggregateRoot) LoadFromHistory(events []Event) {
	for range events {
		ar.version++
	}
}

// Repository represents the base repository interface
type Repository interface {
	// Save saves an aggregate
	Save(aggregate interface{}) error
	// Load loads an aggregate by ID
	Load(id ID) (interface{}, error)
	// Exists checks if aggregate exists
	Exists(id ID) (bool, error)
}

// Domain error codes
const (
	ErrCodeInvalidInput      = 1001
	ErrCodeNotFound          = 1002
	ErrCodeAlreadyExists     = 1003
	ErrCodeInvalidOperation  = 1004
	ErrCodeInsufficientFunds = 1005

	// Trainer specific errors (2000-2999)
	ErrCodeInvalidNickname      = 2001
	ErrCodeInventoryFull        = 2002
	ErrCodePartyFull            = 2003
	ErrCodeAnimalNotInParty     = 2004
	ErrCodeAnimalAlreadyInParty = 2005
	ErrCodeInsufficientItems    = 2006
	ErrCodeMaxLevel             = 2007
	ErrCodeInvalidExp           = 2008
	ErrCodeInvalidAmount        = 2009
	ErrCodeInvalidItemType      = 2010
	ErrCodeItemNotFound         = 2011

	// Animal specific errors (3000-3999)
	ErrCodeInvalidAnimalType      = 3001
	ErrCodeInvalidState           = 3002
	ErrCodeNotCaptured            = 3003
	ErrCodeAlreadyFainted         = 3004
	ErrCodeInvalidDamage          = 3005
	ErrCodeInvalidHeal            = 3006
	ErrCodeNoEquipment            = 3007
	ErrCodeInvalidStateTransition = 3008

	// Equipment specific errors (4000-4999)
	ErrCodeInvalidEquipmentType = 4001
	ErrCodeInvalidRarity        = 4002
	ErrCodeAlreadyEquipped      = 4003
	ErrCodeNotEquipped          = 4004

	// World specific errors (5000-5999)
	ErrCodeInvalidWorldSize    = 5001
	ErrCodeInvalidPosition     = 5002
	ErrCodeTileNotFound        = 5003
	ErrCodeEntityAlreadyOnTile = 5004
	ErrCodeEntityNotOnTile     = 5005
	ErrCodeInvalidMove         = 5006
)

// NewDomainError creates a new domain error using oops
func NewDomainError(code int, message string) error {
	return oops.
		Code(codeToString(code)).
		In("domain").
		With("error_code", code).
		Errorf(message)
}

// NewDomainErrorf creates a new domain error with formatted message
func NewDomainErrorf(code int, format string, args ...interface{}) error {
	return oops.
		Code(codeToString(code)).
		In("domain").
		With("error_code", code).
		Errorf(format, args...)
}

// WrapDomainError wraps an existing error with domain context
func WrapDomainError(err error, code int, message string) error {
	return oops.
		Code(codeToString(code)).
		In("domain").
		With("error_code", code).
		Wrapf(err, message)
}

// codeToString converts int error code to string
func codeToString(code int) string {
	switch code {
	case ErrCodeInvalidInput:
		return "INVALID_INPUT"
	case ErrCodeNotFound:
		return "NOT_FOUND"
	case ErrCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrCodeInvalidOperation:
		return "INVALID_OPERATION"
	case ErrCodeInsufficientFunds:
		return "INSUFFICIENT_FUNDS"
	case ErrCodeInvalidNickname:
		return "INVALID_NICKNAME"
	case ErrCodeInventoryFull:
		return "INVENTORY_FULL"
	case ErrCodePartyFull:
		return "PARTY_FULL"
	case ErrCodeAnimalNotInParty:
		return "ANIMAL_NOT_IN_PARTY"
	case ErrCodeAnimalAlreadyInParty:
		return "ANIMAL_ALREADY_IN_PARTY"
	case ErrCodeInsufficientItems:
		return "INSUFFICIENT_ITEMS"
	case ErrCodeMaxLevel:
		return "MAX_LEVEL"
	case ErrCodeInvalidExp:
		return "INVALID_EXP"
	case ErrCodeInvalidAmount:
		return "INVALID_AMOUNT"
	case ErrCodeInvalidItemType:
		return "INVALID_ITEM_TYPE"
	case ErrCodeItemNotFound:
		return "ITEM_NOT_FOUND"
	case ErrCodeInvalidAnimalType:
		return "INVALID_ANIMAL_TYPE"
	case ErrCodeInvalidState:
		return "INVALID_STATE"
	case ErrCodeNotCaptured:
		return "NOT_CAPTURED"
	case ErrCodeAlreadyFainted:
		return "ALREADY_FAINTED"
	case ErrCodeInvalidDamage:
		return "INVALID_DAMAGE"
	case ErrCodeInvalidHeal:
		return "INVALID_HEAL"
	case ErrCodeNoEquipment:
		return "NO_EQUIPMENT"
	case ErrCodeInvalidStateTransition:
		return "INVALID_STATE_TRANSITION"
	case ErrCodeInvalidEquipmentType:
		return "INVALID_EQUIPMENT_TYPE"
	case ErrCodeInvalidRarity:
		return "INVALID_RARITY"
	case ErrCodeAlreadyEquipped:
		return "ALREADY_EQUIPPED"
	case ErrCodeNotEquipped:
		return "NOT_EQUIPPED"
	case ErrCodeInvalidWorldSize:
		return "INVALID_WORLD_SIZE"
	case ErrCodeInvalidPosition:
		return "INVALID_POSITION"
	case ErrCodeTileNotFound:
		return "TILE_NOT_FOUND"
	case ErrCodeEntityAlreadyOnTile:
		return "ENTITY_ALREADY_ON_TILE"
	case ErrCodeEntityNotOnTile:
		return "ENTITY_NOT_ON_TILE"
	case ErrCodeInvalidMove:
		return "INVALID_MOVE"
	default:
		return "UNKNOWN_ERROR"
	}
}

// Common domain error builders
func ErrInvalidInput(msg string) error {
	return NewDomainError(ErrCodeInvalidInput, msg)
}

func ErrNotFound(resource string) error {
	return NewDomainErrorf(ErrCodeNotFound, "%s not found", resource)
}

func ErrAlreadyExists(resource string) error {
	return NewDomainErrorf(ErrCodeAlreadyExists, "%s already exists", resource)
}

func ErrInvalidOperation(operation string) error {
	return NewDomainErrorf(ErrCodeInvalidOperation, "Invalid operation: %s", operation)
}

func ErrInsufficientFunds() error {
	return NewDomainError(ErrCodeInsufficientFunds, "Insufficient funds")
}
