package world

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

// NewRedisRepository creates a new Redis-based world repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndUpsert implements IoC pattern with callback for concurrency control
func (r *RedisRepository) FindOneAndUpsert(ctx context.Context, id WorldID, callback func(*World) (*World, error)) error {
	key := fmt.Sprintf("world:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current world
		var current *World
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil && data.Err() != redis.Nil {
			return data.Err()
		}

		if len(data.Val()) > 0 {
			current = &World{}
			if err := r.deserializeWorld(data.Val(), current); err != nil {
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
		fields, err := r.serializeWorld(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			return nil
		})

		return err
	}, key)
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id WorldID, callback func() (*World, error)) error {
	key := fmt.Sprintf("world:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
			return shared.ErrAlreadyExists("world")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil world")
		}

		// Serialize and store
		fields, err := r.serializeWorld(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id WorldID, callback func(*World) (*World, error)) error {
	key := fmt.Sprintf("world:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current world
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("world")
		}

		current := &World{}
		if err := r.deserializeWorld(data.Val(), current); err != nil {
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
		fields, err := r.serializeWorld(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves a world by ID
func (r *RedisRepository) GetByID(ctx context.Context, id WorldID) (*World, error) {
	key := fmt.Sprintf("world:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	w := &World{}
	if err := r.deserializeWorld(data, w); err != nil {
		return nil, err
	}

	return w, nil
}

// GetTileEntities retrieves all entities at a specific position
func (r *RedisRepository) GetTileEntities(ctx context.Context, worldID WorldID, position shared.Position) ([]shared.ID, error) {
	world, err := r.GetByID(ctx, worldID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, shared.ErrNotFound("world")
	}

	tile, err := world.GetTile(position)
	if err != nil {
		return nil, err
	}

	// Return copy of entities
	entities := make([]shared.ID, len(tile.Entities))
	copy(entities, tile.Entities)

	return entities, nil
}

// Delete removes a world
func (r *RedisRepository) Delete(ctx context.Context, id WorldID) error {
	key := fmt.Sprintf("world:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() == 0 {
			return shared.ErrNotFound("world")
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

			return nil
		})

		return err
	}, key)
}

// serializeWorld converts world to Redis hash fields
func (r *RedisRepository) serializeWorld(w *World) (map[string]interface{}, error) {
	data, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeWorld converts Redis hash fields to world
func (r *RedisRepository) deserializeWorld(fields map[string]string, w *World) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("world data not found in hash")
	}

	return json.Unmarshal([]byte(data), w)
}
