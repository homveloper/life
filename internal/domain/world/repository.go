package world

import (
	"context"

	"github.com/danghamo/life/internal/domain/shared"
)

// Repository defines the interface for world persistence operations with IoC pattern
type Repository interface {
	// FindOneAndUpsert finds world by ID and applies callback for atomic upsert
	FindOneAndUpsert(ctx context.Context, id WorldID, callback func(*World) (*World, error)) error

	// FindOneAndInsert inserts new world with callback for initialization
	FindOneAndInsert(ctx context.Context, id WorldID, callback func() (*World, error)) error

	// FindOneAndUpdate finds world by ID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id WorldID, callback func(*World) (*World, error)) error

	// GetByID retrieves world by ID (read-only)
	GetByID(ctx context.Context, id WorldID) (*World, error)

	// GetTileEntities retrieves all entities at a specific position (read-only)
	GetTileEntities(ctx context.Context, worldID WorldID, position shared.Position) ([]shared.ID, error)

	// Delete removes world
	Delete(ctx context.Context, id WorldID) error
}
