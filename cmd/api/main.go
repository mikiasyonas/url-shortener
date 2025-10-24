package main

import (
	"context"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikiasyonas/url-shortener/internal/adapters/http"
	"github.com/mikiasyonas/url-shortener/internal/adapters/repository/gorm"
	"github.com/mikiasyonas/url-shortener/internal/app/service"
	"github.com/mikiasyonas/url-shortener/pkg/config"
	"github.com/mikiasyonas/url-shortener/pkg/database"
	"github.com/mikiasyonas/url-shortener/pkg/shortcode"

	"github.com/joho/godotenv"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "development" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		log.Fatal("❌ Invalid configuration:", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	urlRepo := gorm.NewURLRepository(db)
	codeGenerator := shortcode.NewGenerator(6)
	urlService := service.NewURLService(urlRepo, codeGenerator)

	router := http.NewRouter(urlService, cfg.App.BaseURL)
	server := &nethttp.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("📡 Server listening on port %s", cfg.Server.Port)
		log.Printf("🌐 Base URL: %s", cfg.App.BaseURL)
		log.Printf("🔧 Environment: %s", cfg.Server.Env)

		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatal("❌ Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server stopped gracefully")
}
