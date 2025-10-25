package http

import (
	"encoding/json"
	"net/http"

	"github.com/mikiasyonas/url-shortener/pkg/monitoring"
)

type HealthHandler struct {
	healthChecker *monitoring.HealthChecker
	metrics       *monitoring.Metrics
}

func NewHealthHandler(healthChecker *monitoring.HealthChecker, metrics *monitoring.Metrics) *HealthHandler {
	return &HealthHandler{
		healthChecker: healthChecker,
		metrics:       metrics,
	}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthStatus := h.healthChecker.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusOK
	switch healthStatus.Status {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthStatus)
}

func (h *HealthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.metrics.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	healthStatus := h.healthChecker.Check(r.Context())

	ready := true
	for _, check := range healthStatus.Checks {
		if check.Critical && check.Status != "healthy" {
			ready = false
			break
		}
	}

	if ready {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}
