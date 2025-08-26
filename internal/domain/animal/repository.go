package animal

import (
	"context"

	"github.com/danghamo/life/internal/domain/shared"
)

// Repository defines the interface for animal persistence operations with IoC pattern
type Repository interface {
	// FindOneAndUpsert finds an animal by ID and applies callback for atomic upsert
	FindOneAndUpsert(ctx context.Context, id AnimalID, callback func(*Animal) (*Animal, error)) error

	// FindOneAndInsert inserts a new animal with callback for initialization
	FindOneAndInsert(ctx context.Context, id AnimalID, callback func() (*Animal, error)) error

	// FindOneAndUpdate finds an animal by ID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id AnimalID, callback func(*Animal) (*Animal, error)) error

	// GetByID retrieves an animal by ID (read-only)
	GetByID(ctx context.Context, id AnimalID) (*Animal, error)

	// GetByPosition retrieves wild animals at a specific position (read-only)
	GetByPosition(ctx context.Context, position shared.Position) ([]*Animal, error)

	// GetByOwner retrieves all animals owned by a trainer (read-only)
	GetByOwner(ctx context.Context, ownerID shared.ID) ([]*Animal, error)

	// GetWildAnimalsNearby retrieves wild animals within radius from position (read-only)
	GetWildAnimalsNearby(ctx context.Context, center shared.Position, radius float64) ([]*Animal, error)

	// Delete removes an animal
	Delete(ctx context.Context, id AnimalID) error
}
