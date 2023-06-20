package data

const (
	MINIMUM_PROFESSION_LEVEL int = 1
	MAXIMUM_PROFESSION_LEVEL int = 450

	UNDEFINED_TYPE = "undefined"
)

// Colors used for indexing into color array
const (
	ColorOrange int = iota
	ColorYellow
	ColorGreen
	ColorGrey
)

// Enum representing playable factions, Alliance and Horde.
type Faction int

const (
	FACTION_UNDEFINED Faction = iota
	FACTION_ALLIANCE
	FACTION_HORDE
)

func (f Faction) String() string {
	switch f {
	case FACTION_ALLIANCE:
		return "alliance"
	case FACTION_HORDE:
		return "horde"
	default:
		return UNDEFINED_TYPE
	}
}

// Represents all learnable primary professions (+ cooking). Uses Wowhead numbers as identifiers. Sort of an enum type,
// but sort of not.
type Profession int

const (
	PROFESSION_ALCHEMY        Profession = 171
	PROFESSION_BLACKSMITHING  Profession = 164
	PROFESSION_COOKING        Profession = 185
	PROFESSION_ENCHANTING     Profession = 333
	PROFESSION_ENGINEERING    Profession = 202
	PROFESSION_INSCRIPTION    Profession = 773
	PROFESSION_JEWELCRAFTING  Profession = 755
	PROFESSION_LEATHERWORKING Profession = 165
	PROFESSION_TAILORING      Profession = 197
)

func (p Profession) String() string {
	switch p {
	case PROFESSION_ALCHEMY:
		return "alchemy"
	case PROFESSION_BLACKSMITHING:
		return "blacksmithing"
	case PROFESSION_COOKING:
		return "cooking"
	case PROFESSION_ENCHANTING:
		return "enchanting"
	case PROFESSION_ENGINEERING:
		return "engineering"
	case PROFESSION_INSCRIPTION:
		return "inscription"
	case PROFESSION_JEWELCRAFTING:
		return "jewelcrafting"
	case PROFESSION_LEATHERWORKING:
		return "leatherworking"
	case PROFESSION_TAILORING:
		return "tailoring"
	default:
		return "undefined"
	}
}

// Represents the lowest allowable difficulty when crafting. That is to say, if you're only willing to craft orange
// level recipes for guaranteed skillups, you can enforce that. If you want to allow green recipes and are somewhat
// willing to accept the indeterminate nature of crafting vs skillups, you can do that too.
//
// N.B. The system is built such that it determines the probability of a crafting success for yellow and beyond, and
// allocates appropriate resources for that.
type SkillupDifficulty int

const (
	SKILLUP_UNDEFINED SkillupDifficulty = iota
	SKILLUP_ORANGE
	SKILLUP_YELLOW
	SKILLUP_GREEN
)

func (s SkillupDifficulty) String() string {
	switch s {
	case SKILLUP_ORANGE:
		return "orange"
	case SKILLUP_YELLOW:
		return "yellow"
	case SKILLUP_GREEN:
		return "green"
	default:
		return UNDEFINED_TYPE
	}
}

// Represents how an item/spell/recipe comes to be in the possession of the character. It uses the same numbers as
// Wowhead in an enum-like fashion however there are a number of undetermined numbers that we skip over.
type Source int

const (
	SOURCE_UNDEFINED Source = iota
	SOURCE_CRAFTED
	SOURCE_DROP
	SOURCE_PVP
	SOURCE_QUEST
	SOURCE_VENDOR
	SOURCE_TRAINER
	SOURCE_DISCOVERY
	_
	_
	_
	_
	SOURCE_ACHIEVEMENT
	_
	_
	SOURCE_DISENCHANTED
	SOURCE_FISHED
	SOURCE_GATHERED
	_
	SOURCE_MINED
	_
	SOURCE_PICKPOCKETED
	_
	SOURCE_SKINNED
	_
	_
	_
	_
	_
	_
	SOURCE_TRANSMUTE_DISCOVERY // 30 (Wowhead doesn't assign this a value, so we set one far away)
)

func (s Source) String() string {
	switch s {
	case SOURCE_CRAFTED:
		return "crafted"
	case SOURCE_DROP:
		return "drop"
	case SOURCE_PVP:
		return "pvp"
	case SOURCE_QUEST:
		return "quest"
	case SOURCE_VENDOR:
		return "vendor"
	case SOURCE_TRAINER:
		return "trainer"
	case SOURCE_ACHIEVEMENT:
		return "achievement"
	case SOURCE_DISENCHANTED:
		return "disenchanted"
	case SOURCE_FISHED:
		return "fished"
	case SOURCE_GATHERED:
		return "gathered"
	case SOURCE_MINED:
		return "mined"
	case SOURCE_PICKPOCKETED:
		return "pickpocketed"
	case SOURCE_SKINNED:
		return "skinned"
	case SOURCE_TRANSMUTE_DISCOVERY:
		return "transmute_discovery"
	default:
		return "undefined"
	}
}

type PurchaseOrder struct {
	Cost  int `json:"cost"`
	Items []struct {
		ID       int `json:"id"`
		Quantity int `json:"quantity"`
	} `json:"items"`
}
