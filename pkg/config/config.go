package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	Path         string
	MaxOpenConns int
}

type AuthConfig struct {
	PrivateKeyPath   string
	PublicKeyPath    string
	AccessTTLMinutes int
	RefreshTTLDays   int
}

type RedisConfig struct {
	URL string
}

type LogConfig struct {
	Level  string
	Format string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Redis    RedisConfig
	Log      LogConfig
}

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
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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
