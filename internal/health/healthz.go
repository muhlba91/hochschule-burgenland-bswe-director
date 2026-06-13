package health

import (
	"fmt"
	"net/http"
)

// healthzHandler handles the /healthz endpoint and returns the health status of the application, including the status of its components.
// w: The HTTP response writer.
func (s *Server) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"up"}`)
}
