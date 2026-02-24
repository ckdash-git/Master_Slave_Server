package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
type Config struct {
	DatabaseURL      string
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	OTCExpiry        time.Duration
	ServerPort       string
}

// Load reads configuration from environment variables (with .env fallback).
func Load() *Config {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://masteruser:masterpass@localhost:5432/master_slave_db?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTAccessExpiry:  parseDuration("JWT_ACCESS_EXPIRY", "15m"),
		JWTRefreshExpiry: parseDuration("JWT_REFRESH_EXPIRY", "168h"),
		OTCExpiry:        parseDuration("OTC_EXPIRY", "30s"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}

	if cfg.JWTSecret == "dev-secret-change-me" {
		log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseDuration(key, fallback string) time.Duration {
	raw := getEnv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("⚠️  Invalid duration for %s=%q, using fallback %s", key, raw, fallback)
		d, _ = time.ParseDuration(fallback)
	}
	return d
}
