package domain

import (
	"time"
)

type URL struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalURL string    `json:"original_url" gorm:"not null;type:text"`
	ShortCode   string    `json:"short_code" gorm:"not null;uniqueIndex;size:10"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:now()"`
	ClickCount  int64     `json:"click_count" gorm:"not null;default:0"`
}

func NewURL(originalURL, shortCode string) (*URL, error) {
	if originalURL == "" {
		return nil, ErrInvalidURL
	}
	if shortCode == "" {
		return nil, ErrInvalidShortCode
	}

	return &URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		ClickCount:  0,
	}, nil
}

func (u *URL) IncrementClickCount() {
	u.ClickCount++
}
