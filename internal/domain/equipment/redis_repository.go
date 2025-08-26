package equipment

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

// NewRedisRepository creates a new Redis-based equipment repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndUpsert implements IoC pattern with callback for concurrency control
func (r *RedisRepository) FindOneAndUpsert(ctx context.Context, id EquipmentID, callback func(*Equipment) (*Equipment, error)) error {
	key := fmt.Sprintf("equipment:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current equipment
		var current *Equipment
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil && data.Err() != redis.Nil {
			return data.Err()
		}

		if len(data.Val()) > 0 {
			current = &Equipment{}
			if err := r.deserializeEquipment(data.Val(), current); err != nil {
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
		fields, err := r.serializeEquipment(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices
			r.updateEquipmentIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id EquipmentID, callback func() (*Equipment, error)) error {
	key := fmt.Sprintf("equipment:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
			return shared.ErrAlreadyExists("equipment")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil equipment")
		}

		// Serialize and store
		fields, err := r.serializeEquipment(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices
			r.updateEquipmentIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id EquipmentID, callback func(*Equipment) (*Equipment, error)) error {
	key := fmt.Sprintf("equipment:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current equipment
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("equipment")
		}

		current := &Equipment{}
		if err := r.deserializeEquipment(data.Val(), current); err != nil {
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
		fields, err := r.serializeEquipment(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices if needed
			r.updateEquipmentIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves equipment by ID
func (r *RedisRepository) GetByID(ctx context.Context, id EquipmentID) (*Equipment, error) {
	key := fmt.Sprintf("equipment:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	e := &Equipment{}
	if err := r.deserializeEquipment(data, e); err != nil {
		return nil, err
	}

	return e, nil
}

// GetByOwner retrieves equipment owned by an animal
func (r *RedisRepository) GetByOwner(ctx context.Context, ownerID shared.ID) ([]*Equipment, error) {
	indexKey := fmt.Sprintf("idx:equipment:owner:%s", ownerID.String())

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var equipments []*Equipment
	for _, id := range ids {
		e, err := r.GetByID(ctx, EquipmentID(id))
		if err != nil {
			return nil, err
		}
		if e != nil {
			equipments = append(equipments, e)
		}
	}

	return equipments, nil
}

// GetByRarity retrieves equipment by rarity
func (r *RedisRepository) GetByRarity(ctx context.Context, rarity Rarity) ([]*Equipment, error) {
	indexKey := fmt.Sprintf("idx:equipment:rarity:%s", rarity.String())

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var equipments []*Equipment
	for _, id := range ids {
		e, err := r.GetByID(ctx, EquipmentID(id))
		if err != nil {
			return nil, err
		}
		if e != nil && e.Rarity == rarity {
			equipments = append(equipments, e)
		}
	}

	return equipments, nil
}

// GetUnequipped retrieves all unequipped equipment
func (r *RedisRepository) GetUnequipped(ctx context.Context) ([]*Equipment, error) {
	indexKey := "idx:equipment:unequipped"

	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	var equipments []*Equipment
	for _, id := range ids {
		e, err := r.GetByID(ctx, EquipmentID(id))
		if err != nil {
			return nil, err
		}
		if e != nil && !e.IsEquipped() {
			equipments = append(equipments, e)
		}
	}

	return equipments, nil
}

// Delete removes equipment
func (r *RedisRepository) Delete(ctx context.Context, id EquipmentID) error {
	key := fmt.Sprintf("equipment:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get equipment for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil || len(data.Val()) == 0 {
			return shared.ErrNotFound("equipment")
		}

		e := &Equipment{}
		if err := r.deserializeEquipment(data.Val(), e); err != nil {
			return err
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

			// Clean up indices
			r.cleanupEquipmentIndices(ctx, pipe, e)

			return nil
		})

		return err
	}, key)
}

// serializeEquipment converts equipment to Redis hash fields
func (r *RedisRepository) serializeEquipment(e *Equipment) (map[string]interface{}, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeEquipment converts Redis hash fields to equipment
func (r *RedisRepository) deserializeEquipment(fields map[string]string, e *Equipment) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("equipment data not found in hash")
	}

	return json.Unmarshal([]byte(data), e)
}

// updateEquipmentIndices updates secondary indices
func (r *RedisRepository) updateEquipmentIndices(ctx context.Context, pipe redis.Pipeliner, e *Equipment) {
	// Owner index (only for equipped equipment)
	if e.IsEquipped() {
		ownerKey := fmt.Sprintf("idx:equipment:owner:%s", e.OwnerID.String())
		pipe.SAdd(ctx, ownerKey, e.ID.String())

		// Remove from unequipped index
		pipe.SRem(ctx, "idx:equipment:unequipped", e.ID.String())
	} else {
		// Add to unequipped index
		pipe.SAdd(ctx, "idx:equipment:unequipped", e.ID.String())
	}

	// Type index
	typeKey := fmt.Sprintf("idx:equipment:type:%s", e.EquipmentType.String())
	pipe.SAdd(ctx, typeKey, e.ID.String())

	// Rarity index
	rarityKey := fmt.Sprintf("idx:equipment:rarity:%s", e.Rarity.String())
	pipe.SAdd(ctx, rarityKey, e.ID.String())
}

// cleanupEquipmentIndices cleans up secondary indices
func (r *RedisRepository) cleanupEquipmentIndices(ctx context.Context, pipe redis.Pipeliner, e *Equipment) {
	// Owner index
	if e.IsEquipped() {
		ownerKey := fmt.Sprintf("idx:equipment:owner:%s", e.OwnerID.String())
		pipe.SRem(ctx, ownerKey, e.ID.String())
	} else {
		// Remove from unequipped index
		pipe.SRem(ctx, "idx:equipment:unequipped", e.ID.String())
	}

	// Type index
	typeKey := fmt.Sprintf("idx:equipment:type:%s", e.EquipmentType.String())
	pipe.SRem(ctx, typeKey, e.ID.String())

	// Rarity index
	rarityKey := fmt.Sprintf("idx:equipment:rarity:%s", e.Rarity.String())
	pipe.SRem(ctx, rarityKey, e.ID.String())
}
