package data

import (
	"github.com/chrishollman/WotLK-Profession-Leveller/internal/validator"
)

// FilterSkillup - Filters out potential recipes based upon the crafting difficulty, and the callers acceptance of a
// failed skillup. E.g. orange difficulty recipes guarantee a skillup, whereas yellow and green do not.
type FilterSkillup struct {
	target SkillupDifficulty
}

// NewFilterSkillup returns a struct used to filter out recipes based upon the callers easiest desired skillup.
func NewFilterSkillup(input SkillupDifficulty) *FilterSkillup {
	return &FilterSkillup{
		target: input,
	}
}

// Filter returns a bool indicating whether a recipe is of a suitable difficulty.
func (f *FilterSkillup) Filter(r *Recipe, current int) bool {
	// case - cannot learn the recipe yet
	if validator.Max(r.LearnedAt, r.Colors[0]) > current {
		return false
	}

	// all else
	switch f.target {
	// If targetting green, check we're not grey
	case SKILLUP_GREEN:
		return current < r.Colors[3]
	// If targetting yellow, check we're not green
	case SKILLUP_YELLOW:
		return current < r.Colors[2]
	// If targetting orange, check we're not yellow
	case SKILLUP_ORANGE:
		return current < r.Colors[1]
	// n.b. shouldn't hit this case
	default:
		return false
	}
}

// FilterSource - Filters out potential recipes based upon the source from which they are obtained.
type FilterSource struct {
	Allowable []Source `json:"-"`
}

// NewFilterSource returns a struct used to filter out recipes based upon the callers provided suitable sources.
func NewFilterSource(input []Source) *FilterSource {
	return &FilterSource{
		Allowable: input,
	}
}

// Filter returns a bool indicating whether a recipe is from a suitable source.
func (f *FilterSource) Filter(r *Recipe) bool {
	// If we have a value for LearnedAt, it should be either automatically given or from a trainer so just accept it
	if r.LearnedAt != 0 {
		return true
	}

	// Otherwise, filter by possible sources
	for _, v := range r.Source {
		if validator.PermittedValue(v, f.Allowable) {
			return true
		}
	}

	return false
}
