package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
	PoolSize int
	TTL      time.Duration
}

type ServerConfig struct {
	Port               string
	Env                string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	CORSAllowedOrigins []string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AppConfig struct {
	BaseURL            string
	ShortCodeLength    int
	MaxURLLength       int
	RateLimitPerSecond int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:               getEnv("SERVER_PORT", "8080"),
			Env:                getEnv("ENVIRONMENT", "development"),
			ReadTimeout:        getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:       getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:        getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:    getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}, ","),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "url_shortener"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		App: AppConfig{
			BaseURL:            getEnv("APP_BASE_URL", "http://localhost:8080"),
			ShortCodeLength:    getEnvAsInt("APP_SHORT_CODE_LENGTH", 6),
			MaxURLLength:       getEnvAsInt("APP_MAX_URL_LENGTH", 2048),
			RateLimitPerSecond: getEnvAsInt("APP_RATE_LIMIT_PER_SECOND", 100),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 100),
			TTL:      getEnvAsDuration("REDIS_TTL", 24*time.Hour),
		},
	}
}

func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.App.BaseURL == "" {
		return fmt.Errorf("APP_BASE_URL is required")
	}
	if c.App.ShortCodeLength < 4 || c.App.ShortCodeLength > 10 {
		return fmt.Errorf("APP_SHORT_CODE_LENGTH must be between 4 and 10")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration for %s, using default: %v", key, defaultValue)
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, sep string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, sep)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean for %s, using default: %t", key, defaultValue)
	}
	return defaultValue
}
