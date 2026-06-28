package model

// Join represents the joining of a new Pokémon game.
type Join struct {
	// SessionID is the unique identifier for the game session to join.
	SessionID string `json:"sessionId"`
	// Player is the URL of the player joining the game.
	Player string `json:"player"`
}

// Joined represents a player who has joined a Pokémon game.
type Joined struct {
	// SessionID is the unique identifier for the joined game session.
	SessionID string `json:"sessionId"`
}
