package state

// Attack represents an attack action taken by a player during their turn in a Pokémon game.
type Attack struct {
	// Name is the name of the attack used by the player.
	Name string `json:"name"`
	// Damage is the amount of damage dealt by the attack. -1 indicates the attack does not deal damage (e.g., status moves).
	Damage int `json:"damage"`
}
