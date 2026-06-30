package action

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// Start represents the initiation of a new Pokémon game.
type Start struct {
	callback.RequestDataBase

	// SessionID is the unique identifier for the game session to start.
	SessionID string `json:"sessionId"`
	// Callback is the URL of the callback to notify when ready to start the game.
	Callback string `json:"callback"`
}

// Started represents a player who has started a Pokémon game.
type Started struct {
	// State is the current state of the player in the game.
	State state.Player `json:"state"`
}
