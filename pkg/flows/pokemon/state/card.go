package state

// Card represents a card in the game.
type Card struct {
	// ID is the unique identifier for the card.
	ID string `json:"id" redis:"id"`
	// Name is the name of the card.
	Name string `json:"name" redis:"name"`
	// Category is the category of the card (e.g., Pokémon, Energy, Trainer).
	Category string `json:"category" redis:"category"`
	// ImageURL is the URL of the card's image.
	ImageURL string `json:"image_url,omitempty" redis:"image_url,omitempty"`
	// ImageData is the binary data of the card's image.
	ImageData []byte `json:"image_data,omitempty" redis:"image_data,omitempty"`
}
