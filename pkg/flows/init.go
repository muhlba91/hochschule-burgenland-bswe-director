package flows

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/requestor"
)

// Init initializes the flows.
// cache: The cache instance for the flows.
// requestor: The requestor instance for the flows.
// registry: The registry to register the flows.
func Init(cache *cache.Cache, requestor *requestor.Requestor, registry *registry.Registry) {
	pokemon.NewGameplay(cache, requestor, registry).Init()
}
