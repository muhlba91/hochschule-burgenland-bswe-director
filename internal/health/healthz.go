package health

import (
	"encoding/json"
	"net/http"
)

// healthzHandler handles the /healthz endpoint and returns the health status of the application, including the status of its components.
// w: The HTTP response writer.
// r: The HTTP request.
func (s *Server) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	status := http.StatusOK
	components := map[string]string{
		"redis": StatusUp,
	}

	if !s.cache.IsConnected() {
		status = http.StatusServiceUnavailable
		components["redis"] = StatusDown
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": components,
	})
}
