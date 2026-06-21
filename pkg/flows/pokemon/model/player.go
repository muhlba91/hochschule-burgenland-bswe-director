package model

// Player represents a player in the Pokémon game.
type Player struct {
	// Name is the name of the player.
	Name string `json:"name" redis:"name"`
	// URL is the URL associated with the player's service.
	URL string `json:"url" redis:"url"`
	// Connected indicates whether the player is currently connected to the game session.
	Connected bool `json:"connected" redis:"connected"`
}
