package callback

import (
	"fmt"
	"strings"
)

// Request represents a request in the system.
type Request struct {
	// ID is the unique identifier for the request.
	ID string `json:"id" redis:"id"`
	// SessionID is the unique identifier for the session associated with the request.
	SessionID string `json:"session_id" redis:"session_id"`
	// InternalID is an internal identifier for the request, used for tracking and logging purposes.
	InternalID string `json:"internal_id" redis:"internal_id"`
	// Endpoint is the endpoint to which the request is directed.
	Endpoint string `json:"endpoint" redis:"endpoint"`
	// Secret is a secret generated for the request to verify callback authenticity.
	Secret string `json:"secret" redis:"secret"`
	// Body contains the body of the request.
	Body string `json:"data" redis:"data"`
	// Action is the action associated with the request.
	Action string `json:"action" redis:"action"`
	// Parallelization indicates the level of parallelization for the request. 0 means unlimited.
	Parallelization int `json:"parallelization" redis:"parallelization"`
	// CreatedAt is the timestamp when the request was created.
	CreatedAt int64 `json:"created_at" redis:"created_at"`
	// Completed indicates whether the request has been completed.
	Completed bool `json:"completed" redis:"completed"`
}

// GetRequestPrefix generates a request prefix based on the flow name.
// flowName: The name of the flow for which the request prefix is generated.
func GetRequestPrefix(flowName string) string {
	return fmt.Sprintf("request:%s", flowName)
}

// GetRequestID generates a unique request ID based on the flow name and a unique identifier.
// flowName: The name of the flow for which the request is created.
// id: A unique identifier for the request.
func GetRequestID(flowName string, id string) string {
	return fmt.Sprintf("%s:%s", GetRequestPrefix(flowName), id)
}

// GetFlowName returns the name of the flow associated with the request ID.
// requestID: The request ID in the format "request:<flow_name>:<id>".
func GetFlowName(requestID string) string {
	parts := strings.Split(requestID, ":")
	return parts[1]
}
