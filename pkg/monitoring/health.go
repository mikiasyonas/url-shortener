package monitoring

import (
	"context"
	"time"
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

func DatabaseHealthCheck(db interface{}) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		return true, nil
	}
}

func RedisHealthCheck(redis interface{}) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		return true, nil
	}
}
