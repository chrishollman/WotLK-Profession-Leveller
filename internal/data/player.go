package data

type Player struct {
	Profession     Profession
	AuctionHouseID int
	SkillCurrent   int
	SkillDesired   int
	Recipes        map[int]bool
	Inventory      map[int]int
}

type InventoryItem struct {
	ID       int
	Quantity int
}

func NewPlayer(profession Profession, skillCurrent, skillDesired, auctionHouseID int) *Player {

	return &Player{
		Profession:     profession,
		AuctionHouseID: auctionHouseID,
		SkillCurrent:   skillCurrent,
		SkillDesired:   skillDesired,
		Recipes:        make(map[int]bool),
		Inventory:      make(map[int]int),
	}
}

// AddRecipe adds a recipe ID to a players list of known recipe's
func (p *Player) AddRecipe(id int) {
	p.Recipes[id] = true
}

// HasRecipe returns whether a player knows a recipe by its provided ID
func (p *Player) HasRecipe(id int) bool {
	return p.Recipes[id]
}

// CheckInventory returns the players quantity of a given item in their inventory
func (p *Player) CheckInventory(id int) int {
	if val, ok := p.Inventory[id]; ok {
		return val
	}
	return 0
}

func (p *Player) AddInventory(id, qty int) {
	val, ok := p.Inventory[id]
	if !ok {
		p.Inventory[id] = val
		return
	}

	p.Inventory[id] = val + qty
}

func (p *Player) ConsumeInventory(id, qty int) bool {
	val, ok := p.Inventory[id]
	if !ok {
		return false
	}

	if val < qty {
		return false
	}

	p.Inventory[id] = val - qty
	return true
}
