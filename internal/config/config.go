package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
}

func Load() *Config {
	// Загружаем .env файл (если есть)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "3007"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
