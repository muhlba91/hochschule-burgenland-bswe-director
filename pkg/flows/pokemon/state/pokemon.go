package state

// Pokemon represents a Pokémon in the game.
type Pokemon struct {
	// Card is the card information of the Pokémon.
	Card

	// Types is the list of types of the Pokémon (e.g., Fire, Water, Grass).
	Types []string `json:"types,omitempty" redis:"types,omitempty"`
	// Weaknesses is the list of weaknesses of the Pokémon.
	Weaknesses []*PokemonWeakness `json:"weaknesses,omitempty" redis:"weaknesses,omitempty"`
	// Description is the description of the Pokémon.
	Description string `json:"description,omitempty" redis:"description,omitempty"`
	// Stage is the stage of the Pokémon (e.g., Basic, Stage 1, Stage 2).
	Stage string `json:"stage,omitempty" redis:"stage,omitempty"`
	// HP is the current health points of the Pokémon.
	HP int `json:"hp,omitempty" redis:"hp,omitempty"`
	// MaxHP is the maximum health points of the Pokémon.
	MaxHP int `json:"max_hp" redis:"max_hp"`
	// EnergyAttached is the list of energy cards attached to the Pokémon.
	EnergyAttached []*Card `json:"energy_attached,omitempty" redis:"energy_attached,omitempty"`
	// RetreatCost is the cost to retreat the Pokémon.
	RetreatCost int `json:"retreat_cost,omitempty" redis:"retreat_cost,omitempty"`
	// Attacks is the list of attacks of the Pokémon.
	Attacks []*PokemonAttack `json:"attacks,omitempty" redis:"attacks,omitempty"`
}

// PokemonWeakness represents a weakness of a Pokémon.
type PokemonWeakness struct {
	// Type is the type of the weakness (e.g., Fire, Water, Grass).
	Type string `json:"type" redis:"type"`
	// Value is the value of the weakness (e.g., x2, x3).
	Value string `json:"value" redis:"value"`
}

// PokemonAttack represents an attack of a Pokémon.
type PokemonAttack struct {
	// Name is the name of the attack.
	Name string `json:"name" redis:"name"`
	// Cost is the list of energy types required to perform the attack.
	Cost []string `json:"cost,omitempty" redis:"cost,omitempty"`
	// Effect is the effect of the attack.
	Effect string `json:"effect,omitempty" redis:"effect,omitempty"`
	// Damage is the damage dealt by the attack.
	Damage int `json:"damage,omitempty" redis:"damage,omitempty"`
}
