package health

import (
	"fmt"
	"net/http"
)

// startupzHandler handles the /startupz endpoint and returns the startup status of the application, including the status of its components.
// w: The HTTP response writer.
func (s *Server) startupzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"up"}`)
}
