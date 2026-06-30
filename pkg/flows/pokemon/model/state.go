package model

// State represents the state of the Pokémon game, including the players and their information.
type State struct {
	// SessionID is the unique identifier for the game session.
	SessionID string `json:"session_id" redis:"session_id"`
	// PlayerAState is the state of player A.
	PlayerAState *PlayerState `json:"player_a_state" redis:"player_a_state"`
	// PlayerBState is the state of player B.
	PlayerBState *PlayerState `json:"player_b_state" redis:"player_b_state"`
}

// PlayerState represents the state of a player in the game.
type PlayerState struct {
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
