package action

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// Attack represents an attack action taken by a player during their turn in a Pokémon game.
type Attack struct {
	RequestBase

	// State is the current state of the player in the game.
	Attack *state.Attack `json:"attack"`
}

// Attacked represents a player who has been attacked during their turn in a Pokémon game.
type Attacked struct {
	ResponseBaseWithState
}
