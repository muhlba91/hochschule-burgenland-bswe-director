package event

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// State represents the state of existing Pokémon games.
type State struct{}

// CurrentState represents the current state of existing Pokémon games.
type CurrentState struct {
	State *state.CurrentState `json:"state"`
}
