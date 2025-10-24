package ports

import (
	"context"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
)

type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (*domain.URL, error)
	Redirect(ctx context.Context, shortCode string) (string, error)
}
