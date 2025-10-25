package http

import (
	"net/http"
	"time"

	"github.com/mikiasyonas/url-shortener/pkg/monitoring"

	"github.com/gorilla/mux"
)

type MonitoringMiddleware struct {
	metrics *monitoring.Metrics
}

func NewMonitoringMiddleware(metrics *monitoring.Metrics) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		metrics: metrics,
	}
}

func (m *MonitoringMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.metrics.IncrementActiveRequests()
		defer m.metrics.DecrementActiveRequests()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		route := m.getRouteName(r)
		m.metrics.RecordRequest(r.Method, rw.statusCode, duration)

		switch route {
		case "shorten":
			m.metrics.RecordURLShortened()
		case "redirect":
			m.metrics.RecordURLRedirected()
		}
	})
}

func (m *MonitoringMiddleware) getRouteName(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if name := route.GetName(); name != "" {
			return name
		}
		if path, err := route.GetPathTemplate(); err == nil {
			return path
		}
	}
	return r.URL.Path
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
