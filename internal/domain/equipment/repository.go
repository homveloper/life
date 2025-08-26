package equipment

import (
	"context"

	"github.com/danghamo/life/internal/domain/shared"
)

// Repository defines the interface for equipment persistence operations with IoC pattern
type Repository interface {
	// FindOneAndUpsert finds equipment by ID and applies callback for atomic upsert
	FindOneAndUpsert(ctx context.Context, id EquipmentID, callback func(*Equipment) (*Equipment, error)) error

	// FindOneAndInsert inserts new equipment with callback for initialization
	FindOneAndInsert(ctx context.Context, id EquipmentID, callback func() (*Equipment, error)) error

	// FindOneAndUpdate finds equipment by ID and applies callback for atomic update
	FindOneAndUpdate(ctx context.Context, id EquipmentID, callback func(*Equipment) (*Equipment, error)) error

	// GetByID retrieves equipment by ID (read-only)
	GetByID(ctx context.Context, id EquipmentID) (*Equipment, error)

	// GetByOwner retrieves all equipment owned by an animal (read-only)
	GetByOwner(ctx context.Context, ownerID shared.ID) ([]*Equipment, error)

	// GetUnequipped retrieves all unequipped equipment (read-only)
	GetUnequipped(ctx context.Context) ([]*Equipment, error)

	// GetByRarity retrieves equipment by rarity level (read-only)
	GetByRarity(ctx context.Context, rarity Rarity) ([]*Equipment, error)

	// Delete removes equipment
	Delete(ctx context.Context, id EquipmentID) error
}
