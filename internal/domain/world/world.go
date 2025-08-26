package world

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// WorldID represents a unique world identifier
type WorldID shared.ID

// NewWorldID creates a new world ID
func NewWorldID() WorldID {
	return WorldID(shared.NewID())
}

// String returns string representation
func (id WorldID) String() string {
	return string(id)
}

// TerrainType represents different terrain types
type TerrainType string

const (
	Grassland TerrainType = "grassland" // Open savanna
	Forest    TerrainType = "forest"    // Dense forest areas
	Water     TerrainType = "water"     // Rivers, lakes
	Mountain  TerrainType = "mountain"  // Rocky areas
)

// String returns string representation
func (tt TerrainType) String() string {
	return string(tt)
}

// IsValid checks if terrain type is valid
func (tt TerrainType) IsValid() bool {
	return tt == Grassland || tt == Forest || tt == Water || tt == Mountain
}

// IsWalkable checks if terrain is walkable
func (tt TerrainType) IsWalkable() bool {
	return tt != Water && tt != Mountain
}

// GetMovementCost returns movement cost for terrain
func (tt TerrainType) GetMovementCost() float64 {
	switch tt {
	case Grassland:
		return 1.0 // Normal movement
	case Forest:
		return 1.5 // Slower movement
	case Water, Mountain:
		return -1.0 // Impassable
	default:
		return 1.0
	}
}

// Tile represents a single tile in the world
type Tile struct {
	Position shared.Position `json:"position"`
	Terrain  TerrainType     `json:"terrain"`
	Entities []shared.ID     `json:"entities"` // IDs of entities on this tile (trainers, animals)
}

// NewTile creates a new tile
func NewTile(position shared.Position, terrain TerrainType) Tile {
	return Tile{
		Position: position,
		Terrain:  terrain,
		Entities: make([]shared.ID, 0),
	}
}

// AddEntity adds an entity to this tile
func (t *Tile) AddEntity(entityID shared.ID) error {
	// Check if entity already exists
	for _, id := range t.Entities {
		if id == entityID {
			return shared.NewDomainError(shared.ErrCodeEntityAlreadyOnTile, "Entity is already on this tile")
		}
	}

	t.Entities = append(t.Entities, entityID)
	return nil
}

// RemoveEntity removes an entity from this tile
func (t *Tile) RemoveEntity(entityID shared.ID) error {
	for i, id := range t.Entities {
		if id == entityID {
			t.Entities = append(t.Entities[:i], t.Entities[i+1:]...)
			return nil
		}
	}
	return shared.NewDomainError(shared.ErrCodeEntityNotOnTile, "Entity is not on this tile")
}

// HasEntity checks if entity is on this tile
func (t *Tile) HasEntity(entityID shared.ID) bool {
	for _, id := range t.Entities {
		if id == entityID {
			return true
		}
	}
	return false
}

// IsWalkable checks if tile is walkable
func (t *Tile) IsWalkable() bool {
	return t.Terrain.IsWalkable()
}

// World represents the game world aggregate
type World struct {
	ID        WorldID          `json:"id"`
	Name      string           `json:"name"`
	Width     int              `json:"width"`
	Height    int              `json:"height"`
	Tiles     map[string]*Tile `json:"tiles"` // position key -> tile
	CreatedAt shared.Timestamp `json:"created_at"`
	UpdatedAt shared.Timestamp `json:"updated_at"`
}

// NewWorld creates a new world
func NewWorld(name string, width, height int) (*World, error) {
	if len(name) < 1 || len(name) > 50 {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "World name must be between 1 and 50 characters")
	}

	if width < 10 || width > 1000 {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidWorldSize, "World width must be between 10 and 1000")
	}

	if height < 10 || height > 1000 {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidWorldSize, "World height must be between 10 and 1000")
	}

	id := NewWorldID()
	timestamp := shared.NewTimestamp()

	world := &World{
		ID:        id,
		Name:      name,
		Width:     width,
		Height:    height,
		Tiles:     make(map[string]*Tile),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	// Generate default world (all grassland for MVP)
	world.generateDefaultTerrain()

	return world, nil
}

// generateDefaultTerrain generates default terrain for the world
func (w *World) generateDefaultTerrain() {
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			position := shared.NewPosition(x, y)

			// Simple terrain generation for MVP
			var terrain TerrainType
			if x == 0 || x == w.Width-1 || y == 0 || y == w.Height-1 {
				terrain = Mountain // Border mountains
			} else if (x+y)%7 == 0 {
				terrain = Forest // Scattered forests
			} else if (x+y)%11 == 0 {
				terrain = Water // Scattered water
			} else {
				terrain = Grassland // Default grassland
			}

			tile := NewTile(position, terrain)
			w.Tiles[position.Key()] = &tile
		}
	}
}

// Dimensions returns world width and height
func (w *World) Dimensions() (int, int) {
	return w.Width, w.Height
}

// GetTile returns tile at position
func (w *World) GetTile(position shared.Position) (*Tile, error) {
	if !w.IsValidPosition(position) {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidPosition, "Position is outside world boundaries")
	}

	tile, exists := w.Tiles[position.Key()]
	if !exists {
		return nil, shared.NewDomainError(shared.ErrCodeTileNotFound, "Tile not found")
	}

	return tile, nil
}

// IsValidPosition checks if position is within world boundaries
func (w *World) IsValidPosition(position shared.Position) bool {
	return position.X >= 0 && position.X < w.Width && position.Y >= 0 && position.Y < w.Height
}

// IsWalkablePosition checks if position is walkable
func (w *World) IsWalkablePosition(position shared.Position) bool {
	tile, err := w.GetTile(position)
	if err != nil {
		return false
	}
	return tile.IsWalkable()
}

// MoveEntity moves an entity from one position to another
func (w *World) MoveEntity(entityID shared.ID, fromPos, toPos shared.Position) error {
	if !w.IsWalkablePosition(toPos) {
		return shared.NewDomainError(shared.ErrCodeInvalidMove, "Destination is not walkable")
	}

	// Remove from old tile (if valid position)
	if w.IsValidPosition(fromPos) {
		fromTile, err := w.GetTile(fromPos)
		if err == nil {
			fromTile.RemoveEntity(entityID) // Ignore error if not found
		}
	}

	// Add to new tile
	toTile, err := w.GetTile(toPos)
	if err != nil {
		return err
	}

	err = toTile.AddEntity(entityID)
	if err != nil {
		return err
	}

	w.UpdatedAt = shared.NewTimestamp()

	return nil
}

// GetEntitiesAt returns all entities at a position
func (w *World) GetEntitiesAt(position shared.Position) ([]shared.ID, error) {
	tile, err := w.GetTile(position)
	if err != nil {
		return nil, err
	}
	result := make([]shared.ID, len(tile.Entities))
	copy(result, tile.Entities)
	return result, nil
}

// GetNeighbors returns walkable neighboring positions
func (w *World) GetNeighbors(position shared.Position) []shared.Position {
	neighbors := make([]shared.Position, 0, 8)

	// Check all 8 directions (including diagonals)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue // Skip current position
			}

			neighborPos := shared.NewPosition(position.X+dx, position.Y+dy)
			if w.IsWalkablePosition(neighborPos) {
				neighbors = append(neighbors, neighborPos)
			}
		}
	}

	return neighbors
}
