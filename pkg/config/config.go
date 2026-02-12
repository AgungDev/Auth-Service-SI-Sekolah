package config

import (
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Driver   string
}

type HttpConfig struct {
	ApiPort string
}

type TokenConfig struct {
	AplicationName      string
	JwtSignatureKey     []byte
	JwtSigningMethod    *jwt.SigningMethodHMAC
	AccessTokenLifeTime int
	RefreshTokenExpiry  int
}

type Config struct {
	DBConfig
	HttpConfig
	TokenConfig
}

func (c *Config) readConfig() error {
	// Try to load .env file, ignore if not found
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	// database
	c.DBConfig = DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		Database: getEnv("DB_NAME", "my_laundry"),
		Username: getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASS", "admin"),
		Driver:   getEnv("DB_DRIVER", "postgres"),
	}

	// port
	c.HttpConfig = HttpConfig{
		ApiPort: getEnv("API_PORT", "8080"),
	}

	c.TokenConfig = TokenConfig{
		AplicationName:      "Enigma Camp",
		JwtSignatureKey:     []byte(getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")),
		JwtSigningMethod:    jwt.SigningMethodHS256,
		AccessTokenLifeTime: getEnvAsInt("JWT_EXPIRY", 1800),
		RefreshTokenExpiry:  getEnvAsInt("REFRESH_TOKEN_EXPIRY", 604800),
	}

	return nil
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := cfg.readConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// type Config struct {
// 	// Database
// 	DBHost string
// 	DBPort string
// 	DBName string
// 	DBUser string
// 	DBPass string

// 	// Server
// 	Port string

// 	// JWT
// 	JWTSecret            string
// 	JWTExpiry            int
// 	RefreshTokenExpiry   int

// 	// Environment
// 	Environment string
// }

// Load loads the configuration from environment variables
// func Load() *Config {
// 	// Try to load .env file, ignore if not found
// 	godotenv.Load()

// 	return &Config{
// 		DBHost: getEnv("DB_HOST", "localhost"),
// 		DBPort: getEnv("DB_PORT", "5433"),
// 		DBName: getEnv("DB_NAME", "auth_db"),
// 		DBUser: getEnv("DB_USER", "postgres"),
// 		DBPass: getEnv("DB_PASS", "postgres"),
// 		Port:   getEnv("PORT", "8000"),
// 		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
// 		JWTExpiry: getEnvAsInt("JWT_EXPIRY", 1800),
// 		RefreshTokenExpiry: getEnvAsInt("REFRESH_TOKEN_EXPIRY", 604800),
// 		Environment: getEnv("ENVIRONMENT", "development"),
// 	}
// }

// GetDSN returns the database connection string
// func (c *Config) GetDSN() string {
// 	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
// }

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
