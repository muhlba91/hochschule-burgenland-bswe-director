package event

// Type represents the type of an event in the Pokémon gameplay.
type Type string

//nolint:gochecknoglobals // Define the event types for the Pokémon gameplay.
var (
	TypeCreate Type = "create"
	TypeList   Type = "list"
	TypeJoin   Type = "join"

	TypeCreated Type = "created"
	TypeListing Type = "listing"
	TypeJoined  Type = "joined"

	TypeNotStarted Type = "not_started"
	TypeStarted    Type = "started"
	TypePaused     Type = "paused"
	TypeResumed    Type = "resumed"
	TypeFinished   Type = "finished"
)
