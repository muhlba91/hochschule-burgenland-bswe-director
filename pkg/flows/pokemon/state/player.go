package state

// Player represents the state of a player in the game.
type Player struct {
	// Active is the information about the currently active Pokémon.
	Active *Pokemon `json:"active" redis:"active"`
	// Bench is the list of Pokémon on the player's bench.
	Bench []*Pokemon `json:"bench" redis:"bench"`
	// Hand is the list of cards in the player's hand.
	Hand []*Card `json:"hand" redis:"hand"`
	// Deck is the list of cards in the player's deck.
	Deck []*Card `json:"deck" redis:"deck"`
	// DiscardPile is the list of cards in the player's discard pile.
	DiscardPile []*Card `json:"discard_pile" redis:"discard_pile"`
	// PrizeCards is the list of prize cards the player has.
	PrizeCards []*Card `json:"prize_cards" redis:"prize_cards"`
}
