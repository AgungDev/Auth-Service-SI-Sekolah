package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string

	// Server
	Port string

	// JWT
	JWTSecret            string
	JWTExpiry            int
	RefreshTokenExpiry   int

	// Environment
	Environment string
}

// Load loads the configuration from environment variables
func Load() *Config {
	// Try to load .env file, ignore if not found
	godotenv.Load()

	return &Config{
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5433"),
		DBName: getEnv("DB_NAME", "auth_db"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASS", "postgres"),
		Port:   getEnv("PORT", "8000"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		JWTExpiry: getEnvAsInt("JWT_EXPIRY", 1800),
		RefreshTokenExpiry: getEnvAsInt("REFRESH_TOKEN_EXPIRY", 604800),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
