package account

import (
	"context"
)

// Repository defines the interface for account persistence operations
type Repository interface {
	// FindOneAndInsert inserts a new account
	FindOneAndInsert(ctx context.Context, id AccountID, callback func() (*Account, error)) error

	// FindOneAndUpdate finds an account by ID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id AccountID, callback func(*Account) (*Account, error)) error

	// GetByID retrieves an account by ID (read-only)
	GetByID(ctx context.Context, id AccountID) (*Account, error)

	// GetByProvider retrieves an account by provider and provider user ID
	GetByProvider(ctx context.Context, provider Provider, providerUserID string) (*Account, error)

	// GetByDeviceID retrieves a guest account by device ID
	GetByDeviceID(ctx context.Context, deviceID string) (*Account, error)

	// GetByUserID retrieves an account by user ID (game domain identifier)
	GetByUserID(ctx context.Context, userID UserID) (*Account, error)

	// GetByEmail retrieves any account by email (for cross-provider UserID linking)
	GetByEmail(ctx context.Context, email string) (*Account, error)

	// ListByUserID retrieves all accounts for a specific UserID (N:1 relationship)
	ListByUserID(ctx context.Context, userID UserID) ([]*Account, error)

	// Delete removes an account
	Delete(ctx context.Context, id AccountID) error
}