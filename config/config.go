package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment     string
	Port            string
	LogLevel        string
	DefaultLanguage string
	JWTSecret       string
	MongoDB         MongoDBConfig
	PostgresDB      PostgresDBConfig
}

type MongoDBConfig struct {
	Enabled  bool
	URI      string
	Database string
	Username string
	Password string
	Host     string
	Port     string
}

type PostgresDBConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string
}

func Load() *Config {
	return &Config{
		Environment:     getEnv("ENVIRONMENT", "development"),
		Port:            getEnv("PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		DefaultLanguage: getEnv("DEFAULT_LANGUAGE", "en"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		MongoDB: MongoDBConfig{
			Enabled:  getBoolEnv("MONGODB_ENABLED", true),
			URI:      getEnv("MONGODB_URI", ""),
			Database: getEnv("MONGODB_DATABASE", "mygo"),
			Username: getEnv("MONGODB_USERNAME", "root"),
			Password: getEnv("MONGODB_PASSWORD", "4jClkoZfth8Jq4lB"),
			Host:     getEnv("MONGODB_HOST", "localhost"),
			Port:     getEnv("MONGODB_PORT", "27017"),
		},
		PostgresDB: PostgresDBConfig{
			Enabled:  getBoolEnv("POSTGRES_ENABLED", false),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			Username: getEnv("POSTGRES_USERNAME", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			Database: getEnv("POSTGRES_DATABASE", "backend_template"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(strings.ToLower(value))
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}
