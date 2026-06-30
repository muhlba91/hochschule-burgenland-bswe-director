package state

// State represents the state of the Pokémon game, including the players and their information.
type State struct {
	// SessionID is the unique identifier for the game session.
	SessionID string `json:"session_id" redis:"session_id"`
	// PlayerAState is the state of player A.
	PlayerAState *Player `json:"player_a_state" redis:"player_a_state"`
	// PlayerBState is the state of player B.
	PlayerBState *Player `json:"player_b_state" redis:"player_b_state"`
}
