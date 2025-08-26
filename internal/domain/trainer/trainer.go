package trainer

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// UserID represents a unique user identifier from Account domain
type UserID shared.ID

// String returns string representation
func (id UserID) String() string {
	return string(id)
}

// Nickname represents a trainer's nickname
type Nickname struct {
	value string
}

// NewNickname creates a new nickname
func NewNickname(value string) (Nickname, error) {
	if len(value) < 3 || len(value) > 20 {
		return Nickname{}, shared.NewDomainError(shared.ErrCodeInvalidNickname, "Nickname must be between 3 and 20 characters")
	}
	return Nickname{value: value}, nil
}

// Value returns the nickname value
func (n Nickname) Value() string {
	return n.value
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
	Nickname   Nickname          `json:"nickname"`
	Level      shared.Level      `json:"level"`
	Experience shared.Experience `json:"experience"`
	Stats      shared.Stats      `json:"stats"`
	Position   shared.Position   `json:"position"`
	Money      shared.Money      `json:"money"`
	Inventory  Inventory         `json:"inventory"`
	Party      AnimalParty       `json:"party"`
	CreatedAt  shared.Timestamp  `json:"created_at"`
	UpdatedAt  shared.Timestamp  `json:"updated_at"`
}

// NewTrainer creates a new trainer with UserID from Account domain
func NewTrainer(userID UserID, nickname Nickname) (*Trainer, error) {
	level, _ := shared.NewLevel(1)
	experience, _ := shared.NewExperience(0, 0)
	stats := shared.NewStats(100, 10, 5, 10, 10) // Starting stats
	position := shared.NewPosition(15, 10)       // Center of 30x20 map
	money, _ := shared.NewMoney(1000)            // Starting money
	inventory := NewInventory(50)                // 50 inventory slots
	party := NewAnimalParty(6)                   // Max 6 animals
	timestamp := shared.NewTimestamp()

	trainer := &Trainer{
		ID:         userID, // Use UserID from Account domain
		Nickname:   nickname,
		Level:      level,
		Experience: experience,
		Stats:      stats,
		Position:   position,
		Money:      money,
		Inventory:  inventory,
		Party:      party,
		CreatedAt:  timestamp,
		UpdatedAt:  timestamp,
	}

	return trainer, nil
}

// MoveTo moves the trainer to a new position
func (t *Trainer) MoveTo(newPosition shared.Position) error {
	// TODO: Add validation for valid positions when world is implemented
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
