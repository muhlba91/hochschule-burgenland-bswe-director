package state

// State represents the state of the Pokémon game, including the players and their information.
type State struct {
	// SessionID is the unique identifier for the game session.
	SessionID string `json:"session_id" redis:"session_id"`
	// Players is a map of player identifiers to their corresponding Player information.
	Players map[string]*Player `json:"players,omitempty" redis:"players,omitempty"`
}
