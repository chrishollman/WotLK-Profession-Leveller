package data

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type VendorItem struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Source []Source `json:"source"`
	Cost   int      `json:"cost"`
}

// Holds the data retrieved from static JSON files for Items
type VendorItemStore struct {
	dataslice []VendorItem
	datamap   map[int]VendorItem
	logger    *zap.SugaredLogger
}

// Instantiates the store, loading and parsing the JSON files to make available via the stores methods.
func NewVendorItemStore(logger *zap.SugaredLogger) *VendorItemStore {
	var data []VendorItem

	bytes, err := os.ReadFile("data/vendor.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	datamap := make(map[int]VendorItem, len(data))
	for _, i := range data {
		datamap[i.ID] = i
	}

	return &VendorItemStore{
		dataslice: data,
		datamap:   datamap,
		logger:    logger,
	}
}

// GetByID returns a recipe from its provided ID
func (r *VendorItemStore) GetByID(id int) (*VendorItem, error) {
	if val, ok := r.datamap[id]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("couldn't locate vendor item with id %v", id)
}
