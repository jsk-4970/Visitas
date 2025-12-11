package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	// GCP
	ProjectID       string
	Region          string
	SpannerInstance string
	SpannerDatabase string
	SpannerEmulator string

	// Server
	Port     string
	Env      string
	LogLevel string

	// Firebase
	FirebaseConfigPath string

	// Google Maps
	GoogleMapsAPIKey string

	// CORS
	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		ProjectID:          getEnv("GCP_PROJECT_ID", "visitas-dev"),
		Region:             getEnv("GCP_REGION", "asia-northeast1"),
		SpannerInstance:    getEnv("SPANNER_INSTANCE", "visitas-dev-instance"),
		SpannerDatabase:    getEnv("SPANNER_DATABASE", "visitas-dev-db"),
		SpannerEmulator:    getEnv("SPANNER_EMULATOR_HOST", ""),
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		FirebaseConfigPath: getEnv("FIREBASE_CONFIG_PATH", ""),
		GoogleMapsAPIKey:   getEnv("GOOGLE_MAPS_API_KEY", ""),
		AllowedOrigins:     strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID is required")
	}
	if c.SpannerInstance == "" {
		return fmt.Errorf("SPANNER_INSTANCE is required")
	}
	if c.SpannerDatabase == "" {
		return fmt.Errorf("SPANNER_DATABASE is required")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) UseEmulator() bool {
	return c.SpannerEmulator != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
