package auth

import (
	"context"
)

// Repository defines the interface for user persistence operations
type Repository interface {
	// FindOneAndInsert inserts a new user
	FindOneAndInsert(ctx context.Context, id UserID, callback func() (*User, error)) error

	// FindOneAndUpdate finds a user by ID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id UserID, callback func(*User) (*User, error)) error

	// GetByID retrieves a user by ID (read-only)
	GetByID(ctx context.Context, id UserID) (*User, error)

	// GetByUsername retrieves a user by username (read-only)
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Delete removes a user
	Delete(ctx context.Context, id UserID) error
}