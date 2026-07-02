package state

import "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"

// State represents the state of the Pokémon game, including the players and their information.
type State struct {
	// SessionID is the unique identifier for the game session.
	SessionID string `json:"session_id" redis:"session_id"`
	// Players is a map of player identifiers to their corresponding Player information.
	Players map[string]*Player `json:"players,omitempty" redis:"players,omitempty"`
}

// CurrentState represents a partial state of the Pokémon game, including the players and their information.
type CurrentState struct {
	// Player is the information about the current player in the game.
	Player *Player `json:"player,omitempty" redis:"player,omitempty"`
	// Opponents is a map of opponent identifiers to their corresponding Player information.
	Opponents map[string]*Opponent `json:"opponents,omitempty" redis:"opponents,omitempty"`
	// Session is the session data for the Pokémon game session.
	Session *session.Session `json:"session,omitempty"`
}

// GlobalState represents the global state of the Pokémon game, including all players and their information.
type GlobalState struct {
	// Players is a map of player identifiers to their corresponding Player information.
	Players map[string]*Opponent `json:"players,omitempty" redis:"players,omitempty"`
	// Session is the session data for the Pokémon game session.
	Session *session.Session `json:"session,omitempty"`
}
