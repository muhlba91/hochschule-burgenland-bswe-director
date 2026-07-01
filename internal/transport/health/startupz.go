package health

import (
	"encoding/json"
	"net/http"
)

// startupzHandler handles the /startupz endpoint and returns the startup status of the application.
// w: The HTTP response writer.
func (s *Server) startupzHandler(w http.ResponseWriter, _ *http.Request) {
	status, components := s.checkHealth()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": components,
	})
}
