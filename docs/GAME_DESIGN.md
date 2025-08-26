# LIFE Game Design Document

## Coordinate System Design

### Position System
- **Coordinate Type**: Real numbers (float64) for free movement
- **Unit Definition**: 1 unit = abstract game unit (not tied to pixels or tiles)
- **Movement**: Continuous movement allowed (e.g., 10.5, 20.7 → 10.6, 20.8)
- **Range**: Flexible based on map size (e.g., 0-1000 x 0-800 units)

```go
type Position struct {
    X, Y float64  // Free-form coordinates
}
```

### Tile System (Separate Concept)
- **Purpose**: Environmental elements, collision detection, terrain types
- **Relation**: Tiles have position and size, independent of player coordinates
- **Flexibility**: Different sized tiles/objects can be placed anywhere

```go
type Tile struct {
    ID       string     // Unique identifier
    Position Position   // Center point of the tile
    Size     Size       // Dimensions of the tile
    Type     TileType   // grass, water, rock, etc.
}

type Size struct {
    Width, Height float64
}
```

### Benefits
1. **Smooth Movement**: No grid constraints, natural physics
2. **Collision Flexibility**: Objects can be any size, anywhere
3. **Scalability**: Easy to add different sized game elements
4. **Precision**: Fine-grained position control for gameplay mechanics

### Implementation Notes
- Server coordinates are authoritative
- Client can interpolate between server updates for smooth visuals
- Collision detection uses tile boundaries, not grid-based
- Movement validation checks against tile properties (e.g., water = impassable)