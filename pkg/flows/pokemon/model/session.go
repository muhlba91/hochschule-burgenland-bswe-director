package model

// Session represents a Pokémon game session.
type Session struct {
	// PlayerA is the first player in the game.
	PlayerA Player `json:"playerA" redis:"player_a"`
	// PlayerB is the second player in the game.
	PlayerB Player `json:"playerB" redis:"player_b"`
	// NextTurn indicates which player's turn is next in the game.
	NextTurn *string `json:"nextTurn,omitempty" redis:"next_turn,omitempty"`
	// Winner indicates which player has won the game.
	Winner *string `json:"winner,omitempty" redis:"winner,omitempty"`
}
