package event

import (
	"fmt"
)

// Type represents the type of an event in the global flow.
type Type string

//nolint:gochecknoglobals // Define the event types for the global flow.
var (
	ConnectionInformation Type = "connection:information"
	Disconnect            Type = "connection:disconnect"
	Disconnected          Type = "connection:disconnected"
)

// GenerateEventName generates the event name for a specific event type.
// event: The event type for which to generate the name.
func GenerateEventName(flowName string, event Type) string {
	return fmt.Sprintf("%s:%s", flowName, event)
}
