package monitoring

import (
	"fmt"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	requestsTotal    map[string]int64
	requestDurations []time.Duration
	statusCodes      map[int]int64
	activeRequests   int64

	urlsShortened  int64
	urlsRedirected int64
	cacheHits      int64
	cacheMisses    int64

	dbQueriesTotal int64
	dbQueryErrors  int64

	startTime time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestsTotal:    make(map[string]int64),
		statusCodes:      make(map[int]int64),
		requestDurations: make([]time.Duration, 0, 1000),
		startTime:        time.Now(),
	}
}

func (m *Metrics) RecordRequest(method string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsTotal[method]++
	m.statusCodes[statusCode]++

	if len(m.requestDurations) >= 1000 {
		m.requestDurations = m.requestDurations[1:]
	}
	m.requestDurations = append(m.requestDurations, duration)
}

func (m *Metrics) RecordURLShortened() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urlsShortened++
}

func (m *Metrics) RecordURLRedirected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urlsRedirected++
}

func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

func (m *Metrics) RecordDBQuery() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbQueriesTotal++
}

func (m *Metrics) RecordDBError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbQueryErrors++
}

func (m *Metrics) IncrementActiveRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeRequests++
}

func (m *Metrics) DecrementActiveRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeRequests--
}

func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalDuration time.Duration
	for _, d := range m.requestDurations {
		totalDuration += d
	}

	var avgDuration time.Duration
	if len(m.requestDurations) > 0 {
		avgDuration = totalDuration / time.Duration(len(m.requestDurations))
	}

	var cacheHitRate float64
	totalCacheOps := m.cacheHits + m.cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(m.cacheHits) / float64(totalCacheOps) * 100
	}

	var errorRate float64
	totalRequests := int64(0)
	for _, count := range m.requestsTotal {
		totalRequests += count
	}
	if totalRequests > 0 {
		errorRate = float64(m.statusCodes[500]+m.statusCodes[502]+m.statusCodes[503]) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"uptime_seconds":          time.Since(m.startTime).Seconds(),
		"active_requests":         m.activeRequests,
		"requests_total":          m.requestsTotal,
		"status_codes":            m.statusCodes,
		"request_duration_avg_ms": avgDuration.Milliseconds(),
		"requests_per_second":     float64(totalRequests) / time.Since(m.startTime).Seconds(),

		"business": map[string]interface{}{
			"urls_shortened":  m.urlsShortened,
			"urls_redirected": m.urlsRedirected,
			"cache": map[string]interface{}{
				"hits":     m.cacheHits,
				"misses":   m.cacheMisses,
				"hit_rate": fmt.Sprintf("%.2f%%", cacheHitRate),
			},
		},

		"database": map[string]interface{}{
			"queries_total": m.dbQueriesTotal,
			"errors":        m.dbQueryErrors,
			"error_rate":    fmt.Sprintf("%.2f%%", float64(m.dbQueryErrors)/float64(m.dbQueriesTotal)*100),
		},

		"system": map[string]interface{}{
			"error_rate": fmt.Sprintf("%.2f%%", errorRate),
		},
	}
}
