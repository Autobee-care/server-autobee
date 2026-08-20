// Package config provides centralized application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Mongo    MongoConfig
	JWT      JWTConfig
	CORS     CORSConfig
	LogLevel string
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env  string
	Name string
	Port string
}

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI      string
	Database string
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// Load reads .env.local / .env (if present) and returns a validated Config.
// The application will fail fast if any required variable is missing or invalid.
func Load() (*Config, error) {
	// Try loading .env.local first, then .env; ignore error if absent (e.g. in Docker/CI).
	_ = godotenv.Load(".env.local", ".env")

	cfg := &Config{}

	// --- App ---
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Name = getEnv("APP_NAME", "autobee-server")
	cfg.App.Port = getEnv("PORT", "8080")

	// --- Mongo ---
	mongoURI, err := requireEnv("MONGO_URI")
	if err != nil {
		return nil, err
	}
	cfg.Mongo.URI = mongoURI

	mongoDB, err := requireEnv("MONGO_DATABASE")
	if err != nil {
		return nil, err
	}
	cfg.Mongo.Database = mongoDB

	// --- JWT ---
	accessSecret, err := requireEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}
	cfg.JWT.AccessSecret = accessSecret

	refreshSecret, err := requireEnv("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}
	cfg.JWT.RefreshSecret = refreshSecret

	accessExp, err := parseDuration("JWT_ACCESS_EXPIRATION", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.JWT.AccessExpiration = accessExp

	refreshExp, err := parseDuration("JWT_REFRESH_EXPIRATION", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.JWT.RefreshExpiration = refreshExp

	// --- CORS ---
	cfg.CORS.AllowedOrigins = splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"))
	cfg.CORS.AllowedMethods = splitCSV(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"))
	cfg.CORS.AllowedHeaders = splitCSV(getEnv("CORS_ALLOWED_HEADERS", "Accept,Authorization,Content-Type,X-Request-ID"))

	allowCreds, _ := strconv.ParseBool(getEnv("CORS_ALLOW_CREDENTIALS", "true"))
	cfg.CORS.AllowCredentials = allowCreds

	// --- Logging ---
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")

	return cfg, nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// requireEnv returns the value of an environment variable or an error if it is empty.
func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// parseDuration parses a Go duration string from an env var, using a default if absent.
func parseDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %q: %w", key, err)
	}
	return d, nil
}

// splitCSV splits a comma-separated string and trims whitespace from each element.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
