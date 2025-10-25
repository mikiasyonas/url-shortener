package main

import (
	"context"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikiasyonas/url-shortener/internal/adapters/cache/redis"
	"github.com/mikiasyonas/url-shortener/internal/adapters/http"
	"github.com/mikiasyonas/url-shortener/internal/adapters/repository/gorm"
	"github.com/mikiasyonas/url-shortener/internal/app/service"
	"github.com/mikiasyonas/url-shortener/internal/core/ports"
	"github.com/mikiasyonas/url-shortener/pkg/config"
	"github.com/mikiasyonas/url-shortener/pkg/database"
	"github.com/mikiasyonas/url-shortener/pkg/monitoring"
	"github.com/mikiasyonas/url-shortener/pkg/shortcode"

	"github.com/joho/godotenv"
)

func main() {
	metrics := monitoring.NewMetrics()
	logger := monitoring.NewLogger(monitoring.INFO)
	healthChecker := monitoring.NewHealthChecker()

	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "development" {
		if err := godotenv.Load(); err != nil {
			logger.Error("No .env file found, using environment variables")
		}
	}

	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration:", err)
	}

	var redisCache ports.Cache
	if cfg.Redis.URL != "" {
		cache, err := redis.NewRedisCache(
			cfg.Redis.URL,
			cfg.Redis.Password,
			cfg.Redis.DB,
			cfg.Redis.TTL,
		)
		if err != nil {
			logger.Info("Redis cache disabled: %v", err)
		} else {
			redisCache = cache
			defer cache.Close()
			logger.Info("Redis cache connected")
		}
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database:", err)
	}

	if err := database.OptimizeConnectionPool(db,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	); err != nil {
		logger.Info("Failed to optimize connection pool: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		logger.Error("Failed to run migrations:", err)
	}

	healthChecker.RegisterCheck("database", monitoring.DatabaseHealthCheck(db), true)
	if redisCache != nil {
		healthChecker.RegisterCheck("redis", monitoring.RedisHealthCheck(redisCache), false)
	}

	urlRepo := gorm.NewURLRepository(db)
	codeGenerator := shortcode.NewGenerator(6)

	baseURLService := service.NewURLService(urlRepo, codeGenerator)

	var urlService ports.URLService = baseURLService
	if redisCache != nil {
		urlService = service.NewCachedURLService(baseURLService, redisCache, urlRepo)
		logger.Info("Cached URL service enabled")
	}

	router := http.NewRouter(urlService, cfg.App.BaseURL, healthChecker, metrics)
	rateLimiter := http.NewRateLimiter(1000, 100)
	router.Use(rateLimiter.Limit)

	monitoring := http.NewMonitoringMiddleware(metrics)
	router.Use(monitoring.Middleware)

	healthHandler := http.NewHealthHandler(healthChecker, metrics)
	router.HandleFunc("/api/health", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/metrics", healthHandler.Metrics).Methods("GET")
	router.HandleFunc("/api/ready", healthHandler.Readiness).Methods("GET")
	router.HandleFunc("/api/live", healthHandler.Liveness).Methods("GET")

	server := &nethttp.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Server listening on port %s", cfg.Server.Port)
		logger.Info("Base URL: %s", cfg.App.BaseURL)
		logger.Info("Environment: %s", cfg.Server.Env)

		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			logger.Error("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}

	logger.Info("Server stopped gracefully")
}
