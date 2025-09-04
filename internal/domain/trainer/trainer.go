package trainer

import (
	"math/rand"

	"github.com/danghamo/life/internal/domain/shared"
)

// UserID represents a unique user identifier from Account domain
type UserID shared.ID

// String returns string representation
func (id UserID) String() string {
	return string(id)
}

// ValidateNickname validates a nickname string
func ValidateNickname(value string) error {
	if len(value) < 3 || len(value) > 20 {
		return shared.NewDomainError(shared.ErrCodeInvalidNickname, "Nickname must be between 3 and 20 characters")
	}
	return nil
}

// ItemID represents a unique item identifier
type ItemID shared.ID

// NewItemID creates a new item ID
func NewItemID() ItemID {
	return ItemID(shared.NewID())
}

// String returns string representation
func (id ItemID) String() string {
	return string(id)
}

// ItemType represents different types of items
type ItemType string

const (
	// Consumables
	HealthPotion ItemType = "health_potion"
	ManaPotion   ItemType = "mana_potion"

	// Capture tools
	BasicNet    ItemType = "basic_net"
	AdvancedNet ItemType = "advanced_net"
	MasterNet   ItemType = "master_net"

	// Materials
	AnimalHide   ItemType = "animal_hide"
	RareGem      ItemType = "rare_gem"
	MagicCrystal ItemType = "magic_crystal"
)

// String returns string representation
func (it ItemType) String() string {
	return string(it)
}

// IsValid checks if item type is valid
func (it ItemType) IsValid() bool {
	validTypes := []ItemType{
		HealthPotion, ManaPotion, BasicNet, AdvancedNet, MasterNet,
		AnimalHide, RareGem, MagicCrystal,
	}
	for _, validType := range validTypes {
		if it == validType {
			return true
		}
	}
	return false
}

// Item represents an individual item instance
type Item struct {
	ID        ItemID           `json:"id"`
	Type      ItemType         `json:"type"`
	Name      string           `json:"name"`
	CreatedAt shared.Timestamp `json:"created_at"`
}

// NewItem creates a new item
func NewItem(itemType ItemType, name string) (*Item, error) {
	if !itemType.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidItemType, "Invalid item type")
	}

	if len(name) < 1 || len(name) > 50 {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Item name must be between 1 and 50 characters")
	}

	return &Item{
		ID:        NewItemID(),
		Type:      itemType,
		Name:      name,
		CreatedAt: shared.NewTimestamp(),
	}, nil
}

// Inventory represents trainer's inventory with unique items
type Inventory struct {
	Items    map[string]*Item `json:"items"` // itemID -> Item
	MaxSlots int              `json:"max_slots"`
}

// NewInventory creates a new inventory
func NewInventory(maxSlots int) Inventory {
	return Inventory{
		Items:    make(map[string]*Item),
		MaxSlots: maxSlots,
	}
}

// AddItem adds an item to inventory
func (inv *Inventory) AddItem(item *Item) error {
	if item == nil {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Item cannot be nil")
	}

	currentItems := len(inv.Items)
	if currentItems >= inv.MaxSlots {
		return shared.NewDomainError(shared.ErrCodeInventoryFull, "Inventory is full")
	}

	inv.Items[item.ID.String()] = item
	return nil
}

// RemoveItem removes an item from inventory by ID
func (inv *Inventory) RemoveItem(itemID ItemID) (*Item, error) {
	item, exists := inv.Items[itemID.String()]
	if !exists {
		return nil, shared.NewDomainError(shared.ErrCodeItemNotFound, "Item not found in inventory")
	}

	delete(inv.Items, itemID.String())
	return item, nil
}

// HasItem checks if inventory has a specific item
func (inv *Inventory) HasItem(itemID ItemID) bool {
	_, exists := inv.Items[itemID.String()]
	return exists
}

// GetItem returns an item by ID
func (inv *Inventory) GetItem(itemID ItemID) (*Item, bool) {
	item, exists := inv.Items[itemID.String()]
	return item, exists
}

// GetAllItems returns all items in inventory
func (inv *Inventory) GetAllItems() []*Item {
	items := make([]*Item, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, item)
	}
	return items
}

// GetItemsByType returns all items of a specific type
func (inv *Inventory) GetItemsByType(itemType ItemType) []*Item {
	items := make([]*Item, 0)
	for _, item := range inv.Items {
		if item.Type == itemType {
			items = append(items, item)
		}
	}
	return items
}

// CountItemsByType returns the count of items of a specific type
func (inv *Inventory) CountItemsByType(itemType ItemType) int {
	count := 0
	for _, item := range inv.Items {
		if item.Type == itemType {
			count++
		}
	}
	return count
}

// IsFull checks if inventory is full
func (inv *Inventory) IsFull() bool {
	return len(inv.Items) >= inv.MaxSlots
}

// GetUsedSlots returns the number of used slots
func (inv *Inventory) GetUsedSlots() int {
	return len(inv.Items)
}

// AnimalParty represents trainer's animal party
type AnimalParty struct {
	animalIDs []shared.ID
	maxSize   int
}

// NewAnimalParty creates a new animal party
func NewAnimalParty(maxSize int) AnimalParty {
	return AnimalParty{
		animalIDs: make([]shared.ID, 0, maxSize),
		maxSize:   maxSize,
	}
}

// AddAnimal adds an animal to the party
func (party *AnimalParty) AddAnimal(animalID shared.ID) error {
	if len(party.animalIDs) >= party.maxSize {
		return shared.NewDomainError(shared.ErrCodePartyFull, "Animal party is full")
	}

	// Check if animal already exists
	for _, id := range party.animalIDs {
		if id == animalID {
			return shared.NewDomainError(shared.ErrCodeAnimalAlreadyInParty, "Animal is already in party")
		}
	}

	party.animalIDs = append(party.animalIDs, animalID)
	return nil
}

// RemoveAnimal removes an animal from the party
func (party *AnimalParty) RemoveAnimal(animalID shared.ID) error {
	for i, id := range party.animalIDs {
		if id == animalID {
			party.animalIDs = append(party.animalIDs[:i], party.animalIDs[i+1:]...)
			return nil
		}
	}
	return shared.NewDomainError(shared.ErrCodeAnimalNotInParty, "Animal is not in party")
}

// GetAnimals returns all animal IDs in the party
func (party *AnimalParty) GetAnimals() []shared.ID {
	result := make([]shared.ID, len(party.animalIDs))
	copy(result, party.animalIDs)
	return result
}

// IsFull checks if party is full
func (party *AnimalParty) IsFull() bool {
	return len(party.animalIDs) >= party.maxSize
}

// Size returns current party size
func (party *AnimalParty) Size() int {
	return len(party.animalIDs)
}

// Trainer represents a trainer aggregate
type Trainer struct {
	ID         UserID            `json:"id"` // UserID from Account domain
	Nickname   string            `json:"nickname"`
	Color      string            `json:"color"`
	Level      shared.Level      `json:"level"`
	Experience shared.Experience `json:"experience"`
	Stats      shared.Stats      `json:"stats"`
	Position   shared.Position   `json:"position"`
	Movement   MovementState     `json:"movement"`
	Money      shared.Money      `json:"money"`
	Inventory  Inventory         `json:"inventory"`
	Party      AnimalParty       `json:"party"`
	CreatedAt  shared.Timestamp  `json:"created_at"`
	UpdatedAt  shared.Timestamp  `json:"updated_at"`
}

// generateRandomColor generates a random hex color from predefined palette
func generateRandomColor() string {
	// 50 pre-defined attractive colors for better visual distinction
	colors := []string{
		"#4444ff", "#ff4444", "#44ff44", "#ffaa44", "#ff44aa",
		"#44aaff", "#aaff44", "#aa44ff", "#ffaa88", "#88aaff",
		"#aaffaa", "#ffaabb", "#ff8844", "#44ff88", "#8844ff",
		"#ff4488", "#88ff44", "#4488ff", "#aa8844", "#44aa88",
		"#ff6644", "#6644ff", "#44ff66", "#ff44cc", "#cc44ff",
		"#44ffcc", "#ffcc44", "#cc44aa", "#44ccff", "#aaccff",
		"#ffaacc", "#ccffaa", "#aaffcc", "#ffccaa", "#ccaaff",
		"#ff7755", "#7755ff", "#55ff77", "#ff5599", "#9955ff",
		"#55ff99", "#ff9955", "#9955aa", "#55aaff", "#aa55ff",
		"#ff8866", "#8866ff", "#66ff88", "#ff6699", "#9966ff",
		"#66ff99", "#ff9966", "#9966aa", "#66aaff", "#aa66ff",
	}
	return colors[rand.Intn(len(colors))]
}

// NewTrainer creates a new trainer with UserID from Account domain
func NewTrainer(userID UserID, nickname string) (*Trainer, error) {
	// Validate nickname
	if err := ValidateNickname(nickname); err != nil {
		return nil, err
	}
	level, _ := shared.NewLevel(1)
	experience, _ := shared.NewExperience(0, 0)
	stats := shared.NewStats(100, 10, 5, 10, 10) // Starting stats
	position := shared.NewPosition(15.0, 10.0)   // Center of 30x20 map
	movement := NewMovementState()               // Initialize movement state
	movement.StartPos = position                 // Set initial position
	money, _ := shared.NewMoney(1000)            // Starting money
	inventory := NewInventory(50)                // 50 inventory slots
	party := NewAnimalParty(6)                   // Max 6 animals
	timestamp := shared.NewTimestamp()
	color := generateRandomColor() // Assign random color

	trainer := &Trainer{
		ID:         userID, // Use UserID from Account domain
		Nickname:   nickname,
		Color:      color,
		Level:      level,
		Experience: experience,
		Stats:      stats,
		Position:   position,
		Movement:   movement,
		Money:      money,
		Inventory:  inventory,
		Party:      party,
		CreatedAt:  timestamp,
		UpdatedAt:  timestamp,
	}

	return trainer, nil
}

// StartMovement starts movement in given direction
func (t *Trainer) StartMovement(dirX, dirY float64) error {
	// Update current position before starting new movement
	t.UpdatePositionFromMovement()

	direction := MovementDirection{X: dirX, Y: dirY}
	t.Movement.StartMovement(direction, t.Position)
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// StopMovement stops current movement
func (t *Trainer) StopMovement() error {
	// Update to final position
	t.UpdatePositionFromMovement()

	t.Movement.StopMovement(t.Position)
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// UpdatePositionFromMovement updates position based on movement state
func (t *Trainer) UpdatePositionFromMovement() {
	t.Position = t.Movement.CalculateCurrentPosition()
}

// MoveTo moves the trainer to a new position (legacy support)
func (t *Trainer) MoveTo(newPosition shared.Position) error {
	// Stop any current movement and set position directly
	t.Movement.StopMovement(newPosition)
	t.Position = newPosition
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// GainExperience adds experience points
func (t *Trainer) GainExperience(points int) error {
	if points <= 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidExp, "Experience points must be positive")
	}

	t.Experience = t.Experience.Add(points)
	t.UpdatedAt = shared.NewTimestamp()

	// Check for level up
	requiredExp := t.calculateRequiredExperience()
	if t.Experience.CanLevelUp(requiredExp) {
		return t.levelUp(requiredExp)
	}

	return nil
}

// levelUp levels up the trainer
func (t *Trainer) levelUp(requiredExp int) error {
	if !t.Level.CanLevelUp() {
		return shared.NewDomainError(shared.ErrCodeMaxLevel, "Trainer is already at max level")
	}

	// Consume experience
	newExp, err := t.Experience.ConsumeForLevelUp(requiredExp)
	if err != nil {
		return err
	}

	// Level up
	newLevel, err := t.Level.LevelUp()
	if err != nil {
		return err
	}

	// Update stats (simple stat growth)
	statBonus := shared.NewStats(20, 3, 2, 1, 1)

	t.Experience = newExp
	t.Level = newLevel
	t.Stats = t.Stats.Add(statBonus)
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// calculateRequiredExperience calculates experience needed for next level
func (t *Trainer) calculateRequiredExperience() int {
	// Simple exponential formula: level * level * 100
	return t.Level.Value() * t.Level.Value() * 100
}

// SpendMoney spends money
func (t *Trainer) SpendMoney(amount int) error {
	newMoney, err := t.Money.Add(-amount)
	if err != nil {
		return err
	}

	t.Money = newMoney
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// EarnMoney earns money
func (t *Trainer) EarnMoney(amount int) error {
	if amount <= 0 {
		return shared.NewDomainError(shared.ErrCodeInvalidAmount, "Amount must be positive")
	}

	newMoney, err := t.Money.Add(amount)
	if err != nil {
		return err
	}

	t.Money = newMoney
	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// AddAnimalToParty adds an animal to the party
func (t *Trainer) AddAnimalToParty(animalID shared.ID) error {
	err := t.Party.AddAnimal(animalID)
	if err != nil {
		return err
	}

	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// RemoveAnimalFromParty removes an animal from the party
func (t *Trainer) RemoveAnimalFromParty(animalID shared.ID) error {
	err := t.Party.RemoveAnimal(animalID)
	if err != nil {
		return err
	}

	t.UpdatedAt = shared.NewTimestamp()

	return nil
}

// CanAfford checks if trainer can afford something
func (t *Trainer) CanAfford(cost int) bool {
	return t.Money.CanAfford(cost)
}
