package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikiasyonas/url-shortener/internal/adapters/http"
	"github.com/mikiasyonas/url-shortener/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) ShortenURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	args := m.Called(ctx, originalURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLService) Redirect(ctx context.Context, shortCode string) (string, error) {
	args := m.Called(ctx, shortCode)
	return args.String(0), args.Error(1)
}

func TestHandlers_ShortenURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	handlers := http.NewHandlers(mockService, "http://localhost:8080")

	expectedURL := &domain.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
	}

	mockService.On("ShortenURL", mock.Anything, "https://example.com").Return(expectedURL, nil)

	reqBody := map[string]string{"url": "https://example.com"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handlers.ShortenURL(rr, req)

	assert.Equal(t, nethttp.StatusCreated, rr.Code)

	var response http.JSONResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be a map")

	assert.True(t, response.Success)
	assert.Equal(t, "https://example.com", dataMap["original_url"])
	assert.Equal(t, "abc123", dataMap["short_code"])

	mockService.AssertExpectations(t)
}

func TestHandlers_ShortenURL_InvalidURL(t *testing.T) {
	mockService := new(MockURLService)
	handlers := http.NewHandlers(mockService, "http://localhost:8080")

	mockService.On("ShortenURL", mock.Anything, "invalid-url").Return((*domain.URL)(nil), domain.ErrInvalidURL)

	reqBody := map[string]string{"url": "invalid-url"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	handlers.ShortenURL(rr, req)

	assert.Equal(t, nethttp.StatusBadRequest, rr.Code)

	var response http.JSONResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid URL", response.Error)

	mockService.AssertExpectations(t)
}

func TestHandlers_Redirect_Success(t *testing.T) {
	mockService := new(MockURLService)
	handlers := http.NewHandlers(mockService, "http://localhost:8080")

	mockService.On("Redirect", mock.Anything, "abc123").Return("https://example.com", nil)

	req := httptest.NewRequest("GET", "/abc123", nil)

	rr := httptest.NewRecorder()
	handlers.Redirect(rr, req)

	assert.Equal(t, nethttp.StatusMovedPermanently, rr.Code)
	assert.Equal(t, "https://example.com", rr.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestHandlers_Redirect_NotFound(t *testing.T) {
	mockService := new(MockURLService)
	handlers := http.NewHandlers(mockService, "http://localhost:8080")

	mockService.On("Redirect", mock.Anything, "notfound").Return("", domain.ErrURLNotFound)

	req := httptest.NewRequest("GET", "/notfound", nil)

	rr := httptest.NewRecorder()
	handlers.Redirect(rr, req)

	assert.Equal(t, nethttp.StatusNotFound, rr.Code)

	mockService.AssertExpectations(t)
}

func TestHandlers_HealthCheck(t *testing.T) {
	mockService := new(MockURLService)
	handlers := http.NewHandlers(mockService, "http://localhost:8080")

	req := httptest.NewRequest("GET", "/api/health", nil)

	rr := httptest.NewRecorder()
	handlers.HealthCheck(rr, req)

	assert.Equal(t, nethttp.StatusOK, rr.Code)

	var response http.JSONResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.True(t, response.Success)
	assert.Equal(t, "URL Shortener Service is healthy", response.Data)
}
