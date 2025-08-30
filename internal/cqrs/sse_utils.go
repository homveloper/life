package cqrs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SSEBroadcastHelper provides utility functions for sending SSE notifications
type SSEBroadcastHelper struct {
	eventPublisher EventPublisher
}

// NewSSEBroadcastHelper creates a new SSE broadcast helper
func NewSSEBroadcastHelper(eventPublisher EventPublisher) *SSEBroadcastHelper {
	return &SSEBroadcastHelper{
		eventPublisher: eventPublisher,
	}
}

// BroadcastToAll broadcasts a message to all connected users across all servers
func (h *SSEBroadcastHelper) BroadcastToAll(ctx context.Context, method string, params interface{}) error {
	event := &SSENotificationEvent{
		Type:      SSENotificationTypeBroadcast,
		Method:    method,
		Params:    params,
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	}

	return h.eventPublisher.Publish(ctx, event)
}

// BroadcastToUsers broadcasts a message to specific users across all servers
// Each server will check if the target users are connected locally and send accordingly
func (h *SSEBroadcastHelper) BroadcastToUsers(ctx context.Context, userIDs []string, method string, params interface{}) error {
	if len(userIDs) == 0 {
		return nil
	}

	event := &SSENotificationEvent{
		Type:        SSENotificationTypeUsers,
		TargetUsers: userIDs,
		Method:      method,
		Params:      params,
		Timestamp:   time.Now(),
		RequestID:   uuid.New().String(),
	}

	return h.eventPublisher.Publish(ctx, event)
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}