package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"github.com/mikiasyonas/url-shortener/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func NewConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		Database: getEnv("DB_NAME", "url_shortener"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	log.Printf("Database connected successfully to %s:%s/%s", cfg.Host, cfg.Port, cfg.Name)
	return db, nil
}
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&domain.URL{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}
	log.Println("Database migration completed")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func IsDuplicateKeyError(err error) bool {
	errStr := err.Error()

	patterns := []string{
		"duplicate key value",
		"unique constraint",
		"violates unique constraint",
		"23505",
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return errors.Is(err, gorm.ErrDuplicatedKey)
}
