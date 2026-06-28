package cache

// Channel represents a channel for communication between different components of the application.
type Channel string

//nolint:gochecknoglobals // Define the channels for the global flow.
var (
	Director Channel = "director"
)
