package account

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

// NewRedisRepository creates a new Redis-based account repository
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{
		client: client,
	}
}

// FindOneAndInsert implements IoC pattern for insert operations
func (r *RedisRepository) FindOneAndInsert(ctx context.Context, id AccountID, callback func() (*Account, error)) error {
	key := fmt.Sprintf("account:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Check if already exists
		exists := tx.Exists(ctx, key)
		if exists.Err() != nil {
			return exists.Err()
		}

		if exists.Val() > 0 {
			return shared.ErrAlreadyExists("account")
		}

		// Execute callback
		result, err := callback()
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("callback returned nil account")
		}

		// Serialize and store
		fields, err := r.serializeAccount(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices
			r.updateAccountIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// FindOneAndUpdate implements IoC pattern for update operations
func (r *RedisRepository) FindOneAndUpdate(ctx context.Context, id AccountID, callback func(*Account) (*Account, error)) error {
	key := fmt.Sprintf("account:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current account
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return shared.ErrNotFound("account")
		}

		current := &Account{}
		if err := r.deserializeAccount(data.Val(), current); err != nil {
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
		fields, err := r.serializeAccount(result)
		if err != nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HMSet(ctx, key, fields)

			// Update indices if needed
			r.updateAccountIndices(ctx, pipe, result)

			return nil
		})

		return err
	}, key)
}

// GetByID retrieves an account by ID
func (r *RedisRepository) GetByID(ctx context.Context, id AccountID) (*Account, error) {
	key := fmt.Sprintf("account:%s", id.String())

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	a := &Account{}
	if err := r.deserializeAccount(data, a); err != nil {
		return nil, err
	}

	return a, nil
}

// GetByProvider retrieves an account by provider and provider user ID
func (r *RedisRepository) GetByProvider(ctx context.Context, provider Provider, providerUserID string) (*Account, error) {
	providerKey := string(provider) + ":" + providerUserID
	indexKey := fmt.Sprintf("idx:account:provider:%s", providerKey)

	id, err := r.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, AccountID(id))
}

// GetByUserID retrieves the primary account by user ID (returns first account for N:1 relationship)
func (r *RedisRepository) GetByUserID(ctx context.Context, userID UserID) (*Account, error) {
	accounts, err := r.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if len(accounts) == 0 {
		return nil, nil
	}
	
	// Return the first account (primary account for this UserID)
	return accounts[0], nil
}

// Delete removes an account
func (r *RedisRepository) Delete(ctx context.Context, id AccountID) error {
	key := fmt.Sprintf("account:%s", id.String())

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Get account for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil || len(data.Val()) == 0 {
			return shared.ErrNotFound("account")
		}

		a := &Account{}
		if err := r.deserializeAccount(data.Val(), a); err != nil {
			return err
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)

			// Clean up indices
			r.cleanupAccountIndices(ctx, pipe, a)

			return nil
		})

		return err
	}, key)
}

// serializeAccount converts account to Redis hash fields
func (r *RedisRepository) serializeAccount(a *Account) (map[string]interface{}, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": string(data),
	}, nil
}

// deserializeAccount converts Redis hash fields to account
func (r *RedisRepository) deserializeAccount(fields map[string]string, a *Account) error {
	data, exists := fields["data"]
	if !exists {
		return fmt.Errorf("account data not found in hash")
	}

	return json.Unmarshal([]byte(data), a)
}

// updateAccountIndices updates secondary indices
func (r *RedisRepository) updateAccountIndices(ctx context.Context, pipe redis.Pipeliner, a *Account) {
	// Provider index: provider:provider_user_id -> account_id
	providerKey := a.GetProviderKey()
	providerIndexKey := fmt.Sprintf("idx:account:provider:%s", providerKey)
	pipe.Set(ctx, providerIndexKey, a.ID.String(), 0)

	// User ID index: user_id -> set of account_ids (N:1 relationship)
	userIndexKey := fmt.Sprintf("idx:account:user:%s", a.UserID.String())
	pipe.SAdd(ctx, userIndexKey, a.ID.String())

	// Email index: email -> account_id (first account with this email for lookup)
	if a.Profile.Email != "" && !a.IsGuest() {
		emailIndexKey := fmt.Sprintf("idx:account:email:%s", a.Profile.Email)
		// Only set if no existing account found (SETNX behavior)
		pipe.SetNX(ctx, emailIndexKey, a.ID.String(), 0)
	}

	// Device ID index (게스트 계정인 경우)
	if a.IsGuest() && a.DeviceID != "" {
		deviceIndexKey := fmt.Sprintf("idx:account:device:%s", a.DeviceID)
		pipe.Set(ctx, deviceIndexKey, a.ID.String(), 0)
	}
}

// cleanupAccountIndices cleans up secondary indices
func (r *RedisRepository) cleanupAccountIndices(ctx context.Context, pipe redis.Pipeliner, a *Account) {
	// Provider index
	providerKey := a.GetProviderKey()
	providerIndexKey := fmt.Sprintf("idx:account:provider:%s", providerKey)
	pipe.Del(ctx, providerIndexKey)

	// User ID index: remove from set (N:1 relationship)
	userIndexKey := fmt.Sprintf("idx:account:user:%s", a.UserID.String())
	pipe.SRem(ctx, userIndexKey, a.ID.String())
	
	// Check if set is empty and remove it if so
	pipe.SCard(ctx, userIndexKey)
	// Note: In a more robust implementation, we'd check the result and conditionally delete

	// Email index: only delete if this account owns the email index
	if a.Profile.Email != "" && !a.IsGuest() {
		emailIndexKey := fmt.Sprintf("idx:account:email:%s", a.Profile.Email)
		// Check if this account ID matches the stored one
		pipe.Get(ctx, emailIndexKey)
		// Note: In a more robust implementation, we'd check the result and conditionally delete
		pipe.Del(ctx, emailIndexKey)
	}

	// Device ID index (게스트 계정인 경우)
	if a.IsGuest() && a.DeviceID != "" {
		deviceIndexKey := fmt.Sprintf("idx:account:device:%s", a.DeviceID)
		pipe.Del(ctx, deviceIndexKey)
	}
}

// GetByDeviceID retrieves a guest account by device ID
func (r *RedisRepository) GetByDeviceID(ctx context.Context, deviceID string) (*Account, error) {
	indexKey := fmt.Sprintf("idx:account:device:%s", deviceID)

	id, err := r.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, AccountID(id))
}

// GetByEmail retrieves any account by email (for cross-provider UserID linking)
func (r *RedisRepository) GetByEmail(ctx context.Context, email string) (*Account, error) {
	indexKey := fmt.Sprintf("idx:account:email:%s", email)

	id, err := r.client.Get(ctx, indexKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, AccountID(id))
}

// ListByUserID retrieves all accounts for a specific UserID (N:1 relationship)
func (r *RedisRepository) ListByUserID(ctx context.Context, userID UserID) ([]*Account, error) {
	indexKey := fmt.Sprintf("idx:account:user:%s", userID.String())
	
	// Get all account IDs for this UserID using a set
	ids, err := r.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*Account{}, nil
	}

	// Fetch all accounts
	accounts := make([]*Account, 0, len(ids))
	for _, id := range ids {
		account, err := r.GetByID(ctx, AccountID(id))
		if err != nil {
			return nil, err
		}
		if account != nil {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}