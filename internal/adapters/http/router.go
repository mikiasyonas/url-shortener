package http

import (
	"github.com/mikiasyonas/url-shortener/internal/core/ports"

	"github.com/gorilla/mux"
)

func NewRouter(urlService ports.URLService, baseUrl string) *mux.Router {
	router := mux.NewRouter()
	handlers := NewHandlers(urlService, baseUrl)

	api := router.PathPrefix("/api").Subrouter()
	router.HandleFunc("/{code}", handlers.Redirect).Methods("GET")
	api.HandleFunc("/shorten", handlers.ShortenURL).Methods("POST")
	api.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	return router
}
