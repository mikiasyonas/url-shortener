package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"github.com/mikiasyonas/url-shortener/internal/core/ports"
)

type urlService struct {
	repo          ports.URLRepository
	codeGenerator ports.ShortCodeGenerator
}

func NewURLService(repo ports.URLRepository, codeGenerator ports.ShortCodeGenerator) *urlService {
	return &urlService{
		repo:          repo,
		codeGenerator: codeGenerator,
	}
}

func (s *urlService) ShortenURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	if err := s.validateURL(originalURL); err != nil {
		return nil, err
	}

	if existing, err := s.repo.FindByOriginalURL(ctx, originalURL); err == nil {
		return existing, nil
	}

	shortCode, err := s.generateUniqueShortCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique short code: %w", err)
	}

	newURL, err := domain.NewURL(originalURL, shortCode)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, newURL); err != nil {
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	return newURL, nil
}

func (s *urlService) Redirect(ctx context.Context, shortCode string) (string, error) {
	if !s.codeGenerator.Validate(shortCode) {
		return "", domain.ErrInvalidShortCode
	}

	url, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	go func() {
		backgroundCtx := context.Background()
		if err := s.repo.IncrementClickCount(backgroundCtx, shortCode); err != nil {
			fmt.Printf("Failed to increment click count: %v\n", err)
		}
	}()

	return url.OriginalURL, nil
}

func (s *urlService) generateUniqueShortCode(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for i := 0; i < maxAttempts; i++ {
		shortCode := s.codeGenerator.Generate()

		exists, err := s.repo.Exists(ctx, shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

func (s *urlService) validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.ErrInvalidURL
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ErrInvalidURL
	}

	return nil
}
