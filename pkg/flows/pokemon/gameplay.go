package pokemon

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
)

// Gameplay represents a gameplay type.
type Gameplay struct {
	Name     string
	Cache    *cache.Cache
	Registry *registry.Registry
}

// NewGameplay creates a new instance of the Pokémon gameplay.
// cache: The cache instance for the gameplay.
// registry: The flow registry to register the gameplay.
func NewGameplay(cache *cache.Cache, registry *registry.Registry) *Gameplay {
	return &Gameplay{
		Name:     "pokemon",
		Cache:    cache,
		Registry: registry,
	}
}

// Init initializes the gameplay and registers all necessary components.
func (gp *Gameplay) Init() {
}
