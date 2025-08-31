package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port int
	Env  string

	// Database
	DatabaseURL   string
	ClickhouseURL string
	RedisURL      string

	// Telegram
	TelegramAPIID   int
	TelegramAPIHash string

	// Security
	SessionEncKey string
	CookieSecret  string

	// Application
	BaseURL  string
	Timezone *time.Location

	// Features
	EnableMetrics bool
	EnablePprof   bool
	LogLevel      string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnvAsInt("PORT", 8080),
		Env:           getEnv("ENV", "development"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		ClickhouseURL: getEnv("CLICKHOUSE_URL", ""),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:3000"),
		SessionEncKey: getEnv("SESSION_ENC_KEY", ""),
		CookieSecret:  getEnv("COOKIE_SECRET", ""),
		EnableMetrics: getEnvAsBool("ENABLE_METRICS", false),
		EnablePprof:   getEnvAsBool("ENABLE_PPROF", false),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	// Parse Telegram API credentials
	apiID := getEnv("TELEGRAM_API_ID", "")
	if apiID != "" {
		id, err := strconv.Atoi(apiID)
		if err != nil {
			return nil, fmt.Errorf("invalid TELEGRAM_API_ID: %w", err)
		}
		cfg.TelegramAPIID = id
	}
	cfg.TelegramAPIHash = getEnv("TELEGRAM_API_HASH", "")

	// Parse timezone
	tzName := getEnv("TIMEZONE", "Europe/Moscow")
	tz, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", tzName, err)
	}
	cfg.Timezone = tz

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration is present
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.ClickhouseURL == "" {
		return fmt.Errorf("CLICKHOUSE_URL is required")
	}
	if c.TelegramAPIID == 0 {
		return fmt.Errorf("TELEGRAM_API_ID is required")
	}
	if c.TelegramAPIHash == "" {
		return fmt.Errorf("TELEGRAM_API_HASH is required")
	}
	if c.SessionEncKey == "" {
		return fmt.Errorf("SESSION_ENC_KEY is required")
	}
	if c.CookieSecret == "" {
		return fmt.Errorf("COOKIE_SECRET is required")
	}
	if len(c.SessionEncKey) < 32 {
		return fmt.Errorf("SESSION_ENC_KEY must be at least 32 characters")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return defaultValue
}
