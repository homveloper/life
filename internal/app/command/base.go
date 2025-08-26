package command

import (
	"context"
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// Command represents a command in CQRS pattern
type Command interface {
	CommandID() string
	CommandType() string
	AggregateID() string
	CreatedAt() time.Time
}

// BaseCommand provides common command functionality
type BaseCommand struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	AggrID    string    `json:"aggregate_id"`
	Timestamp time.Time `json:"created_at"`
}

// CommandID returns the command ID
func (c BaseCommand) CommandID() string {
	return c.ID
}

// CommandType returns the command type
func (c BaseCommand) CommandType() string {
	return c.Type
}

// AggregateID returns the aggregate ID
func (c BaseCommand) AggregateID() string {
	return c.AggrID
}

// CreatedAt returns when the command was created
func (c BaseCommand) CreatedAt() time.Time {
	return c.Timestamp
}

// NewBaseCommand creates a new base command
func NewBaseCommand(commandType, aggregateID string) BaseCommand {
	return BaseCommand{
		ID:        shared.NewID().String(),
		Type:      commandType,
		AggrID:    aggregateID,
		Timestamp: time.Now(),
	}
}

// CommandHandler handles commands
type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

// CommandBus dispatches commands to handlers
type CommandBus interface {
	Send(ctx context.Context, cmd Command) error
	Register(commandType string, handler CommandHandler)
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   error       `json:"error,omitempty"`
}

// NewSuccessResult creates a successful command result
func NewSuccessResult(message string, data interface{}) CommandResult {
	return CommandResult{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResult creates an error command result
func NewErrorResult(message string, err error) CommandResult {
	return CommandResult{
		Success: false,
		Message: message,
		Error:   err,
	}
}
