package health

import (
	"encoding/json"
	"net/http"
)

// startupzHandler handles the /startupz endpoint and returns the startup status of the application.
// w: The HTTP response writer.
func (s *Server) startupzHandler(w http.ResponseWriter, _ *http.Request) {
	status := http.StatusOK
	components := map[string]string{
		"sessionStore":   StatusUp,
		"requestStore":   StatusUp,
		"broadcastStore": StatusUp,
	}

	if !s.sessionStore.IsConnected() {
		status = http.StatusServiceUnavailable
		components["sessionStore"] = StatusDown
	}

	if !s.requestStore.IsConnected() {
		status = http.StatusServiceUnavailable
		components["requestStore"] = StatusDown
	}

	if !s.broadcastStore.IsConnected() {
		status = http.StatusServiceUnavailable
		components["broadcastStore"] = StatusDown
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": components,
	})
}
