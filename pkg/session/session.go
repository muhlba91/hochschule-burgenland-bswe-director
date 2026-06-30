package session

import (
	"fmt"
	"slices"
	"strings"
)

// Session represents a session in the flow.
type Session interface {
	// GetID returns the unique identifier of the session.
	GetID() string

	// GetNextRequests returns a list of pending request IDs for this session.
	GetNextRequests() []string

	// UpdateNextRequests updates the list of pending request IDs for this session.
	// requestID: The ID of the request to be added or removed from the pending requests.
	UpdateNextRequests(requestID string)

	// DeleteNextRequest removes a request ID from the list of pending requests for this session.
	// requestID: The ID of the request to be removed from the pending requests.
	DeleteNextRequest(requestID string)

	// IsRequestExpected checks if a given request ID is expected in the session's flow.
	// requestID: The ID of the request to check.
	IsRequestExpected(requestID string) bool
}

// Base provides a base implementation of the Session interface.
type Base struct {
	// ID is the unique identifier for the session.
	ID string `json:"id" redis:"id"`
	// NextRequests indicates which request IDDs are expected next in the flow.
	NextRequests []string `json:"nextRequests,omitempty" redis:"next_requests,omitempty"`
}

// GetID returns the unique identifier of the session.
func (s *Base) GetID() string {
	return s.ID
}

// GetNextRequests returns a list of pending request IDs for this session.
func (s *Base) GetNextRequests() []string {
	return s.NextRequests
}

// UpdateNextRequests updates the list of pending request IDs for this session.
// requestID: The ID of the request to be added to the pending requests.
func (s *Base) UpdateNextRequests(requestID string) {
	s.NextRequests = append(s.NextRequests, requestID)
}

// DeleteNextRequest removes a request ID from the list of pending requests for this session.
// requestID: The ID of the request to be removed from the pending requests.
func (s *Base) DeleteNextRequest(requestID string) {
	s.NextRequests = slices.DeleteFunc(s.NextRequests, func(id string) bool {
		return id == requestID
	})
}

// IsRequestExpected checks if a given request ID is expected in the session's flow.
// requestID: The ID of the request to check.
func (s *Base) IsRequestExpected(requestID string) bool {
	for _, reqID := range s.NextRequests {
		if reqID == requestID {
			return true
		}
	}

	return false
}

// GetSessionPrefix generates a session prefix based on the flow name.
// flowName: The name of the flow for which the session prefix is generated.
func GetSessionPrefix(flowName string) string {
	return fmt.Sprintf("session:%s", flowName)
}

// GetSessionID generates a unique session ID based on the flow name and a unique identifier.
// flowName: The name of the flow for which the session is created.
// id: A unique identifier for the session.
func GetSessionID(flowName string, id string) string {
	return fmt.Sprintf("%s:%s", GetSessionPrefix(flowName), id)
}

// GetFlowName returns the name of the flow associated with the session ID.
// sessionID: The session ID in the format "session:<flow_name>:<id>".
func GetFlowName(sessionID string) string {
	parts := strings.Split(sessionID, ":")
	return parts[1]
}
