package action

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// RequestBase represents the base structure for Pokémon game requests, containing common fields shared by different request types.
type RequestBase struct {
	callback.RequestDataBase

	// State is the current state of the player in the game.
	State *state.State `json:"state,omitempty"`
}

// ResponseBaseWithState represents the base structure for Pokémon game responses that include the player's state, containing common fields shared by different response types.
type ResponseBaseWithState struct {
	// State is the current state of the player in the game.
	State *state.Player `json:"state"`
}
