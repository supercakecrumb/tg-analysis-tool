package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"DATABASE_URL":      os.Getenv("DATABASE_URL"),
		"CLICKHOUSE_URL":    os.Getenv("CLICKHOUSE_URL"),
		"TELEGRAM_API_ID":   os.Getenv("TELEGRAM_API_ID"),
		"TELEGRAM_API_HASH": os.Getenv("TELEGRAM_API_HASH"),
		"SESSION_ENC_KEY":   os.Getenv("SESSION_ENC_KEY"),
		"COOKIE_SECRET":     os.Getenv("COOKIE_SECRET"),
	}

	// Restore env vars after test
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"DATABASE_URL":      "postgres://user:pass@localhost/db",
				"CLICKHOUSE_URL":    "clickhouse://localhost:9000/db",
				"TELEGRAM_API_ID":   "12345",
				"TELEGRAM_API_HASH": "abcdef123456",
				"SESSION_ENC_KEY":   "12345678901234567890123456789012",
				"COOKIE_SECRET":     "secret123",
				"PORT":              "9090",
				"TIMEZONE":          "America/New_York",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != 9090 {
					t.Errorf("expected port 9090, got %d", cfg.Port)
				}
				if cfg.TelegramAPIID != 12345 {
					t.Errorf("expected telegram API ID 12345, got %d", cfg.TelegramAPIID)
				}
				if cfg.Timezone.String() != "America/New_York" {
					t.Errorf("expected timezone America/New_York, got %s", cfg.Timezone.String())
				}
			},
		},
		{
			name: "missing required DATABASE_URL",
			envVars: map[string]string{
				"CLICKHOUSE_URL":    "clickhouse://localhost:9000/db",
				"TELEGRAM_API_ID":   "12345",
				"TELEGRAM_API_HASH": "abcdef123456",
				"SESSION_ENC_KEY":   "12345678901234567890123456789012",
				"COOKIE_SECRET":     "secret123",
			},
			wantErr: true,
		},
		{
			name: "invalid TELEGRAM_API_ID",
			envVars: map[string]string{
				"DATABASE_URL":      "postgres://user:pass@localhost/db",
				"CLICKHOUSE_URL":    "clickhouse://localhost:9000/db",
				"TELEGRAM_API_ID":   "not-a-number",
				"TELEGRAM_API_HASH": "abcdef123456",
				"SESSION_ENC_KEY":   "12345678901234567890123456789012",
				"COOKIE_SECRET":     "secret123",
			},
			wantErr: true,
		},
		{
			name: "short SESSION_ENC_KEY",
			envVars: map[string]string{
				"DATABASE_URL":      "postgres://user:pass@localhost/db",
				"CLICKHOUSE_URL":    "clickhouse://localhost:9000/db",
				"TELEGRAM_API_ID":   "12345",
				"TELEGRAM_API_HASH": "abcdef123456",
				"SESSION_ENC_KEY":   "tooshort",
				"COOKIE_SECRET":     "secret123",
			},
			wantErr: true,
		},
		{
			name: "default values",
			envVars: map[string]string{
				"DATABASE_URL":      "postgres://user:pass@localhost/db",
				"CLICKHOUSE_URL":    "clickhouse://localhost:9000/db",
				"TELEGRAM_API_ID":   "12345",
				"TELEGRAM_API_HASH": "abcdef123456",
				"SESSION_ENC_KEY":   "12345678901234567890123456789012",
				"COOKIE_SECRET":     "secret123",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Port != 9090 {
					t.Errorf("expected port 9090 (from env), got %d", cfg.Port)
				}
				if cfg.Env != "development" {
					t.Errorf("expected default env development, got %s", cfg.Env)
				}
				if cfg.LogLevel != "info" {
					t.Errorf("expected default log level info, got %s", cfg.LogLevel)
				}
				if cfg.EnableMetrics != false {
					t.Errorf("expected metrics disabled by default")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			for k := range originalEnv {
				os.Unsetenv(k)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if got := getEnv("TEST_VAR", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want %v", got, "test_value")
	}

	if got := getEnv("NONEXISTENT_VAR", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want %v", got, "default")
	}
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_INVALID", "not-a-number")
	defer os.Unsetenv("TEST_INT")
	defer os.Unsetenv("TEST_INVALID")

	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 42)
	}

	if got := getEnvAsInt("TEST_INVALID", 99); got != 99 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 99)
	}

	if got := getEnvAsInt("NONEXISTENT", 100); got != 100 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 100)
	}
}

func TestGetEnvAsBool(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
	}

	for _, tc := range testCases {
		os.Setenv("TEST_BOOL", tc.value)
		if got := getEnvAsBool("TEST_BOOL", false); got != tc.expected {
			t.Errorf("getEnvAsBool(%s) = %v, want %v", tc.value, got, tc.expected)
		}
		os.Unsetenv("TEST_BOOL")
	}

	// Test invalid value
	os.Setenv("TEST_BOOL", "invalid")
	if got := getEnvAsBool("TEST_BOOL", true); got != true {
		t.Errorf("getEnvAsBool(invalid) = %v, want %v", got, true)
	}
	os.Unsetenv("TEST_BOOL")

	// Test nonexistent
	if got := getEnvAsBool("NONEXISTENT", false); got != false {
		t.Errorf("getEnvAsBool(nonexistent) = %v, want %v", got, false)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				DatabaseURL:     "postgres://localhost",
				ClickhouseURL:   "clickhouse://localhost",
				TelegramAPIID:   12345,
				TelegramAPIHash: "hash",
				SessionEncKey:   "12345678901234567890123456789012",
				CookieSecret:    "secret",
			},
			wantErr: false,
		},
		{
			name: "missing database URL",
			config: Config{
				ClickhouseURL:   "clickhouse://localhost",
				TelegramAPIID:   12345,
				TelegramAPIHash: "hash",
				SessionEncKey:   "12345678901234567890123456789012",
				CookieSecret:    "secret",
			},
			wantErr: true,
			errMsg:  "DATABASE_URL is required",
		},
		{
			name: "missing telegram API ID",
			config: Config{
				DatabaseURL:     "postgres://localhost",
				ClickhouseURL:   "clickhouse://localhost",
				TelegramAPIHash: "hash",
				SessionEncKey:   "12345678901234567890123456789012",
				CookieSecret:    "secret",
			},
			wantErr: true,
			errMsg:  "TELEGRAM_API_ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
