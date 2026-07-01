package health

import (
	"net/http"
	"sync"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
)

// checkHealth concurrently checks the health of all components and returns the combined status and component statuses.
func (s *Server) checkHealth() (int, map[string]string) {
	stores := map[string]store.Store{
		"sessionStore":   s.sessionStore,
		"requestStore":   s.requestStore,
		"broadcastStore": s.broadcastStore,
	}

	results := make(chan healthCheckResult, len(stores))
	var wg sync.WaitGroup

	for name, st := range stores {
		wg.Add(1)
		go func(name string, st store.Store) {
			defer wg.Done()
			up := st.IsConnected()
			status := StatusUp
			if !up {
				status = StatusDown
			}
			results <- healthCheckResult{name: name, status: status, up: up}
		}(name, st)
	}

	wg.Wait()
	close(results)

	status := http.StatusOK
	components := make(map[string]string)
	for res := range results {
		components[res.name] = res.status
		if !res.up {
			status = http.StatusServiceUnavailable
		}
	}

	return status, components
}
