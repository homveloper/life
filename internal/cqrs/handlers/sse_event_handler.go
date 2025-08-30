package handlers

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	cqrsevents "github.com/danghamo/life/internal/cqrs"
	"github.com/danghamo/life/pkg/logger"
)

// SSEBroadcaster interface for broadcasting SSE messages
type SSEBroadcaster interface {
	BroadcastToUsers(targetUsers []string, notification jsonrpcx.JsonRpcNotification)
	BroadcastToAll(notification jsonrpcx.JsonRpcNotification)
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

// SSEEventHandler handles events and converts them to SSE notifications
type SSEEventHandler struct {
	sseBroadcaster SSEBroadcaster
	eventPublisher EventPublisher
	logger         *logger.Logger
}

// NewSSEEventHandler creates a new SSE event handler
func NewSSEEventHandler(
	sseBroadcaster SSEBroadcaster,
	eventPublisher EventPublisher,
	logger *logger.Logger,
) *SSEEventHandler {
	return &SSEEventHandler{
		sseBroadcaster: sseBroadcaster,
		eventPublisher: eventPublisher,
		logger:         logger.WithComponent("sse-event-handler"),
	}
}

// HandleTrainerMovedEvent handles TrainerMovedEvent and broadcasts to SSE clients
func (h *SSEEventHandler) HandleTrainerMovedEvent(ctx context.Context, event *cqrsevents.TrainerMovedEvent) error {
	h.logger.Debug("Handling trainer moved event",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	// Create JSON-RPC notification for the trainer who moved (send changes only)
	userNotification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "trainer.position.updated",
		Params: map[string]interface{}{
			"user_id":    event.UserID,
			"changes":    event.Changes,
			"timestamp":  event.Timestamp.Format(time.RFC3339),
			"request_id": event.RequestID,
		},
	}

	// Send changes to the user who initiated the move
	h.sseBroadcaster.BroadcastToUsers([]string{event.UserID}, userNotification)

	// Create JSON-RPC notification for other users (send full position data)
	broadcastNotification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "trainer.position.broadcast",
		Params: map[string]interface{}{
			"user_id":   event.UserID,
			"position":  event.Position,
			"movement":  event.Movement,
			"timestamp": event.Timestamp.Format(time.RFC3339),
		},
	}

	// Broadcast position to all other users for real-time sync
	h.sseBroadcaster.BroadcastToAll(broadcastNotification)

	h.logger.Debug("Trainer moved event handled and broadcast",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	return nil
}

// HandleTrainerStoppedEvent handles TrainerStoppedEvent and broadcasts to SSE clients
func (h *SSEEventHandler) HandleTrainerStoppedEvent(ctx context.Context, event *cqrsevents.TrainerStoppedEvent) error {
	h.logger.Debug("Handling trainer stopped event",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	// Create JSON-RPC notification for the trainer who stopped (send changes only)
	userNotification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "trainer.movement.stopped",
		Params: map[string]interface{}{
			"user_id":    event.UserID,
			"changes":    event.Changes,
			"timestamp":  event.Timestamp.Format(time.RFC3339),
			"request_id": event.RequestID,
		},
	}

	// Send changes to the user who initiated the stop
	h.sseBroadcaster.BroadcastToUsers([]string{event.UserID}, userNotification)

	// Create JSON-RPC notification for other users (send full movement state)
	broadcastNotification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "trainer.movement.broadcast",
		Params: map[string]interface{}{
			"user_id":   event.UserID,
			"position":  event.Position,
			"movement":  event.Movement,
			"timestamp": event.Timestamp.Format(time.RFC3339),
		},
	}

	// Broadcast movement state to all other users
	h.sseBroadcaster.BroadcastToAll(broadcastNotification)

	h.logger.Debug("Trainer stopped event handled and broadcast",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	return nil
}

// HandleTrainerCreatedEvent handles TrainerCreatedEvent and broadcasts to SSE clients
func (h *SSEEventHandler) HandleTrainerCreatedEvent(ctx context.Context, event *cqrsevents.TrainerCreatedEvent) error {
	h.logger.Debug("Handling trainer created event",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	// Create JSON-RPC notification for trainer creation
	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "trainer.created",
		Params: map[string]interface{}{
			"user_id":   event.UserID,
			"nickname":  event.Trainer.Nickname,
			"position":  event.Trainer.Position,
			"level":     event.Trainer.Level,
			"timestamp": event.Timestamp.Format(time.RFC3339),
		},
	}

	// Broadcast trainer creation to all users
	h.sseBroadcaster.BroadcastToAll(notification)

	h.logger.Debug("Trainer created event handled and broadcast",
		zap.String("userId", event.UserID),
		zap.String("requestId", event.RequestID))

	return nil
}

// HandleSSENotificationEvent handles SSENotificationEvent for distributed SSE messaging
func (h *SSEEventHandler) HandleSSENotificationEvent(ctx context.Context, event *cqrsevents.SSENotificationEvent) error {
	h.logger.Debug("Handling SSE notification event",
		zap.String("type", event.Type),
		zap.Strings("targetUsers", event.TargetUsers),
		zap.String("method", event.Method),
		zap.String("requestId", event.RequestID))

	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  event.Method,
		Params:  event.Params,
	}

	switch event.Type {
	case cqrsevents.SSENotificationTypeUsers:
		// Array-based user targeting - only send if users are connected to this server
		if len(event.TargetUsers) > 0 {
			h.sseBroadcaster.BroadcastToUsers(event.TargetUsers, notification)
		}
	case cqrsevents.SSENotificationTypeBroadcast:
		// Broadcast to all connected users on this server
		h.sseBroadcaster.BroadcastToAll(notification)
	default:
		h.logger.Warn("Unknown SSE notification type", zap.String("type", event.Type))
	}

	h.logger.Debug("SSE notification event handled",
		zap.String("type", event.Type),
		zap.String("requestId", event.RequestID))

	return nil
}
