package data

import (
	"encoding/json"
	"errors"
	"os"

	"go.uber.org/zap"
)

var (
	ErrorItemNotExist     = errors.New("no entry for item")
	ErrorItemNotCraftable = errors.New("item is not craftable")
)

// Contains information about physical items that you can possess, and related information such as whether it itself is
// craftable with a profession, whether it's a recipe used in crafting, etc.
type Item struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	CraftedBy    []int    `json:"craftedby,omitempty"`    // [professionID, recipeSpellID]
	TeachesCraft []int    `json:"teachescraft,omitempty"` // [professionID, recipeSpellID]
	Source       []Source `json:"source"`
}

// Holds the data retrieved from static JSON files for Items
type ItemStore struct {
	dataslice []Item
	datamap   map[int]Item
	logger    *zap.SugaredLogger
}

// Instantiates the store, loading and parsing the JSON files to make available via the stores methods.
func NewItemStore(logger *zap.SugaredLogger) *ItemStore {
	var data []Item

	bytes, err := os.ReadFile("data/items.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	datamap := make(map[int]Item, len(data))
	for _, i := range data {
		datamap[i.ID] = i
	}

	return &ItemStore{
		dataslice: data,
		datamap:   datamap,
		logger:    logger,
	}
}

func (i *ItemStore) GetCraftingRecipeID(itemID int) (int, error) {
	item, ok := i.datamap[itemID]
	if !ok {
		return 0, ErrorItemNotExist
	}

	if len(item.CraftedBy) != 2 {
		return 0, ErrorItemNotCraftable
	}

	return item.CraftedBy[1], nil
}
