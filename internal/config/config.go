package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
	TgToken     string
	LogLevel    string
	APIUrl      string
	WebhookURL  string
}

func MustLoadConfig() *Config {
	_ = godotenv.Load(".env")

	_ = godotenv.Load("/app/.env")
	_ = godotenv.Load("/root/.env")

	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", ""),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		TgToken:     getEnv("TG_TOKEN", ""),
		LogLevel:    getEnv("LOG_LEVEL", "Debug"),
		APIUrl:      getEnv("API_URL", ""),
		WebhookURL:  getEnv("WEBHOOK_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
