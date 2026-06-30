package health

import (
	"fmt"
	"net/http"
)

// livezHandler handles the /livez endpoint and returns the liveness status of the application.
// w: The HTTP response writer.
func (s *Server) livezHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"%s"}`, StatusUp)
}
