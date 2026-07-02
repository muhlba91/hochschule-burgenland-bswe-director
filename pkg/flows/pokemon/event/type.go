package event

// Type represents the type of an event in the Pokémon gameplay.
type Type string

//nolint:gochecknoglobals // Define the event types for the Pokémon gameplay.
var (
	TypeCreate            Type = "create"
	TypeList              Type = "list"
	TypeJoin              Type = "join"
	TypeState             Type = "state"
	TypePlayerInformation Type = "player_information"

	TypeCreated       Type = "created"
	TypeListing       Type = "listing"
	TypeJoined        Type = "joined"
	TypeCurrentState  Type = "current_state"
	TypePlayerDetails Type = "player_details"

	TypeNotStarted Type = "not_started"
	TypeStarted    Type = "started"
	TypePaused     Type = "paused"
	TypeResumed    Type = "resumed"
	TypeFinished   Type = "finished"
)
