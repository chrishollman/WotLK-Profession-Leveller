package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/chrishollman/WotLK-Profession-Leveller/internal/data"
	"github.com/chrishollman/WotLK-Profession-Leveller/internal/tsm"
	"github.com/chrishollman/WotLK-Profession-Leveller/internal/validator"
)

type plRequestPayload struct {
	Region        string                 `json:"region"`
	Server        string                 `json:"server"`
	Faction       data.Faction           `json:"faction"`
	StartLevel    int                    `json:"start_level"`
	FinishLevel   int                    `json:"finish_level"`
	Profession    data.Profession        `json:"profession"`
	FilterSource  []data.Source          `json:"filter_source"`
	FilterSkillup data.SkillupDifficulty `json:"filter_skillup"`
}

// professionLevellingHandler is the handler for a profession levelling request. It handles various housekeeping aspects
// of the request including calls to ingest and decode the payload, validate it for correctness, and running the main
// levelling function. It then returns the information via JSON.
func (app *application) professionLevellingHandler(w http.ResponseWriter, r *http.Request) {
	validator := validator.New()
	input := &plRequestPayload{}

	err := app.readJSON(w, r, input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.validateProfessionLevellingRequest(validator, input)
	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	res, err := app.levelup(input)
	if err != nil {
		// TODO: better error handling
		app.badRequestResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"data": res}, nil)
}

// levelup is the main function
func (app *application) levelup(input *plRequestPayload) (any, error) {

	totalCost := 0

	// Get the Auction House ID for the players server and faction
	server, err := app.stores.Servers.GetByName(input.Server)
	if err != nil {
		return nil, err
	}

	// Preheat the cache with AH data if necessary
	if err := app.tsmService.Preheat(server.AHIds[input.Faction]); err != nil {
		app.logger.Fatalf("unable to fetch TSM auction house data: %v", err)
		// TODO: Meaningful error return
		return nil, err
	}

	// maintain a player to remember known recipes and inventory items used for future crafts
	player := data.NewPlayer(input.Profession, input.StartLevel, input.FinishLevel, server.AHIds[input.Faction])

	// setup filters
	fSource := data.NewFilterSource(input.FilterSource)
	fSkillup := data.NewFilterSkillup(input.FilterSkillup)

	for player.SkillCurrent < player.SkillDesired {
		app.logger.Debugf("Assessing potential recipes for %v -> %v", player.SkillCurrent, player.SkillCurrent+1)

		// get candidate recipes
		var selectedCraft data.Recipe
		var lowestCost int = math.MaxInt
		candidates := app.stores.Recipes.GetFiltered(player.SkillCurrent, input.Profession, fSource, fSkillup)
		for _, recipe := range candidates {
			var cost int
			numCrafts := app.getRequiredCrafts(player, &recipe)
			price, list, err := app.recipeCost(player, &recipe)

			switch {
			case errors.Is(err, tsm.ErrIsBlacklisted):
				// skip over this candidate because it has items we can't determine a cost for
				continue
			case err != nil:
				app.logger.Errorf("unhandled error getting crafting cost of %v", recipe.ID)
				continue
			default:
				// continue
			}

			cost = numCrafts * price
			if cost < lowestCost {
				lowestCost = cost
				selectedCraft = recipe
			}
			app.logger.Debugf("	%v (%v) costs %v per craft and must (convervatively) be crafted %v times", recipe.ID, recipe.Name, intToGold(price), numCrafts)
			app.logger.Debugf(" Requires purchases of following: %v\n", list)
		}

		if lowestCost == math.MaxInt {
			app.logger.Debugf("Unable to find a suitable crafts for %v -> %v, abandoning...", player.SkillCurrent, player.SkillCurrent+1)
			return nil, nil
		}
		totalCost += lowestCost
		app.logger.Debugf("Based upon the determined costs, %v is the cheapest craft costing %v for level %v", selectedCraft.Name, intToGold(lowestCost), player.SkillCurrent)

		player.SkillCurrent++
	}

	app.logger.Debugf("Total cost going from %v to %v was %v", input.StartLevel, player.SkillCurrent, intToGold(totalCost))

	return nil, nil
}

// validateProfessionLevellingRequests runs various tests against the input received for the profession levelling
// request to ensure the payload is valid.
func (app *application) validateProfessionLevellingRequest(v *validator.Validator, input *plRequestPayload) {
	v.Check(validator.PermittedValue(input.Server, app.getServers()), "server", "must be a valid server")
	v.Check(validator.PermittedValue(input.Region, app.getRegions()), "region", "must be either 'EU' or 'US'")
	v.Check(validator.PermittedValue(int(input.Faction), app.getFactions()), "faction", "must be 'Horde' or 'Alliance'")
	v.Check(input.Profession.String() != data.UNDEFINED_TYPE, "profession", "must be a valid profession")
	v.Check(input.StartLevel >= data.MINIMUM_PROFESSION_LEVEL, "start_level", fmt.Sprintf("must be at least %v", data.MINIMUM_PROFESSION_LEVEL))
	v.Check(input.FinishLevel > input.StartLevel, "finish_level", "must be greater than start_level")
	v.Check(input.FinishLevel <= data.MAXIMUM_PROFESSION_LEVEL, "finish_level", fmt.Sprintf("must be at most %v", data.MAXIMUM_PROFESSION_LEVEL))
}

func (app *application) recipeCost(p *data.Player, r *data.Recipe) (int, [][]int, error) {
	totalCost := 0
	reqPurchases := [][]int{}

	for _, item := range r.Reagents {
		cost, purchases, err := app.reagentCost(p, item)
		if err != nil {
			return 0, nil, fmt.Errorf("couldn't craft item: %w", err)
		}

		totalCost += cost
		reqPurchases = append(reqPurchases, purchases...)
	}
	return totalCost, reqPurchases, nil
}

func (app *application) reagentCost(p *data.Player, reagent []int) (int, [][]int, error) {
	var (
		ahCost    int
		craftCost int

		ahList    [][]int
		craftList [][]int
	)

	id, qty := reagent[0], reagent[1]

	// Get cost to buy from vendor (and assume vendor is always cheapest)
	vendorItem, err := app.stores.VendorItems.GetByID(id)
	if err == nil {
		return vendorItem.Cost * qty, [][]int{{id, qty}}, nil
	}

	// Get cost to craft it
	craftCost, craftList, err = app.craftingCost(id, p)
	if err != nil {
		craftCost = math.MaxInt
		craftList = nil
	}

	// Get cost to buy it from the AH
	tsmItem, err := app.tsmService.GetPrice(p.AuctionHouseID, id)
	if err != nil {
		ahCost = math.MaxInt
		ahList = nil
	} else {
		// TODO: Make this configurable?
		ahCost = tsmItem.MinBuyout
		ahList = append(ahList, []int{id, qty})
	}

	// Determine cheapest route
	switch {
	case ahCost < craftCost:
		return ahCost, ahList, nil
	case craftCost < ahCost:
		return craftCost, craftList, nil
	case craftCost == ahCost && (craftCost == math.MaxInt || ahCost == math.MaxInt):
		return 0, nil, errors.New("couldn't buy or craft this reagent")
	default:
		// TODO: Refactor(?)
		return ahCost, ahList, nil
	}
}

func (app *application) craftingCost(itemID int, p *data.Player) (int, [][]int, error) {
	recipeID, err := app.stores.Items.GetCraftingRecipeID(itemID)
	if err != nil {
		return math.MaxInt, nil, fmt.Errorf("couldn't get crafting cost: %w", err)
	}

	recipe, err := app.stores.Recipes.GetByID(recipeID)

	switch {
	case err != nil:
		return math.MaxInt, nil, fmt.Errorf("couldn't get recipe: %w", err)
	case len(recipe.Profession) <= 0 && recipe.Profession[0] != p.Profession:
		return math.MaxInt, nil, errors.New("can't craft recipe with this profession")
	case strings.Contains(recipe.Name, "Transmute"):
		return math.MaxInt, nil, errors.New("avoid cyclical transmutes")
	}

	cost, items, err := app.recipeCost(p, recipe)
	if err != nil {
		return math.MaxInt, nil, fmt.Errorf("couldn't get recipe cost: %w", err)
	}

	return cost, items, nil
}

func (app *application) getRequiredCrafts(p *data.Player, r *data.Recipe) int {
	// Chance of success
	chance := float64(r.Colors[data.ColorGrey]-p.SkillCurrent) / float64(r.Colors[data.ColorGrey]-r.Colors[data.ColorYellow])

	if chance >= 1 {
		return 1
	}

	return int(math.Ceil(1 / chance))
}
