package event

// Type represents the type of an event in the Pokémon gameplay.
type Type string

//nolint:gochecknoglobals // Define the event types for the Pokémon gameplay.
var (
	Create Type = "create"
	List   Type = "list"
	Join   Type = "join"

	Created Type = "created"
	Listing Type = "listing"
)
