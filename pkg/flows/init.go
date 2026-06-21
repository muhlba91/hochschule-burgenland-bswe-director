package flows

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
)

// Init initializes the flows.
// cache: The cache instance for the flows.
// registry: The registry to register the flows.
func Init(cache *cache.Cache, registry *registry.Registry) {
	pokemon.NewGameplay(cache, registry).Init()
}
