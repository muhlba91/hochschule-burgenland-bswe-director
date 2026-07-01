package action

// Type represents the type of an action in the Pokémon gameplay.
type Type string

//nolint:gochecknoglobals // Define the action types for the Pokémon gameplay.
var (
	TypeStart  Type = "start"
	TypeTurn   Type = "turn"
	TypeAttack Type = "attack"
)
