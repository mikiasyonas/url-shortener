package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"github.com/mikiasyonas/url-shortener/internal/core/ports"
)

type Handlers struct {
	urlService ports.URLService
	baseUrl    string
}

func NewHandlers(urlService ports.URLService, baseUrl string) *Handlers {
	return &Handlers{
		urlService: urlService,
		baseUrl:    baseUrl,
	}
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		h.respondError(w, http.StatusBadRequest, "URL is required")
		return
	}

	url, err := h.urlService.ShortenURL(r.Context(), req.URL)
	if err != nil {
		switch err {
		case domain.ErrInvalidURL:
			h.respondError(w, http.StatusBadRequest, "Invalid URL")
		default:
			h.respondError(w, http.StatusInternalServerError, "Failed to shorten URL")
		}
		return
	}

	response := ShortenResponse{
		ShortURL:    "http://" + r.Host + "/" + url.ShortCode,
		OriginalURL: url.OriginalURL,
		ShortCode:   url.ShortCode,
	}

	h.respondJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["code"]
	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "Short code is required")
		return
	}

	originalURL, err := h.urlService.Redirect(r.Context(), shortCode)
	if err != nil {
		switch err {
		case domain.ErrURLNotFound:
			http.NotFound(w, r)
		case domain.ErrInvalidShortCode:
			h.respondError(w, http.StatusBadRequest, "Invalid short code")
		default:
			h.respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	log.Println("Original", originalURL)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    "URL Shortener Service is healthy",
	})
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, response JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, JSONResponse{
		Success: false,
		Error:   message,
	})
}
