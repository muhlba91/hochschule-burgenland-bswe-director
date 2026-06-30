package pokemon

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	globalEvent "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/event"
)

// Init initializes the gameplay and registers all necessary components.
// reg: The flow registry to register the gameplay.
func (gp *Gameplay) Init(reg *orchestrator.Registry) {
	orchestrator.RegisterSessionGeneratorHandler(reg, gp.generateEventName(event.TypeCreate), gp.Create)
	orchestrator.RegisterSessionGeneratorHandler(reg, gp.generateEventName(event.TypeJoin), gp.Join)

	orchestrator.RegisterEventHandler(reg, gp.generateEventName(event.TypeList), gp.List)
	orchestrator.RegisterEventHandler(
		reg,
		globalEvent.GenerateEventName(constants.Name, globalEvent.Disconnect),
		gp.Disconnect,
	)
	orchestrator.RegisterEventHandler(
		reg,
		globalEvent.GenerateEventName(constants.Name, globalEvent.ConnectionInformation),
		gp.ConnectionInformation,
	)

	orchestrator.RegisterCallbackHandler(reg, gp.generateActionName(action.TypeStart), gp.StartCallback)
}
