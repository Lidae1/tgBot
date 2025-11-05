package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoadConfig(t *testing.T) {
	originalEnv := map[string]string{}
	envVars := []string{
		"DATABASE_URL",
		"TG_TOKEN",
		"LOG_LEVEL",
		"HTTP_PORT",
		"APP_URL",
		"WEBHOOK_URL",
	}

	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	os.Setenv("DATABASE_URL", "test_db_url")
	os.Setenv("TG_TOKEN", "test_token")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("APP_URL", "https://test-api.com")
	os.Setenv("WEBHOOK_URL", "https://test-webhook.com")

	cfg := MustLoadConfig()

	assert.Equal(t, "test_db_url", cfg.DatabaseURL)
	assert.Equal(t, "test_token", cfg.TgToken)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "9090", cfg.HTTPPort)
	assert.Equal(t, "https://test-api.com", cfg.APIUrl)
	assert.Equal(t, "test_webhook_url", cfg.WebhookURL)

	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("API_URL")

	cfg = MustLoadConfig()
	assert.Equal(t, "8080", cfg.HTTPPort)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "https://api.bybit.com/v5/market/tickers", cfg.APIUrl)
}

func TestGetEnv(t *testing.T) {

	originalValue := os.Getenv("TEST_VAR")
	os.Unsetenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", originalValue)

	result := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "default_value", result)

	os.Setenv("TEST_VAR", "custom_value")
	result = getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "custom_value", result)
}
