package trainer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/danghamo/life/internal/domain/shared"
)

// RedisRepository implements Repository using Redis Hash
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis-based trainer repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndUpsert implements IoC pattern with callback for concurrency control
func (r *RedisRepository) FindOneAndUpsert(ctx context.Context, id UserID, callback func(*Trainer) (*Trainer, error)) error {
	key := fmt.Sprintf("trainer:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current trainer
		var current *Trainer
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil && data.Err() != redis.Nil {
			return data.Err()
		}

		if len(data.Val()) > 0 {
			current = &Trainer{}
			if err := r.deserializeTrainer(data.Val(), current); err != nil {
				return err
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

		// Serialize and store
		fields, err := r.serializeTrainer(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

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
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
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

		// Serialize and store
		fields, err := r.serializeTrainer(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

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
		// Get current trainer
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("trainer")
		}

		current := &Trainer{}
		if err := r.deserializeTrainer(data.Val(), current); err != nil {
			return err
		}

		// Execute callback
		result, err := callback(current)
		if err != nil {
			return err
		}

		if result == nil {
			return nil // No changes
		}

		// Serialize and store
		fields, err := r.serializeTrainer(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices if needed
			r.updateTrainerIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves a trainer by UserID
func (r *RedisRepository) GetByID(ctx context.Context, id UserID) (*Trainer, error) {
	key := fmt.Sprintf("trainer:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	t := &Trainer{}
	if err := r.deserializeTrainer(data, t); err != nil {
		return nil, err
	}

	return t, nil
}

// GetByPosition retrieves trainers at a specific position
func (r *RedisRepository) GetByPosition(ctx context.Context, position shared.Position) ([]*Trainer, error) {
	indexKey := fmt.Sprintf("idx:trainer:position:%d:%d", position.X, position.Y)

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
		// Get trainer for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil || len(data.Val()) == 0 {
			return shared.ErrNotFound("trainer")
		}

		t := &Trainer{}
		if err := r.deserializeTrainer(data.Val(), t); err != nil {
			return err
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

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

// serializeTrainer converts trainer to Redis hash fields
func (r *RedisRepository) serializeTrainer(t *Trainer) (map[string]interface{}, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeTrainer converts Redis hash fields to trainer
func (r *RedisRepository) deserializeTrainer(fields map[string]string, t *Trainer) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("trainer data not found in hash")
	}

	return json.Unmarshal([]byte(data), t)
}

// updateTrainerIndices updates secondary indices
func (r *RedisRepository) updateTrainerIndices(ctx context.Context, pipe redis.Pipeliner, t *Trainer) {
	// Position index
	positionKey := fmt.Sprintf("idx:trainer:position:%d:%d", t.Position.X, t.Position.Y)
	pipe.SAdd(ctx, positionKey, t.ID.String())

	// Nickname index
	nicknameKey := fmt.Sprintf("idx:trainer:nickname:%s", t.Nickname.Value())
	pipe.Set(ctx, nicknameKey, t.ID.String(), 0)
}

// cleanupTrainerIndices cleans up secondary indices
func (r *RedisRepository) cleanupTrainerIndices(ctx context.Context, pipe redis.Pipeliner, t *Trainer) {
	// Position index
	positionKey := fmt.Sprintf("idx:trainer:position:%d:%d", t.Position.X, t.Position.Y)
	pipe.SRem(ctx, positionKey, t.ID.String())

	// Nickname index
	nicknameKey := fmt.Sprintf("idx:trainer:nickname:%s", t.Nickname.Value())
	pipe.Del(ctx, nicknameKey)
}
