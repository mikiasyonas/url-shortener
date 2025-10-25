package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"gorm.io/gorm"
)

type HealthChecker struct {
	checks []HealthCheck
}

type HealthCheck struct {
	Name     string
	Checker  func(ctx context.Context) (bool, error)
	Critical bool
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Critical  bool      `json:"critical"`
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make([]HealthCheck, 0),
	}
}

func (h *HealthChecker) RegisterCheck(name string, checker func(ctx context.Context) (bool, error), critical bool) {
	h.checks = append(h.checks, HealthCheck{
		Name:     name,
		Checker:  checker,
		Critical: critical,
	})
}

func (h *HealthChecker) Check(ctx context.Context) *HealthStatus {
	results := make(map[string]CheckResult)
	overallStatus := "healthy"

	for _, check := range h.checks {
		status := "healthy"
		var errorMsg string

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		healthy, err := check.Checker(checkCtx)
		cancel()

		if err != nil {
			status = "unhealthy"
			errorMsg = err.Error()
			if check.Critical {
				overallStatus = "unhealthy"
			}
		} else if !healthy {
			status = "degraded"
			if check.Critical {
				overallStatus = "degraded"
			}
		}

		results[check.Name] = CheckResult{
			Status:    status,
			Error:     errorMsg,
			Timestamp: time.Now(),
			Critical:  check.Critical,
		}
	}

	return &HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
	}
}

func DatabaseHealthCheck(db *gorm.DB) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		sqlDB, err := db.DB()
		if err != nil {
			return false, fmt.Errorf("failed to get database instance: %w", err)
		}

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(pingCtx); err != nil {
			return false, fmt.Errorf("database ping failed: %w", err)
		}

		stats := sqlDB.Stats()
		if stats.OpenConnections >= stats.MaxOpenConnections {
			return false, fmt.Errorf("database connection pool exhausted: %d/%d", stats.OpenConnections, stats.MaxOpenConnections)
		}

		return true, nil
	}
}

func RedisHealthCheck(cache interface{}) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		if redisCache, ok := cache.(interface{ GetClient() interface{} }); ok {
			client := redisCache.GetClient()
			if pingable, ok := client.(interface{ Ping(context.Context) error }); ok {
				pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				if err := pingable.Ping(pingCtx); err != nil {
					return false, fmt.Errorf("redis ping failed: %w", err)
				}
				return true, nil
			}
		}

		if testableCache, ok := cache.(interface {
			GetURL(ctx context.Context, shortCode string) (*domain.URL, error)
		}); ok {
			testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			// Try to get a non-existent key to test connectivity
			_, err := testableCache.GetURL(testCtx, "health-check-test")
			if err != nil && err != domain.ErrURLNotFound {
				return false, fmt.Errorf("redis operation failed: %w", err)
			}
			return true, nil
		}

		return true, nil
	}
}
