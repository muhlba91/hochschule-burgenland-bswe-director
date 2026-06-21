package model

// Create represents the creation of a new Pokémon game.
type Create struct {
	// PlayerA is the first player in the game.
	PlayerA Player `json:"playerA"`
	// PlayerB is the second player in the game.
	PlayerB Player `json:"playerB"`
}

// Created represents a newly created Pokémon game.
type Created struct {
	// SessionID is the unique identifier for the created game session.
	SessionID string `json:"sessionId"`
}
