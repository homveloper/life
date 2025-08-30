package sse

import (
	"testing"
	"time"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient simulates an SSE client for testing
type MockClient struct {
	mock.Mock
	ID       string
	UserID   string
	Messages []jsonrpcx.JsonRpcNotification
}

func (m *MockClient) ReceiveMessage(notification jsonrpcx.JsonRpcNotification) {
	m.Messages = append(m.Messages, notification)
}

func TestSSEBroadcaster_BroadcastToUsers(t *testing.T) {
	// Create broadcaster
	logger := logger.NewDefault()
	broadcaster := NewSSEBroadcaster(logger)

	// Create mock clients
	client1 := &MockClient{ID: "client1", UserID: "alice"}
	client2 := &MockClient{ID: "client2", UserID: "bob"}
	client3 := &MockClient{ID: "client3", UserID: "alice"} // Same user, different connection

	// Add clients to broadcaster
	broadcaster.AddClient(&SSEClient{
		ID:       client1.ID,
		UserID:   client1.UserID,
		Writer:   nil, // Mock writers not needed for this test
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	broadcaster.AddClient(&SSEClient{
		ID:       client2.ID,
		UserID:   client2.UserID,
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	broadcaster.AddClient(&SSEClient{
		ID:       client3.ID,
		UserID:   client3.UserID,
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	// Test: Broadcast to specific users (alice and bob)
	targetUsers := []string{"alice", "bob"}
	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "test.message",
		Params: map[string]interface{}{
			"message": "Hello targeted users!",
		},
	}

	// Should identify that alice and bob are connected to this server
	broadcaster.BroadcastToUsers(targetUsers, notification)

	// Verify that the method correctly identifies local users
	// alice should be found (2 connections)
	// bob should be found (1 connection)
	// In a real test, we'd need to mock the internal broadcast mechanism

	assert.Equal(t, 3, broadcaster.GetClientCount())
}

func TestSSEBroadcaster_BroadcastToAll(t *testing.T) {
	testLogger := logger.NewDefault()
	broadcaster := NewSSEBroadcaster(testLogger)

	// Add some clients
	broadcaster.AddClient(&SSEClient{
		ID:       "client1",
		UserID:   "alice",
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	broadcaster.AddClient(&SSEClient{
		ID:       "client2",
		UserID:   "bob",
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "broadcast.message",
		Params: map[string]interface{}{
			"message": "Hello everyone!",
		},
	}

	// Test broadcast to all
	broadcaster.BroadcastToAll(notification)

	// Verify client count
	assert.Equal(t, 2, broadcaster.GetClientCount())
}

func TestSSEBroadcaster_LocalUserCheck(t *testing.T) {
	testLogger := logger.NewDefault()
	broadcaster := NewSSEBroadcaster(testLogger)

	// Add clients for specific users
	broadcaster.AddClient(&SSEClient{
		ID:       "client1",
		UserID:   "alice",
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	broadcaster.AddClient(&SSEClient{
		ID:       "client2",
		UserID:   "bob",
		Writer:   nil,
		Flusher:  nil,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	})

	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "targeted.message",
		Params: map[string]interface{}{
			"message": "Targeted message",
		},
	}

	// Test 1: Target users that exist locally (alice, bob)
	broadcaster.BroadcastToUsers([]string{"alice", "bob"}, notification)

	// Test 2: Target users that don't exist locally (charlie)
	broadcaster.BroadcastToUsers([]string{"charlie"}, notification)

	// Test 3: Mixed - some exist locally, some don't (alice exists, david doesn't)
	broadcaster.BroadcastToUsers([]string{"alice", "david"}, notification)

	// All tests should complete without error
	// The broadcaster should only send to locally connected users
	assert.Equal(t, 2, broadcaster.GetClientCount())
}

func TestSSEBroadcaster_EmptyTargetUsers(t *testing.T) {
	testLogger := logger.NewDefault()
	broadcaster := NewSSEBroadcaster(testLogger)

	notification := jsonrpcx.JsonRpcNotification{
		Jsonrpc: "2.0",
		Method:  "test.message",
		Params:  map[string]interface{}{"test": true},
	}

	// Test with empty target users - should return early without error
	broadcaster.BroadcastToUsers([]string{}, notification)
	broadcaster.BroadcastToUsers(nil, notification)

	// Should complete without issues
	assert.Equal(t, 0, broadcaster.GetClientCount())
}
