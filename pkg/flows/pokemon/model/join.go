package model

// Join represents the joining of a new Pokémon game.
type Join struct {
	// SessionID is the unique identifier for the game session to join.
	SessionID string `json:"sessionId"`
	// Player is the name of the player joining the game.
	Player string `json:"player"`
}
