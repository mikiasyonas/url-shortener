package http

import (
	"github.com/mikiasyonas/url-shortener/internal/core/ports"
	"github.com/mikiasyonas/url-shortener/pkg/monitoring"

	"github.com/gorilla/mux"
)

func NewRouter(urlService ports.URLService, baseUrl string, healthChecker *monitoring.HealthChecker, metrics *monitoring.Metrics) *mux.Router {
	router := mux.NewRouter()
	handlers := NewHandlers(urlService, baseUrl)

	healthHandler := NewHealthHandler(healthChecker, metrics)
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/metrics", healthHandler.Metrics).Methods("GET")
	router.HandleFunc("/ready", healthHandler.Readiness).Methods("GET")
	router.HandleFunc("/live", healthHandler.Liveness).Methods("GET")

	api := router.PathPrefix("/api").Subrouter()
	router.HandleFunc("/{code}", handlers.Redirect).Methods("GET")
	api.HandleFunc("/shorten", handlers.ShortenURL).Methods("POST")

	return router
}
