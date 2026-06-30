package state

// Card represents a card in the game.
// FIXME: add picture, etc...
type Card struct {
	// ID is the unique identifier for the card.
	ID string `json:"id" redis:"id"`
	// Name is the name of the card.
	Name string `json:"name" redis:"name"`
	// Type is the type of the card (e.g., Pokémon, Energy, Trainer).
	Type string `json:"type" redis:"type"`
}
