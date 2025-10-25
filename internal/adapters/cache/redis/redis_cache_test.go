package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikiasyonas/url-shortener/internal/adapters/cache/redis"
	"github.com/mikiasyonas/url-shortener/internal/core/domain"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_GetSetURL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := redis.NewRedisCache(mr.Addr(), "", 0, time.Hour)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()
	url := &domain.URL{
		ID:          "test-id",
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
		CreatedAt:   time.Now(),
		ClickCount:  0,
	}

	err = cache.SetURL(ctx, url, 3600)
	assert.NoError(t, err)

	retrieved, err := cache.GetURL(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, url.OriginalURL, retrieved.OriginalURL)
	assert.Equal(t, url.ShortCode, retrieved.ShortCode)
}

func TestRedisCache_GetURL_NotFound(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := redis.NewRedisCache(mr.Addr(), "", 0, time.Hour)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	_, err = cache.GetURL(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrURLNotFound)
}

func TestRedisCache_IncrementClickCount(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := redis.NewRedisCache(mr.Addr(), "", 0, time.Hour)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	err = cache.IncrementClickCount(ctx, "abc123")
	assert.NoError(t, err)

	count, err := cache.GetClickCount(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	err = cache.IncrementClickCount(ctx, "abc123")
	assert.NoError(t, err)

	count, err = cache.GetClickCount(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRedisCache_DeleteURL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := redis.NewRedisCache(mr.Addr(), "", 0, time.Hour)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()
	url := &domain.URL{
		ID:          "test-id",
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
		CreatedAt:   time.Now(),
		ClickCount:  0,
	}

	err = cache.SetURL(ctx, url, 3600)
	assert.NoError(t, err)

	err = cache.DeleteURL(ctx, "abc123")
	assert.NoError(t, err)

	_, err = cache.GetURL(ctx, "abc123")
	assert.ErrorIs(t, err, domain.ErrURLNotFound)
}
