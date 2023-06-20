package data

import "go.uber.org/zap"

type Stores struct {
	Items       *ItemStore
	NexusHub    *NexusHubStore
	Recipes     *RecipeStore
	Servers     *ServerStore
	VendorItems *VendorItemStore
}

func NewStores(logger *zap.SugaredLogger) *Stores {
	return &Stores{
		Items:       NewItemStore(logger),
		NexusHub:    NewNexusHubStore(logger),
		Recipes:     NewRecipeStore(logger),
		Servers:     NewServerStore(logger),
		VendorItems: NewVendorItemStore(logger),
	}
}
