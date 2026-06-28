package event

// Type represents the type of an event in the global flow.
type Type string

//nolint:gochecknoglobals // Define the event types for the global flow.
var (
	ConnectionInformation Type = "connection:information"
	Disconnect            Type = "connection:disconnect"
	Disconnected          Type = "connection:disconnected"
)
