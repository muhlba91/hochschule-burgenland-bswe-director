package model

// Session represents a Pokémon game session.
type Session struct {
	// ID is the unique identifier for the session.
	ID string `json:"id" redis:"id"`
	// PlayerA is the first player in the game.
	PlayerA Player `json:"playerA" redis:"player_a"`
	// PlayerB is the second player in the game.
	PlayerB Player `json:"playerB" redis:"player_b"`
	// StartedAt is the timestamp when the game session started.
	StartedAt int64 `json:"startedAt" redis:"started_at"`
	// NextTurn indicates which player's turn is next in the game.
	NextTurn *string `json:"nextTurn,omitempty" redis:"next_turn,omitempty"`
	// NextRequests indicates which request IDDs are expected next in the game flow.
	NextRequests []string `json:"nextRequests,omitempty" redis:"next_requests,omitempty"`
	// Winner indicates which player has won the game.
	Winner *string `json:"winner,omitempty" redis:"winner,omitempty"`
}
