package gorm

import (
	"context"
	"errors"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"

	"gorm.io/gorm"
)

type URLRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Save(ctx context.Context, url *domain.URL) error {
	result := r.db.WithContext(ctx).Create(url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrShortCodeTaken
		}
		return result.Error
	}
	return nil
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	var url domain.URL
	result := r.db.WithContext(ctx).Where("short_code = ?", shortCode).First(&url)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrURLNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &url, nil
}

func (r *URLRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	var url domain.URL
	result := r.db.WithContext(ctx).Where("original_url = ?", originalURL).First(&url)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrURLNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &url, nil
}

func (r *URLRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&domain.URL{}).Where("short_code = ?", shortCode).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (r *URLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	result := r.db.WithContext(ctx).Model(&domain.URL{}).
		Where("short_code = ?", shortCode).
		Update("click_count", gorm.Expr("click_count + ?", 1))

	return result.Error
}
