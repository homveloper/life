package trainer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/danghamo/life/internal/domain/shared"
)

// RedisRepository implements Repository using Redis JSON
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis JSON-based trainer repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndUpsert implements IoC pattern with callback for concurrency control
func (r *RedisRepository) FindOneAndUpsert(ctx context.Context, id UserID, callback func(*Trainer) (*Trainer, error)) error {
	key := fmt.Sprintf("trainer:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current trainer using JSON.GET
		var current *Trainer
		jsonData, err := tx.JSONGet(ctx, key, "$").Result()
		if err == redis.Nil {
			// Key doesn't exist, current remains nil
			current = nil
		} else if err != nil {
			return err
		} else if jsonData == "" || jsonData == "null" {
			// JSON exists but is null
			current = nil
		} else {
			// Key exists, parse JSON array result
			var jsonArray []json.RawMessage
			if err := json.Unmarshal([]byte(jsonData), &jsonArray); err != nil {
				return fmt.Errorf("failed to parse JSON array from Redis: %w", err)
			}
			
			if len(jsonArray) > 0 {
				current = &Trainer{}
				if err := json.Unmarshal(jsonArray[0], current); err != nil {
					return fmt.Errorf("failed to deserialize trainer: %w", err)
				}
			}
		}

		// Execute callback
		result, err := callback(current)
		if err != nil {
			return err
		}

		if result == nil {
			return nil // No changes
		}

		// Serialize and store using JSON.SET
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to serialize trainer: %w", err)
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.JSONSet(ctx, key, "$", string(jsonBytes))

			// Update indices
			r.updateTrainerIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id UserID, callback func() (*Trainer, error)) error {
	key := fmt.Sprintf("trainer:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists using JSON.GET
		jsonData, err := tx.JSONGet(ctx, key, "$").Result()
		if err == redis.Nil {
			// Key doesn't exist, proceed with insert
		} else if err != nil {
			return err
		} else if jsonData != "" && jsonData != "null" {
			// Key exists and has valid data, return error
			return shared.ErrAlreadyExists("trainer")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil trainer")
		}

		// Serialize and store using JSON.SET
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to serialize trainer: %w", err)
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.JSONSet(ctx, key, "$", string(jsonBytes))

			// Update indices
			r.updateTrainerIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id UserID, callback func(*Trainer) (*Trainer, error)) error {
	key := fmt.Sprintf("trainer:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current trainer using JSON.GET
		jsonData, err := tx.JSONGet(ctx, key, "$").Result()
		if err == redis.Nil {
			return shared.ErrNotFound("trainer")
		}
		if err != nil {
			return err
		}

		// Check for null response
		if jsonData == "" || jsonData == "null" {
			return shared.ErrNotFound("trainer")
		}

		// Parse JSON array result
		var jsonArray []json.RawMessage
		if err := json.Unmarshal([]byte(jsonData), &jsonArray); err != nil {
			return fmt.Errorf("failed to parse JSON array from Redis: %w", err)
		}
		
		if len(jsonArray) == 0 {
			return shared.ErrNotFound("trainer")
		}

		current := &Trainer{}
		if err := json.Unmarshal(jsonArray[0], current); err != nil {
			return fmt.Errorf("failed to deserialize trainer: %w", err)
		}

		// Execute callback
		updateResult, err := callback(current)
		if err != nil {
			return err
		}

		if updateResult == nil {
			return nil // No changes
		}

		// Serialize and store using JSON.SET
		jsonBytes, err := json.Marshal(updateResult)
		if err != nil {
			return fmt.Errorf("failed to serialize trainer: %w", err)
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.JSONSet(ctx, key, "$", string(jsonBytes))

			// Update indices if needed
			r.updateTrainerIndices(ctx, pipe, updateResult)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves a trainer by UserID using Redis JSON
func (r *RedisRepository) GetByID(ctx context.Context, id UserID) (*Trainer, error) {
	key := fmt.Sprintf("trainer:%s", id.String())

	jsonData, err := r.client.JSONGet(ctx, key, "$").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trainer from Redis: %w", err)
	}

	// Check for null response
	if jsonData == "" || jsonData == "null" {
		return nil, nil
	}

	// Parse JSON array result
	var jsonArray []json.RawMessage
	if err := json.Unmarshal([]byte(jsonData), &jsonArray); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array from Redis: %w", err)
	}
	
	if len(jsonArray) == 0 {
		return nil, nil // Path doesn't exist
	}

	t := &Trainer{}
	if err := json.Unmarshal(jsonArray[0], t); err != nil {
		return nil, fmt.Errorf("failed to deserialize trainer: %w", err)
	}

	return t, nil
}

// GetByPosition retrieves trainers at a specific position
func (r *RedisRepository) GetByPosition(ctx context.Context, position shared.Position) ([]*Trainer, error) {
	indexKey := fmt.Sprintf("idx:trainer:position:%.1f:%.1f", position.X, position.Y)

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var trainers []*Trainer
	for _, id := range ids {
		t, err := r.GetByID(ctx, UserID(id))
		if err != nil {
			return nil, err
		}
		if t != nil {
			trainers = append(trainers, t)
		}
	}

	return trainers, nil
}

// Delete removes a trainer
func (r *RedisRepository) Delete(ctx context.Context, id UserID) error {
	key := fmt.Sprintf("trainer:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get trainer for index cleanup using JSON.GET
		jsonData, err := tx.JSONGet(ctx, key, "$").Result()
		if err == redis.Nil {
			return shared.ErrNotFound("trainer")
		}
		if err != nil {
			return err
		}

		// Check for null response
		if jsonData == "" || jsonData == "null" {
			return shared.ErrNotFound("trainer")
		}

		// Parse JSON array result
		var jsonArray []json.RawMessage
		if err := json.Unmarshal([]byte(jsonData), &jsonArray); err != nil {
			return fmt.Errorf("failed to parse JSON array from Redis: %w", err)
		}
		
		if len(jsonArray) == 0 {
			return shared.ErrNotFound("trainer")
		}

		t := &Trainer{}
		if err := json.Unmarshal(jsonArray[0], t); err != nil {
			return fmt.Errorf("failed to deserialize trainer: %w", err)
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.JSONDel(ctx, key, "$")

			// Clean up indices
			r.cleanupTrainerIndices(ctx, pipe, t)

			return nil
		})

		return err
	}, key)
}

// FindByNickname finds a trainer by nickname
func (r *RedisRepository) FindByNickname(ctx context.Context, nickname string) (*Trainer, error) {
	indexKey := fmt.Sprintf("idx:trainer:nickname:%s", nickname)

	id, err := r.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, UserID(id))
}

// updateTrainerIndices updates secondary indices
func (r *RedisRepository) updateTrainerIndices(ctx context.Context, pipe redis.Pipeliner, t *Trainer) {
	// Position index
	positionKey := fmt.Sprintf("idx:trainer:position:%.1f:%.1f", t.Position.X, t.Position.Y)
	pipe.SAdd(ctx, positionKey, t.ID.String())

	// Nickname index
	nicknameKey := fmt.Sprintf("idx:trainer:nickname:%s", t.Nickname)
	pipe.Set(ctx, nicknameKey, t.ID.String(), 0)
}

// cleanupTrainerIndices cleans up secondary indices
func (r *RedisRepository) cleanupTrainerIndices(ctx context.Context, pipe redis.Pipeliner, t *Trainer) {
	// Position index
	positionKey := fmt.Sprintf("idx:trainer:position:%.1f:%.1f", t.Position.X, t.Position.Y)
	pipe.SRem(ctx, positionKey, t.ID.String())

	// Nickname index
	nicknameKey := fmt.Sprintf("idx:trainer:nickname:%s", t.Nickname)
	pipe.Del(ctx, nicknameKey)
}

// GetAll retrieves all trainers from Redis using JSON
func (r *RedisRepository) GetAll(ctx context.Context) ([]*Trainer, error) {
	// Use SCAN to get all trainer keys
	pattern := "trainer:*"
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	
	var trainers []*Trainer
	
	for iter.Next(ctx) {
		key := iter.Val()
		
		// Get trainer data using JSON.GET
		jsonData, err := r.client.JSONGet(ctx, key, "$").Result()
		if err != nil {
			continue // Skip errors for individual trainers
		}
		
		// Skip null responses
		if jsonData == "" || jsonData == "null" {
			continue // Skip null entries
		}
		
		// Parse JSON array result
		var jsonArray []json.RawMessage
		if err := json.Unmarshal([]byte(jsonData), &jsonArray); err != nil {
			continue // Skip malformed JSON arrays
		}
		
		if len(jsonArray) == 0 {
			continue // Skip empty arrays
		}
		
		trainer := &Trainer{}
		if err := json.Unmarshal(jsonArray[0], trainer); err != nil {
			continue // Skip malformed entries
		}
		
		trainers = append(trainers, trainer)
	}
	
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan trainer keys: %w", err)
	}
	
	return trainers, nil
}