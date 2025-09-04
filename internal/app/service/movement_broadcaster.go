package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	cqrscommands "github.com/danghamo/life/internal/cqrs"
	"github.com/danghamo/life/internal/domain/trainer"
	"github.com/danghamo/life/pkg/logger"
)

// MovementBroadcaster handles periodic broadcasting of moving trainer positions using Redis
type MovementBroadcaster struct {
	logger          *logger.Logger
	repository      trainer.Repository
	eventBus        *cqrs.EventBus
	redisClient     *redis.Client
	stopChan        chan struct{}
	broadcastTicker *time.Ticker
}

const (
	// Redis key pattern for moving trainers: "moving:trainer:{userID}"
	movingTrainerKeyPrefix = "moving:trainer:"
	// TTL for moving trainer keys (30 seconds)
	movingTrainerTTL = 30 * time.Second
)

// NewMovementBroadcaster creates a new Redis-based movement broadcaster
func NewMovementBroadcaster(
	logger *logger.Logger,
	repository trainer.Repository,
	eventBus *cqrs.EventBus,
	redisClient *redis.Client,
) *MovementBroadcaster {
	return &MovementBroadcaster{
		logger:      logger.WithComponent("movement-broadcaster"),
		repository:  repository,
		eventBus:    eventBus,
		redisClient: redisClient,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the periodic broadcasting
func (mb *MovementBroadcaster) Start(ctx context.Context) {
	// Broadcast moving trainers at 60Hz for ultra smooth movement (16.67ms)
	mb.broadcastTicker = time.NewTicker(time.Second / 60)

	mb.logger.Info("Starting Redis-based movement broadcaster",
		zap.Duration("broadcast_interval", time.Second/60), // ~16.67ms for 60Hz
		zap.String("frequency", "60Hz"),
		zap.Duration("ttl", movingTrainerTTL))

	go mb.broadcastLoop(ctx)
}

// Stop stops the periodic broadcasting
func (mb *MovementBroadcaster) Stop() {
	mb.logger.Info("Stopping Redis-based movement broadcaster")
	
	if mb.broadcastTicker != nil {
		mb.broadcastTicker.Stop()
	}
	
	close(mb.stopChan)
}

// AddMovingTrainer adds a trainer to Redis with TTL
func (mb *MovementBroadcaster) AddMovingTrainer(userID, _, color string) {
	key := movingTrainerKeyPrefix + userID
	value := fmt.Sprintf("%s:%s", userID, color) // Store userID:color
	
	mb.logger.Info("AddMovingTrainer called",
		zap.String("userID", userID),
		zap.String("color", color))
	
	err := mb.redisClient.Set(context.Background(), key, value, movingTrainerTTL).Err()
	if err != nil {
		mb.logger.Error("Failed to add moving trainer to Redis",
			zap.String("userID", userID),
			zap.Error(err))
		return
	}

	mb.logger.Info("Successfully added moving trainer to Redis",
		zap.String("userID", userID),
		zap.String("key", key))
}

// RemoveMovingTrainer removes a trainer from Redis
func (mb *MovementBroadcaster) RemoveMovingTrainer(userID string) {
	key := movingTrainerKeyPrefix + userID
	
	err := mb.redisClient.Del(context.Background(), key).Err()
	if err != nil {
		mb.logger.Error("Failed to remove moving trainer from Redis",
			zap.String("userID", userID),
			zap.Error(err))
		return
	}

	mb.logger.Debug("Removed moving trainer from Redis", zap.String("userID", userID))
}

// UpdateTrainerActivity refreshes the TTL for a moving trainer
func (mb *MovementBroadcaster) UpdateTrainerActivity(userID string) {
	key := movingTrainerKeyPrefix + userID
	
	// Refresh TTL to keep trainer active
	err := mb.redisClient.Expire(context.Background(), key, movingTrainerTTL).Err()
	if err != nil {
		mb.logger.Debug("Failed to refresh moving trainer TTL",
			zap.String("userID", userID),
			zap.Error(err))
	}
}

// broadcastLoop periodically broadcasts positions of moving trainers
func (mb *MovementBroadcaster) broadcastLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mb.stopChan:
			return
		case <-mb.broadcastTicker.C:
			mb.broadcastMovingTrainers(ctx)
		}
	}
}

// broadcastMovingTrainers discovers moving trainers from Redis and broadcasts their positions
func (mb *MovementBroadcaster) broadcastMovingTrainers(ctx context.Context) {
	// Scan Redis for all moving trainer keys
	keys, err := mb.redisClient.Keys(ctx, movingTrainerKeyPrefix+"*").Result()
	if err != nil {
		mb.logger.Error("Failed to scan Redis for moving trainers", zap.Error(err))
		return
	}

	if len(keys) == 0 {
		return
	}

	broadcastCount := 0
	for _, key := range keys {
		// Extract userID from key
		userID := strings.TrimPrefix(key, movingTrainerKeyPrefix)
		
		// Get trainer info from Redis
		value, err := mb.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				mb.logger.Debug("Failed to get moving trainer value",
					zap.String("userID", userID),
					zap.Error(err))
			}
			continue
		}

		// Parse userID:color from value (userID for display, color for rendering)
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			mb.logger.Debug("Invalid trainer value format",
				zap.String("userID", userID),
				zap.String("value", value))
			continue
		}
		_, color := parts[0], parts[1] // parts[0] is userID which we already have

		// Get current trainer state from repository
		trainerEntity, err := mb.repository.GetByID(ctx, trainer.UserID(userID))
		if err != nil {
			mb.logger.Debug("Failed to get moving trainer",
				zap.String("userID", userID),
				zap.Error(err))
			// Remove stale key
			mb.RemoveMovingTrainer(userID)
			continue
		}

		// Update position from movement
		trainerEntity.UpdatePositionFromMovement()

		// Check if trainer is still moving
		if !trainerEntity.Movement.IsMoving {
			// Remove from Redis since trainer stopped moving
			mb.RemoveMovingTrainer(userID)
			continue
		}

		// Create and publish movement event using userID as display name
		event := &cqrscommands.TrainerMovedEvent{
			UserID:    userID,
			Nickname:  userID, // Use userID as display identifier
			Color:     color,
			Position:  trainerEntity.Position,
			Movement:  trainerEntity.Movement,
			Timestamp: time.Now(),
			RequestID: "broadcast-" + userID + "-" + time.Now().Format("150405.000"),
			Changes:   nil, // No changes needed for position broadcasts
		}

		if err := mb.eventBus.Publish(ctx, event); err != nil {
			mb.logger.Error("Failed to publish movement broadcast",
				zap.String("userID", userID),
				zap.Error(err))
		} else {
			broadcastCount++
		}
	}

	if broadcastCount > 0 {
		mb.logger.Debug("Broadcasted positions from Redis",
			zap.Int("moving_trainers", broadcastCount))
	}
}

// GetMovingTrainersCount returns the number of currently moving trainers from Redis
func (mb *MovementBroadcaster) GetMovingTrainersCount() int {
	keys, err := mb.redisClient.Keys(context.Background(), movingTrainerKeyPrefix+"*").Result()
	if err != nil {
		mb.logger.Error("Failed to count moving trainers", zap.Error(err))
		return 0
	}
	return len(keys)
}

// GetCurrentOnlineTrainers returns current positions of all online trainers
func (mb *MovementBroadcaster) GetCurrentOnlineTrainers(ctx context.Context) []cqrscommands.TrainerMovedEvent {
	// Get all moving trainers from Redis
	keys, err := mb.redisClient.Keys(ctx, movingTrainerKeyPrefix+"*").Result()
	if err != nil {
		mb.logger.Error("Failed to get online trainers", zap.Error(err))
		return nil
	}

	var onlineTrainers []cqrscommands.TrainerMovedEvent
	
	for _, key := range keys {
		// Extract userID from key
		userID := strings.TrimPrefix(key, movingTrainerKeyPrefix)
		
		// Get trainer info from Redis
		value, err := mb.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				mb.logger.Debug("Failed to get moving trainer value",
					zap.String("userID", userID),
					zap.Error(err))
			}
			continue
		}

		// Parse userID:color from value
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			mb.logger.Debug("Invalid trainer value format",
				zap.String("userID", userID),
				zap.String("value", value))
			continue
		}
		_, color := parts[0], parts[1]

		// Get current trainer state from repository
		trainerEntity, err := mb.repository.GetByID(ctx, trainer.UserID(userID))
		if err != nil {
			mb.logger.Debug("Failed to get trainer for initial sync",
				zap.String("userID", userID),
				zap.Error(err))
			continue
		}

		// Update position from movement
		trainerEntity.UpdatePositionFromMovement()

		// Add to online trainers list
		onlineTrainers = append(onlineTrainers, cqrscommands.TrainerMovedEvent{
			UserID:    userID,
			Nickname:  userID,
			Color:     color,
			Position:  trainerEntity.Position,
			Movement:  trainerEntity.Movement,
			Timestamp: time.Now(),
			RequestID: "initial-sync-" + userID,
			Changes:   nil,
		})
	}

	mb.logger.Debug("Retrieved online trainers for initial sync",
		zap.Int("count", len(onlineTrainers)))
	
	return onlineTrainers
}

