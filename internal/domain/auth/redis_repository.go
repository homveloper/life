package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/danghamo/life/internal/domain/shared"
)

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis-based auth repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id UserID, callback func() (*User, error)) error {
	key := fmt.Sprintf("user:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
			return shared.ErrAlreadyExists("user")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil user")
		}

		// Serialize and store
		fields, err := r.serializeUser(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update username index
			r.updateUserIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id UserID, callback func(*User) (*User, error)) error {
	key := fmt.Sprintf("user:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current user
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("user")
		}

		current := &User{}
		if err := r.deserializeUser(data.Val(), current); err != nil {
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
		fields, err := r.serializeUser(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices if needed
			r.updateUserIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves a user by ID
func (r *RedisRepository) GetByID(ctx context.Context, id UserID) (*User, error) {
	key := fmt.Sprintf("user:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	u := &User{}
	if err := r.deserializeUser(data, u); err != nil {
		return nil, err
	}

	return u, nil
}

// GetByUsername retrieves a user by username
func (r *RedisRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	indexKey := fmt.Sprintf("idx:user:username:%s", username)

	id, err := r.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, UserID(id))
}

// Delete removes a user
func (r *RedisRepository) Delete(ctx context.Context, id UserID) error {
	key := fmt.Sprintf("user:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get user for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil || len(data.Val()) == 0 {
			return shared.ErrNotFound("user")
		}

		u := &User{}
		if err := r.deserializeUser(data.Val(), u); err != nil {
			return err
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

			// Clean up indices
			r.cleanupUserIndices(ctx, pipe, u)

			return nil
		})

		return err
	}, key)
}

// serializeUser converts user to Redis hash fields
func (r *RedisRepository) serializeUser(u *User) (map[string]interface{}, error) {
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeUser converts Redis hash fields to user
func (r *RedisRepository) deserializeUser(fields map[string]string, u *User) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("user data not found in hash")
	}

	return json.Unmarshal([]byte(data), u)
}

// updateUserIndices updates secondary indices
func (r *RedisRepository) updateUserIndices(ctx context.Context, pipe redis.Pipeliner, u *User) {
	// Username index
	usernameKey := fmt.Sprintf("idx:user:username:%s", u.Username.Value())
	pipe.Set(ctx, usernameKey, u.ID.String(), 0)
}

// cleanupUserIndices cleans up secondary indices
func (r *RedisRepository) cleanupUserIndices(ctx context.Context, pipe redis.Pipeliner, u *User) {
	// Username index
	usernameKey := fmt.Sprintf("idx:user:username:%s", u.Username.Value())
	pipe.Del(ctx, usernameKey)
}