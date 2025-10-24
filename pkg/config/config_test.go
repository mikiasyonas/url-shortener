package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/mikiasyonas/url-shortener/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := config.Load()

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Env)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 6, cfg.App.ShortCodeLength)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("DB_HOST", "prod-db")
	os.Setenv("APP_SHORT_CODE_LENGTH", "8")

	cfg := config.Load()

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "production", cfg.Server.Env)
	assert.Equal(t, "prod-db", cfg.Database.Host)
	assert.Equal(t, 8, cfg.App.ShortCodeLength)

	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("APP_SHORT_CODE_LENGTH")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := config.Load()
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestIsProduction(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	cfg := config.Load()
	assert.True(t, cfg.IsProduction())
	assert.False(t, cfg.IsDevelopment())
	os.Unsetenv("ENVIRONMENT")
}

func TestIsDevelopment(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	cfg := config.Load()
	assert.True(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
	os.Unsetenv("ENVIRONMENT")
}
