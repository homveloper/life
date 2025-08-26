package animal

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

// NewRedisRepository creates a new Redis-based animal repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndUpsert implements IoC pattern with callback for concurrency control
func (r *RedisRepository) FindOneAndUpsert(ctx context.Context, id AnimalID, callback func(*Animal) (*Animal, error)) error {
	key := fmt.Sprintf("animal:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current animal
		var current *Animal
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil && data.Err() != redis.Nil {
			return data.Err()
		}

		if len(data.Val()) > 0 {
			current = &Animal{}
			if err := r.deserializeAnimal(data.Val(), current); err != nil {
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
		fields, err := r.serializeAnimal(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices
			r.updateAnimalIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id AnimalID, callback func() (*Animal, error)) error {
	key := fmt.Sprintf("animal:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
			return shared.ErrAlreadyExists("animal")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil animal")
		}

		// Serialize and store
		fields, err := r.serializeAnimal(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices
			r.updateAnimalIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id AnimalID, callback func(*Animal) (*Animal, error)) error {
	key := fmt.Sprintf("animal:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current animal
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("animal")
		}

		current := &Animal{}
		if err := r.deserializeAnimal(data.Val(), current); err != nil {
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
		fields, err := r.serializeAnimal(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices if needed
			r.updateAnimalIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves an animal by ID
func (r *RedisRepository) GetByID(ctx context.Context, id AnimalID) (*Animal, error) {
	key := fmt.Sprintf("animal:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	a := &Animal{}
	if err := r.deserializeAnimal(data, a); err != nil {
		return nil, err
	}

	return a, nil
}

// GetByPosition retrieves wild animals at a specific position
func (r *RedisRepository) GetByPosition(ctx context.Context, position shared.Position) ([]*Animal, error) {
	indexKey := fmt.Sprintf("idx:animal:wild_position:%d:%d", position.X, position.Y)

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var animals []*Animal
	for _, id := range ids {
		a, err := r.GetByID(ctx, AnimalID(id))
		if err != nil {
			return nil, err
		}
		if a != nil && a.IsWild() {
			animals = append(animals, a)
		}
	}

	return animals, nil
}

// GetWildAnimalsNearby retrieves wild animals within radius from position
func (r *RedisRepository) GetWildAnimalsNearby(ctx context.Context, center shared.Position, radius int) ([]*Animal, error) {
	var animals []*Animal

	// Simple implementation: check positions within square radius
	for x := center.X - radius; x <= center.X+radius; x++ {
		for y := center.Y - radius; y <= center.Y+radius; y++ {
			position := shared.NewPosition(x, y)
			nearby, err := r.GetByPosition(ctx, position)
			if err != nil {
				return nil, err
			}
			animals = append(animals, nearby...)
		}
	}

	return animals, nil
}

// GetByOwner retrieves animals owned by a trainer
func (r *RedisRepository) GetByOwner(ctx context.Context, ownerID shared.ID) ([]*Animal, error) {
	indexKey := fmt.Sprintf("idx:animal:owner:%s", ownerID.String())

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var animals []*Animal
	for _, id := range ids {
		a, err := r.GetByID(ctx, AnimalID(id))
		if err != nil {
			return nil, err
		}
		if a != nil {
			animals = append(animals, a)
		}
	}

	return animals, nil
}

// GetByState retrieves animals by state
func (r *RedisRepository) GetByState(ctx context.Context, state AnimalState) ([]*Animal, error) {
	indexKey := fmt.Sprintf("idx:animal:state:%s", state.String())

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var animals []*Animal
	for _, id := range ids {
		a, err := r.GetByID(ctx, AnimalID(id))
		if err != nil {
			return nil, err
		}
		if a != nil && a.State == state {
			animals = append(animals, a)
		}
	}

	return animals, nil
}

// GetByType retrieves animals by type
func (r *RedisRepository) GetByType(ctx context.Context, animalType AnimalType) ([]*Animal, error) {
	indexKey := fmt.Sprintf("idx:animal:type:%s", animalType.String())

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var animals []*Animal
	for _, id := range ids {
		a, err := r.GetByID(ctx, AnimalID(id))
		if err != nil {
			return nil, err
		}
		if a != nil && a.AnimalType == animalType {
			animals = append(animals, a)
		}
	}

	return animals, nil
}

// Delete removes an animal
func (r *RedisRepository) Delete(ctx context.Context, id AnimalID) error {
	key := fmt.Sprintf("animal:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get animal for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil || len(data.Val()) == 0 {
			return shared.ErrNotFound("animal")
		}

		a := &Animal{}
		if err := r.deserializeAnimal(data.Val(), a); err != nil {
			return err
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

			// Clean up indices
			r.cleanupAnimalIndices(ctx, pipe, a)

			return nil
		})

		return err
	}, key)
}

// serializeAnimal converts animal to Redis hash fields
func (r *RedisRepository) serializeAnimal(a *Animal) (map[string]interface{}, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeAnimal converts Redis hash fields to animal
func (r *RedisRepository) deserializeAnimal(fields map[string]string, a *Animal) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("animal data not found in hash")
	}

	return json.Unmarshal([]byte(data), a)
}

// updateAnimalIndices updates secondary indices
func (r *RedisRepository) updateAnimalIndices(ctx context.Context, pipe redis.Pipeliner, a *Animal) {
	// Position index (only for wild animals)
	if a.IsWild() {
		positionKey := fmt.Sprintf("idx:animal:wild_position:%d:%d", a.Position.X, a.Position.Y)
		pipe.SAdd(ctx, positionKey, a.ID.String())
	}

	// Owner index (only for captured animals)
	if a.IsCaptured() {
		ownerKey := fmt.Sprintf("idx:animal:owner:%s", a.OwnerID.String())
		pipe.SAdd(ctx, ownerKey, a.ID.String())
	}

	// State index
	stateKey := fmt.Sprintf("idx:animal:state:%s", a.State.String())
	pipe.SAdd(ctx, stateKey, a.ID.String())

	// Type index
	typeKey := fmt.Sprintf("idx:animal:type:%s", a.AnimalType.String())
	pipe.SAdd(ctx, typeKey, a.ID.String())
}

// cleanupAnimalIndices cleans up secondary indices
func (r *RedisRepository) cleanupAnimalIndices(ctx context.Context, pipe redis.Pipeliner, a *Animal) {
	// Position index
	if a.IsWild() {
		positionKey := fmt.Sprintf("idx:animal:wild_position:%d:%d", a.Position.X, a.Position.Y)
		pipe.SRem(ctx, positionKey, a.ID.String())
	}

	// Owner index
	if a.IsCaptured() {
		ownerKey := fmt.Sprintf("idx:animal:owner:%s", a.OwnerID.String())
		pipe.SRem(ctx, ownerKey, a.ID.String())
	}

	// State index
	stateKey := fmt.Sprintf("idx:animal:state:%s", a.State.String())
	pipe.SRem(ctx, stateKey, a.ID.String())

	// Type index
	typeKey := fmt.Sprintf("idx:animal:type:%s", a.AnimalType.String())
	pipe.SRem(ctx, typeKey, a.ID.String())
}
