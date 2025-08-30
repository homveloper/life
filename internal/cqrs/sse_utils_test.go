package cqrs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventPublisher for testing
type MockEventPublisher struct {
	mock.Mock
	PublishedEvents []interface{}
}

func (m *MockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	m.PublishedEvents = append(m.PublishedEvents, event)
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestSSEBroadcastHelper_BroadcastToAll(t *testing.T) {
	// Setup
	mockPublisher := &MockEventPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	
	helper := NewSSEBroadcastHelper(mockPublisher)
	ctx := context.Background()

	// Test broadcast to all
	method := "game.world.update"
	params := map[string]interface{}{
		"time": "12:00",
		"weather": "sunny",
	}

	err := helper.BroadcastToAll(ctx, method, params)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, mockPublisher.PublishedEvents, 1)
	
	event, ok := mockPublisher.PublishedEvents[0].(*SSENotificationEvent)
	assert.True(t, ok)
	assert.Equal(t, SSENotificationTypeBroadcast, event.Type)
	assert.Equal(t, method, event.Method)
	assert.Equal(t, params, event.Params)
	assert.Empty(t, event.TargetUsers) // Should be empty for broadcast
	assert.NotEmpty(t, event.RequestID)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
}

func TestSSEBroadcastHelper_BroadcastToUsers(t *testing.T) {
	// Setup
	mockPublisher := &MockEventPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	
	helper := NewSSEBroadcastHelper(mockPublisher)
	ctx := context.Background()

	// Test broadcast to specific users
	targetUsers := []string{"alice", "bob", "charlie"}
	method := "player.joined"
	params := map[string]interface{}{
		"player_id": "david",
		"nickname": "NewPlayer",
	}

	err := helper.BroadcastToUsers(ctx, targetUsers, method, params)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, mockPublisher.PublishedEvents, 1)
	
	event, ok := mockPublisher.PublishedEvents[0].(*SSENotificationEvent)
	assert.True(t, ok)
	assert.Equal(t, SSENotificationTypeUsers, event.Type)
	assert.Equal(t, method, event.Method)
	assert.Equal(t, params, event.Params)
	assert.Equal(t, targetUsers, event.TargetUsers)
	assert.NotEmpty(t, event.RequestID)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
}

func TestSSEBroadcastHelper_BroadcastToUsers_EmptyList(t *testing.T) {
	// Setup
	mockPublisher := &MockEventPublisher{}
	helper := NewSSEBroadcastHelper(mockPublisher)
	ctx := context.Background()

	// Test with empty user list - should return early without publishing
	err := helper.BroadcastToUsers(ctx, []string{}, "test.method", nil)
	
	assert.NoError(t, err)
	assert.Len(t, mockPublisher.PublishedEvents, 0) // No events should be published

	// Test with nil user list
	err = helper.BroadcastToUsers(ctx, nil, "test.method", nil)
	
	assert.NoError(t, err)
	assert.Len(t, mockPublisher.PublishedEvents, 0) // Still no events
}

func TestSSEBroadcastHelper_RealWorldUsageExample(t *testing.T) {
	// This test demonstrates how the SSE broadcast helper would be used in practice
	
	mockPublisher := &MockEventPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	
	helper := NewSSEBroadcastHelper(mockPublisher)
	ctx := context.Background()

	// Example 1: Player movement - broadcast to all players in the same area
	nearbyPlayers := []string{"alice", "bob", "charlie"}
	movementData := map[string]interface{}{
		"player_id": "alice",
		"x": 15.5,
		"y": 10.2,
		"direction": "north",
	}
	
	err := helper.BroadcastToUsers(ctx, nearbyPlayers, "player.movement", movementData)
	assert.NoError(t, err)

	// Example 2: Global announcement - broadcast to all connected players
	announcementData := map[string]interface{}{
		"message": "Server maintenance in 10 minutes",
		"type": "warning",
	}
	
	err = helper.BroadcastToAll(ctx, "system.announcement", announcementData)
	assert.NoError(t, err)

	// Example 3: Party invite - send to specific players
	partyMembers := []string{"alice", "bob"}
	inviteData := map[string]interface{}{
		"party_leader": "alice",
		"party_id": "party_123",
		"message": "You've been invited to join a party!",
	}
	
	err = helper.BroadcastToUsers(ctx, partyMembers, "party.invite", inviteData)
	assert.NoError(t, err)

	// Verify all events were published
	assert.Len(t, mockPublisher.PublishedEvents, 3)
	
	// Verify event types
	event1 := mockPublisher.PublishedEvents[0].(*SSENotificationEvent)
	event2 := mockPublisher.PublishedEvents[1].(*SSENotificationEvent)
	event3 := mockPublisher.PublishedEvents[2].(*SSENotificationEvent)
	
	assert.Equal(t, SSENotificationTypeUsers, event1.Type)
	assert.Equal(t, SSENotificationTypeBroadcast, event2.Type)
	assert.Equal(t, SSENotificationTypeUsers, event3.Type)
	
	// Verify target users
	assert.Equal(t, nearbyPlayers, event1.TargetUsers)
	assert.Empty(t, event2.TargetUsers) // Broadcast has empty target users
	assert.Equal(t, partyMembers, event3.TargetUsers)
}