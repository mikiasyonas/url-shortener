package ports

import (
	"context"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
)

type Cache interface {
	GetURL(ctx context.Context, shortCode string) (*domain.URL, error)
	SetURL(ctx context.Context, url *domain.URL, ttl int) error
	DeleteURL(ctx context.Context, shortCode string) error

	IncrementClickCount(ctx context.Context, shortCode string) error
	GetClickCount(ctx context.Context, shortCode string) (int64, error)
}
