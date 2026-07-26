package state

// Trainer represents a trainer card in the game.
type Trainer struct {
	// Card is the card information of the trainer card.
	Card

	// Effect is the effect of the trainer card.
	Effect string `json:"effect" redis:"effect"`
}
