package session

// Player represents a player in the Pokémon game.
type Player struct {
	// ID is the unique identifier for the player.
	ID string `json:"id" redis:"id"`
	// Name is the name of the player.
	Name string `json:"name" redis:"name"`
	// URL is the URL associated with the player's service.
	URL string `json:"url" redis:"url"`
	// Connected indicates whether the player is currently connected to the game session.
	Connected bool `json:"connected" redis:"connected"`
	// ConnectionID is the unique identifier for the player's connection.
	ConnectionID *string `json:"connection_id" redis:"connection_id"`
}
