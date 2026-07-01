package action

// Start represents the initiation of a new Pokémon game.
type Start struct {
	RequestBase
}

// Started represents a player who has started a Pokémon game.
type Started struct {
	ResponseBaseWithState
}
