package trainer

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danghamo/life/internal/domain/shared"
)

// setupTestRedis creates a Redis client for testing
func setupTestRedis(t *testing.T) *redis.Client {
	// Skip test if REDIS_URL is not set
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL environment variable not set, skipping Redis integration tests")
	}

	opt, err := redis.ParseURL(redisURL)
	require.NoError(t, err, "Failed to parse Redis URL")

	client := redis.NewClient(opt)
	
	// Test connection
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	require.NoError(t, err, "Failed to connect to Redis")

	return client
}

// createTestTrainer creates a test trainer instance
func createTestTrainer() *Trainer {
	userID := UserID("test-user-123")
	nickname := "TestTrainer"
	
	// Create trainer
	trainer, _ := NewTrainer(userID, nickname)
	// Set position after creation
	trainer.Position = shared.Position{X: 10.0, Y: 20.0}
	return trainer
}

// Tests for GetByID method
func TestRedisRepository_GetByID(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()
	
	repo := NewRedisRepository(client)
	ctx := context.Background()
	
	t.Run("should return nil when trainer does not exist", func(t *testing.T) {
		// Test with non-existent trainer
		userID := UserID("non-existent-user")
		result, err := repo.GetByID(ctx, userID)
		
		// Assert
		require.NoError(t, err)
		assert.Nil(t, result)
	})
	
	t.Run("should return trainer after insert and get", func(t *testing.T) {
		// Setup test data
		trainer := createTestTrainer()
		
		// Insert trainer first
		err := repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		require.NoError(t, err)
		
		// Test GetByID
		result, err := repo.GetByID(ctx, trainer.ID)
		
		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, trainer.ID, result.ID)
		assert.Equal(t, trainer.Nickname, result.Nickname)
		assert.Equal(t, trainer.Position.X, result.Position.X)
		assert.Equal(t, trainer.Position.Y, result.Position.Y)
		
		// Cleanup
		repo.Delete(ctx, trainer.ID)
	})
}

// Tests for FindOneAndInsert method  
func TestRedisRepository_FindOneAndInsert(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()
	
	repo := NewRedisRepository(client)
	ctx := context.Background()
	
	t.Run("should insert new trainer when not exists", func(t *testing.T) {
		// Setup
		trainer := createTestTrainer()
		trainer.ID = UserID(fmt.Sprintf("test-insert-%s", t.Name()))
		
		// Test
		err := repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		
		// Assert
		require.NoError(t, err)
		
		// Verify data was stored
		result, err := repo.GetByID(ctx, trainer.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, trainer.ID, result.ID)
		
		// Cleanup
		repo.Delete(ctx, trainer.ID)
	})
	
	t.Run("should return error when trainer already exists", func(t *testing.T) {
		// Setup existing trainer
		trainer := createTestTrainer()
		trainer.ID = UserID(fmt.Sprintf("test-duplicate-%s", t.Name()))
		
		// Insert first time
		err := repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		require.NoError(t, err)
		
		// Test duplicate insert
		err = repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		
		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		
		// Cleanup
		repo.Delete(ctx, trainer.ID)
	})
}

// Tests for FindOneAndUpdate method
func TestRedisRepository_FindOneAndUpdate(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()
	
	repo := NewRedisRepository(client)
	ctx := context.Background()
	
	t.Run("should update existing trainer", func(t *testing.T) {
		// Setup existing trainer
		trainer := createTestTrainer()
		trainer.ID = UserID(fmt.Sprintf("test-update-%s", t.Name()))
		
		// Insert first
		err := repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		require.NoError(t, err)
		
		// Test update
		newPosition := shared.Position{X: 50.0, Y: 60.0}
		err = repo.FindOneAndUpdate(ctx, trainer.ID, func(t *Trainer) (*Trainer, error) {
			t.Position = newPosition
			return t, nil
		})
		
		// Assert
		require.NoError(t, err)
		
		// Verify data was updated
		result, err := repo.GetByID(ctx, trainer.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, newPosition.X, result.Position.X)
		assert.Equal(t, newPosition.Y, result.Position.Y)
		
		// Cleanup
		repo.Delete(ctx, trainer.ID)
	})
	
	t.Run("should return error when trainer not found", func(t *testing.T) {
		// Test update on non-existent trainer
		userID := UserID("non-existent-update")
		err := repo.FindOneAndUpdate(ctx, userID, func(t *Trainer) (*Trainer, error) {
			return t, nil
		})
		
		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// Test for Delete method
func TestRedisRepository_Delete(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()
	
	repo := NewRedisRepository(client)
	ctx := context.Background()
	
	t.Run("should delete existing trainer", func(t *testing.T) {
		// Setup existing trainer
		trainer := createTestTrainer()
		trainer.ID = UserID(fmt.Sprintf("test-delete-%s", t.Name()))
		
		// Insert first
		err := repo.FindOneAndInsert(ctx, trainer.ID, func() (*Trainer, error) {
			return trainer, nil
		})
		require.NoError(t, err)
		
		// Verify it exists
		result, err := repo.GetByID(ctx, trainer.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		
		// Test delete
		err = repo.Delete(ctx, trainer.ID)
		require.NoError(t, err)
		
		// Verify it's gone
		result, err = repo.GetByID(ctx, trainer.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
	
	t.Run("should return error when trainer not found", func(t *testing.T) {
		// Test delete on non-existent trainer
		userID := UserID("non-existent-delete")
		err := repo.Delete(ctx, userID)
		
		// Assert  
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}