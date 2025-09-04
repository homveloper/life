package trainer

import (
	"context"

	"github.com/danghamo/life/internal/domain/shared"
)

// Repository defines the interface for trainer persistence operations with IoC pattern
type Repository interface {
	// FindOneAndUpsert finds a trainer by UserID and applies callback for atomic upsert
	FindOneAndUpsert(ctx context.Context, id UserID, callback func(*Trainer) (*Trainer, error)) error

	// FindOneAndInsert inserts a new trainer with callback for initialization
	FindOneAndInsert(ctx context.Context, id UserID, callback func() (*Trainer, error)) error

	// FindOneAndUpdate finds a trainer by UserID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id UserID, callback func(*Trainer) (*Trainer, error)) error

	// GetByID retrieves a trainer by UserID (read-only)
	GetByID(ctx context.Context, id UserID) (*Trainer, error)

	// GetByPosition retrieves trainers at a specific position (read-only)
	GetByPosition(ctx context.Context, position shared.Position) ([]*Trainer, error)

	// Delete removes a trainer
	Delete(ctx context.Context, id UserID) error

	// FindByNickname finds a trainer by nickname (for uniqueness checking)
	FindByNickname(ctx context.Context, nickname string) (*Trainer, error)

	// GetAll retrieves all trainers (read-only)
	GetAll(ctx context.Context) ([]*Trainer, error)
}
