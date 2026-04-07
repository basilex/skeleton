// Package config provides configuration loading and management for the application.
// It supports loading configuration values from environment files and provides
// typed configuration structures for various application components.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds application-level configuration settings.
type AppConfig struct {
	Env  string
	Port string
	Name string
}

// DatabaseConfig holds database connection configuration settings.
type DatabaseConfig struct {
	Path         string
	MaxOpenConns int
}

// AuthConfig holds authentication and JWT configuration settings.
type AuthConfig struct {
	PrivateKeyPath   string
	PublicKeyPath    string
	AccessTTLMinutes int
	RefreshTTLDays   int
}

// RedisConfig holds Redis connection configuration settings.
type RedisConfig struct {
	URL string
}

// SessionConfig holds session management configuration settings.
type SessionConfig struct {
	CookieName   string
	CookieDomain string
	CookieSecure bool
	TTLMinutes   int
}

// LogConfig holds logging configuration settings.
type LogConfig struct {
	Level  string
	Format string
}

// Config is the root configuration structure that aggregates all
// application configuration settings.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Redis    RedisConfig
	Session  SessionConfig
	Log      LogConfig
}

// Load reads configuration from environment files and environment variables.
// It loads the appropriate .env file based on APP_ENV (defaults to "dev").
// Returns a fully populated Config struct or an error if the env file cannot be loaded.
func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	envFile := fmt.Sprintf("configs/.env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("load env file %s: %w", envFile, err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", env),
			Port: getEnv("APP_PORT", "8080"),
			Name: getEnv("APP_NAME", "skeleton"),
		},
		Database: DatabaseConfig{
			Path:         getEnv("DB_PATH", "./data/skeleton.db"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 1),
		},
		Auth: AuthConfig{
			PrivateKeyPath:   getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:    getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
			AccessTTLMinutes: getEnvInt("JWT_ACCESS_TTL_MINUTES", 15),
			RefreshTTLDays:   getEnvInt("JWT_REFRESH_TTL_DAYS", 7),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		Session: SessionConfig{
			CookieName:   getEnv("SESSION_COOKIE_NAME", "session"),
			CookieDomain: getEnv("SESSION_COOKIE_DOMAIN", ""),
			CookieSecure: getEnvBool("SESSION_COOKIE_SECURE", false),
			TTLMinutes:   getEnvInt("SESSION_TTL_MINUTES", 1440),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

// getEnv retrieves an environment variable value or returns the fallback if not set.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt retrieves an environment variable as an integer or returns the fallback
// if not set or if parsing fails.
func getEnvInt(key string, fallback int) int {
	s := getEnv(key, "")
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

// getEnvBool retrieves an environment variable as a boolean or returns the fallback
// if not set. Accepts "true", "1", or "yes" as true values.
func getEnvBool(key string, fallback bool) bool {
	s := getEnv(key, "")
	if s == "" {
		return fallback
	}
	return s == "true" || s == "1" || s == "yes"
}
