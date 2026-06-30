package model

// List represents the listing of existing Pokémon games.
type List struct{}

// Listing represents a list of existing Pokémon games.
type Listing struct {
	// Sessions is a list of existing Pokémon game sessions.
	Sessions map[string]*Session `json:"sessions"`
}
