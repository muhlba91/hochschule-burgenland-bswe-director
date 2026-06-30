package pokemon

import (
	"fmt"

	globalCache "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	pokemonCache "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	pokemonRequestor "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/requestor"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
	globalRequestor "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/requestor"
	globalEvent "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/event"
)

// Gameplay represents a gameplay type.
type Gameplay struct {
	Cache     *pokemonCache.Wrapper
	Registry  *registry.Registry
	Requestor *pokemonRequestor.Wrapper
}

// NewGameplay creates a new instance of the Pokémon gameplay.
// cache: The cache instance for the gameplay.
// registry: The flow registry to register the gameplay.
func NewGameplay(
	cache *globalCache.Cache,
	requestor *globalRequestor.Requestor,
	registry *registry.Registry,
) *Gameplay {
	cacheWrapper := pokemonCache.NewWrapper(cache)
	requestorWrapper := pokemonRequestor.NewWrapper(requestor, cacheWrapper)

	return &Gameplay{
		Cache:     cacheWrapper,
		Registry:  registry,
		Requestor: requestorWrapper,
	}
}

// Init initializes the gameplay and registers all necessary components.
func (gp *Gameplay) Init() {
	registry.RegisterSessionGeneratorHandler(gp.Registry, gp.generateEventName(event.Create), gp.Create)
	registry.RegisterSessionGeneratorHandler(gp.Registry, gp.generateEventName(event.Join), gp.Join)

	registry.RegisterEventHandler(gp.Registry, gp.generateEventName(event.List), gp.List)
	registry.RegisterEventHandler(gp.Registry, string(globalEvent.Disconnect), gp.Disconnect)
	registry.RegisterEventHandler(gp.Registry, string(globalEvent.ConnectionInformation), gp.ConnectionInformation)
}

// generateEventName generates the event name for a specific event type.
// event: The event type for which to generate the name.
func (gp *Gameplay) generateEventName(event event.Type) string {
	return fmt.Sprintf("%s:%s", constants.Name, event)
}

// generateActionName generates the action name for a specific action type.
// action: The action type for which to generate the name.
func (gp *Gameplay) generateActionName(action action.Type) string {
	return fmt.Sprintf("%s:%s", constants.Name, action)
}
