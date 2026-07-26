package state

// Energy represents an energy card in the game.
type Energy struct {
	// Card is the card information of the energy card.
	Card

	// Element is the type of energy (e.g., Fire, Water, Grass).
	Element string `json:"element" redis:"element"`
}
