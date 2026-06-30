package state

// Pokemon represents a Pokémon in the game.
type Pokemon struct {
	// Card is the card information of the Pokémon.
	Card

	// HP is the current health points of the Pokémon.
	HP int `json:"hp" redis:"hp"`
	// MaxHP is the maximum health points of the Pokémon.
	MaxHP int `json:"max_hp" redis:"max_hp"`
	// EnergyAttached is the list of energy cards attached to the Pokémon.
	EnergyAttached []*Card `json:"energy_attached" redis:"energy_attached"`
}
