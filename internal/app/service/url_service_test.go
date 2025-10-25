package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikiasyonas/url-shortener/internal/app/service"
	"github.com/mikiasyonas/url-shortener/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

type MockURLService struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, url *domain.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	args := m.Called(ctx, originalURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

type MockShortCodeGenerator struct {
	mock.Mock
}

func (m *MockShortCodeGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShortCodeGenerator) Validate(code string) bool {
	args := m.Called(code)
	return args.Bool(0)
}

func TestURLService_ShortenURL_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	mockRepo.On("FindByOriginalURL", ctx, "https://example.com").Return((*domain.URL)(nil), domain.ErrURLNotFound)
	mockGenerator.On("Generate").Return("abc123")
	mockRepo.On("Exists", ctx, "abc123").Return(false, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.URL")).Return(nil)

	result, err := service.ShortenURL(ctx, "https://example.com")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://example.com", result.OriginalURL)
	assert.Equal(t, "abc123", result.ShortCode)

	mockRepo.AssertExpectations(t)
	mockGenerator.AssertExpectations(t)
}

func TestURLService_ShortenURL_ExistingURL(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	existingURL := &domain.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "existing",
	}

	mockRepo.On("FindByOriginalURL", ctx, "https://example.com").Return(existingURL, nil)

	result, err := service.ShortenURL(ctx, "https://example.com")

	assert.NoError(t, err)
	assert.Equal(t, existingURL, result)

	mockRepo.AssertExpectations(t)
	mockGenerator.AssertExpectations(t)
}

func TestURLService_ShortenURL_InvalidURL(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"://example.com",
	}

	for _, invalidURL := range invalidURLs {
		t.Run(invalidURL, func(t *testing.T) {
			result, err := service.ShortenURL(ctx, invalidURL)
			assert.ErrorIs(t, err, domain.ErrInvalidURL)
			assert.Nil(t, result)
		})
	}
}

func TestURLService_Redirect_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	expectedURL := &domain.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
	}

	incrementCalled := make(chan bool, 1)
	mockGenerator.On("Validate", "abc123").Return(true)
	mockRepo.On("FindByShortCode", ctx, "abc123").Return(expectedURL, nil)
	mockRepo.On("IncrementClickCount", mock.Anything, "abc123").Return(nil).Run(func(args mock.Arguments) {
		incrementCalled <- true
	})

	originalURL, err := service.Redirect(ctx, "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", originalURL)

	select {
	case <-incrementCalled:
		// Success - goroutine completed
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for IncrementClickCount to be called")
	}

	mockRepo.AssertExpectations(t)
	mockGenerator.AssertExpectations(t)
}

func TestURLService_Redirect_InvalidShortCode(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	mockGenerator.On("Validate", "invalid").Return(false)

	originalURL, err := service.Redirect(ctx, "invalid")

	assert.ErrorIs(t, err, domain.ErrInvalidShortCode)
	assert.Equal(t, "", originalURL)

	mockRepo.AssertExpectations(t)
	mockGenerator.AssertExpectations(t)
}

func TestURLService_Redirect_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepository)
	mockGenerator := new(MockShortCodeGenerator)

	service := service.NewURLService(mockRepo, mockGenerator)

	mockGenerator.On("Validate", "notfound").Return(true)
	mockRepo.On("FindByShortCode", ctx, "notfound").Return((*domain.URL)(nil), domain.ErrURLNotFound)

	originalURL, err := service.Redirect(ctx, "notfound")

	assert.ErrorIs(t, err, domain.ErrURLNotFound)
	assert.Equal(t, "", originalURL)

	mockRepo.AssertExpectations(t)
	mockGenerator.AssertExpectations(t)
}
