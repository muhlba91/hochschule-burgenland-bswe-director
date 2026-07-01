package action

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// Turn represents the action of taking a turn in a Pokémon game.
type Turn struct {
	RequestBase
}

// TurnResult represents the outcome of a player taking a turn in a Pokémon game.
type TurnResult struct {
	ResponseBaseWithState

	// Attack is the attack action taken by the player during their turn.
	Attack *state.Attack `json:"attack,omitempty"`
}
