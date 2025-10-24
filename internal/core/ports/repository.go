package ports

import (
	"context"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
)

type URLRepository interface {
	Save(ctx context.Context, url *domain.URL) error
	FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error)
	FindByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error)
	Exists(ctx context.Context, shortCode string) (bool, error)
	IncrementClickCount(ctx context.Context, shortCode string) error
}

type ShortCodeGenerator interface {
	Generate() string
	Validate(code string) bool
}
