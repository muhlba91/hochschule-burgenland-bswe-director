package event

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
)

// PlayerInformation represents the information of existing Pokémon games.
type PlayerInformation struct{}

// PlayerDetails represents the details of a specific player in the game.
type PlayerDetails struct {
	// State is the current state of the player in the game.
	State *state.CurrentState `json:"player"`
	// Connection is the connection data for the websocket connection.
	Connection *connection.Data `json:"connection"`
}
