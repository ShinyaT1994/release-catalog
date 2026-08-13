package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	DatabaseDriver string
	DatabaseDSN    string
	ServerPort     int
	DTBaseURL      string
	DTAPIKey       string
	DTStubMode     bool
	DTTimeout      time.Duration
	LogLevelStr    string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseDriver: getEnv("RC_DATABASE_DRIVER", "sqlite"),
		DatabaseDSN:    getEnv("RC_DATABASE_DSN", "./release-catalog.db"),
		ServerPort:     getEnvInt("RC_SERVER_PORT", 8080),
		DTBaseURL:      getEnv("RC_DT_BASE_URL", "http://localhost:8081"),
		DTAPIKey:       getEnv("RC_DT_API_KEY", ""),
		DTStubMode:     getEnvBool("RC_DT_STUB_MODE", true),
		DTTimeout:      time.Duration(getEnvInt("RC_DT_TIMEOUT_SECONDS", 30)) * time.Second,
		LogLevelStr:    getEnv("RC_LOG_LEVEL", "info"),
	}
}

// LogLevel returns the slog level from config
func (c *Config) LogLevel() slog.Level {
	switch c.LogLevelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}
