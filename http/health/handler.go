package health

import (
	"net/http"
	"time"
)

// LivenessHandler returns an HTTP handler for liveness probes.
// Liveness probes check if the application is running.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}
}

// ReadinessHandler returns an HTTP handler for readiness probes.
// Readiness probes check if the application is ready to serve traffic.
func (m *Manager) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		results := m.CheckAll(ctx)

		// Check if all checks passed
		allHealthy := true
		for _, err := range results {
			if err != nil {
				allHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if allHealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not_ready","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		}
	}
}

// HealthHandler returns a detailed health check handler.
// This provides more detailed information about each health check.
func (m *Manager) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		results := m.CheckAll(ctx)

		// Build response
		response := `{"status":"`
		allHealthy := true
		checks := `,"checks":{`
		first := true

		for name, err := range results {
			if !first {
				checks += ","
			}
			first = false

			if err != nil {
				allHealthy = false
				checks += `"` + name + `":{"status":"unhealthy","error":"` + err.Error() + `"}`
			} else {
				checks += `"` + name + `":{"status":"healthy"}`
			}
		}

		checks += `}`

		if allHealthy {
			response += `healthy"`
		} else {
			response += `unhealthy"`
		}

		response += `,"timestamp":"` + time.Now().Format(time.RFC3339) + `"`
		response += checks
		response += `}`

		w.Header().Set("Content-Type", "application/json")

		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Write([]byte(response))
	}
}
