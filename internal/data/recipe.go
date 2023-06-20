package data

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/chrishollman/WotLK-Profession-Leveller/internal/validator"
)

// Recipe the structure of this is largely kept intact from how it is scraped from Wowhead
type Recipe struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Source       []Source     `json:"source"`
	LearnedAt    int          `json:"learnedat"`
	Profession   []Profession `json:"skill"`  // 'skill' refers to a profession number
	Colors       []int        `json:"colors"` // order is: orange, yellow, green, grey
	Creates      []int        `json:"creates"`
	Reagents     [][]int      `json:"reagents"`
	TrainingCost int          `json:"trainingcost"`
}

type RecipeStore struct {
	dataslice []Recipe
	datamap   map[int]Recipe
	logger    *zap.SugaredLogger
}

func NewRecipeStore(logger *zap.SugaredLogger) *RecipeStore {
	var data []Recipe

	bytes, err := os.ReadFile("data/recipes.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	datamap := make(map[int]Recipe, len(data))
	for _, r := range data {
		datamap[r.ID] = r
	}

	return &RecipeStore{
		dataslice: data,
		datamap:   datamap,
		logger:    logger,
	}
}

// GetByID returns a recipe from its provided ID
func (r *RecipeStore) GetByID(id int) (*Recipe, error) {
	if val, ok := r.datamap[id]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("couldn't locate recipe with id %v", id)
}

// GetBatchByID returns a slice of recipes from their provided slice of ID's
func (r *RecipeStore) GetBatchByID(ids []int) ([]*Recipe, error) {
	var res []*Recipe

	for _, id := range ids {
		tmp, err := r.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("couldn't locate all recipes with provided ids: %v", ids)
		}
		res = append(res, tmp)
	}

	return res, nil
}

// GetFiltered returns a slice of all recipes relevant at a point in time (i.e. skill level)
func (r *RecipeStore) GetFiltered(
	currentSkill int,
	profession Profession,
	fSource *FilterSource,
	fSkillup *FilterSkillup,
) []Recipe {
	var res []Recipe

	// Iterate through all recipe's in the slice and filter
	for _, recipe := range r.dataslice {
		if recipe.Profession[0] == profession && fSource.Filter(&recipe) && fSkillup.Filter(&recipe, currentSkill) {
			res = append(res, recipe)
		}
	}

	return res
}

// Validate determines whether a provided Recipe contains full and valid information.
// TODO: Use or remove
func (r *Recipe) Validate(v *validator.Validator) {
	v.Check(r.ID > 0, "id", validateRecipeMessage(r, "must be greater than zero"))
	v.Check(r.Name != "", "name", validateRecipeMessage(r, "must have a name"))
	v.Check(r.Source != nil, "source", validateRecipeMessage(r, "must be provided"))
	v.Check(len(r.Source) >= 1, "source", validateRecipeMessage(r, "must have one or more nominated"))
	for _, s := range r.Source {
		v.Check(s.String() != UNDEFINED_TYPE, "source", validateRecipeMessage(r, "must be of a defined type"))
	}

	v.Check(len(r.Profession) == 1, "profession/skill", validateRecipeMessage(r, "must only be associated with one"))
	v.Check(r.Profession[0].String() != UNDEFINED_TYPE, "profession/skill", validateRecipeMessage(r, "must be of a defined type"))

	v.Check(len(r.Colors) == 4, "colors", validateRecipeMessage(r, "must have all four skillup colors defined"))
	v.Check(validator.Max(r.LearnedAt, r.Colors[0]) >= 1, "colors.orange", validateRecipeMessage(r, "must be greater than zero"))
	v.Check(r.Colors[1] > validator.Max(r.LearnedAt, r.Colors[0]), "colors.yellow", validateRecipeMessage(r, "must be greater than orange or learnedat"))
	v.Check(r.Colors[2] > r.Colors[1], "colors.green", validateRecipeMessage(r, "must be greater than yellow"))
	v.Check(r.Colors[3] > r.Colors[2], "colors.grey", validateRecipeMessage(r, "must be greater than green"))

	v.Check(len(r.Creates) == 3, "creates", validateRecipeMessage(r, "must have an ID and a quantity"))
	v.Check(r.Creates[0] != 0, "creates.id", validateRecipeMessage(r, "must feature a valid ID"))
	v.Check(r.Creates[1] > 0, "crafts.quantityMin", validateRecipeMessage(r, "must create at least one"))
	v.Check(r.Creates[2] >= r.Creates[1], "crafts.quantityMax", validateRecipeMessage(r, "must be equal to, or greater than, the minimum quantity"))

	v.Check(r.Reagents != nil, "reagents", validateRecipeMessage(r, "must be provided"))
	v.Check(len(r.Reagents) >= 1, "reagents", validateRecipeMessage(r, "must have one or more nominated"))
}

func validateRecipeMessage(r *Recipe, msg string) string {
	return fmt.Sprintf("recipe with id %v %v", r.ID, msg)
}
