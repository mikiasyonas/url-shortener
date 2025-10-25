package service

import (
	"context"
	"fmt"
	"log"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"github.com/mikiasyonas/url-shortener/internal/core/ports"
)

type cachedURLService struct {
	urlService ports.URLService
	cache      ports.Cache
	repo       ports.URLRepository
}

func NewCachedURLService(urlService ports.URLService, cache ports.Cache, repo ports.URLRepository) *cachedURLService {
	log.Println("Initialized Cached URL Service")
	return &cachedURLService{
		urlService: urlService,
		cache:      cache,
		repo:       repo,
	}
}

func (s *cachedURLService) ShortenURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	return s.urlService.ShortenURL(ctx, originalURL)
}

func (s *cachedURLService) Redirect(ctx context.Context, shortCode string) (string, error) {
	if s.cache != nil {
		if url, err := s.cache.GetURL(ctx, shortCode); err == nil {
			go s.incrementClickCount(shortCode)
			return url.OriginalURL, nil
		}
	}

	originalURL, err := s.urlService.Redirect(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if s.cache != nil {
		go s.cacheURL(context.Background(), shortCode)
	}

	return originalURL, nil
}

func (s *cachedURLService) cacheURL(ctx context.Context, shortCode string) {
	url, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		fmt.Printf("Failed to get URL for caching: %v\n", err)
		return
	}

	if err := s.cache.SetURL(ctx, url, 3600); err != nil {
		fmt.Printf("Failed to cache URL: %v\n", err)
	}
}

func (s *cachedURLService) incrementClickCount(shortCode string) {
	ctx := context.Background()
	if s.cache != nil {
		s.cache.IncrementClickCount(ctx, shortCode)
	}
}
