package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Cache       CacheConfig
	Providers   ProvidersConfig
	Log         LogConfig
	Auth        AuthConfig
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RateLimit       int
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	SSLMode        string
	MaxConnections int
	MaxIdleTime    time.Duration
	MaxLifetime    time.Duration
}

type CacheConfig struct {
	Type     string
	Host     string
	Port     int
	Password string
	DB       int
	TTL      time.Duration
	MaxSize  int
}

type ProvidersConfig struct {
	Provider1 ProviderConfig
	Provider2 ProviderConfig
}

type ProviderConfig struct {
	URL        string
	RateLimit  int
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
}

type LogConfig struct {
	Level  string
	Output string
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			RateLimit:       getEnvAsInt("SERVER_RATE_LIMIT", 100),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnvAsInt("DB_PORT", 5432),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", "postgres"),
			DBName:         getEnv("DB_NAME", "search_engine"),
			SSLMode:        getEnv("DB_SSLMODE", "disable"),
			MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleTime:    getEnvAsDuration("DB_MAX_IDLE_TIME", 5*time.Minute),
			MaxLifetime:    getEnvAsDuration("DB_MAX_LIFETIME", 1*time.Hour),
		},
		Cache: CacheConfig{
			Type:     getEnv("CACHE_TYPE", "memory"),
			Host:     getEnv("CACHE_HOST", "localhost"),
			Port:     getEnvAsInt("CACHE_PORT", 6379),
			Password: getEnv("CACHE_PASSWORD", ""),
			DB:       getEnvAsInt("CACHE_DB", 0),
			TTL:      getEnvAsDuration("CACHE_TTL", 5*time.Minute),
			MaxSize:  getEnvAsInt("CACHE_MAX_SIZE", 1000),
		},
		Providers: ProvidersConfig{
			Provider1: ProviderConfig{
				URL:        getEnv("PROVIDER1_URL", "http://localhost:3001/api/content"),
				RateLimit:  getEnvAsInt("PROVIDER1_RATE_LIMIT", 60),
				Timeout:    getEnvAsDuration("PROVIDER1_TIMEOUT", 5*time.Second),
				RetryCount: getEnvAsInt("PROVIDER1_RETRY_COUNT", 3),
				RetryDelay: getEnvAsDuration("PROVIDER1_RETRY_DELAY", 1*time.Second),
			},
			Provider2: ProviderConfig{
				URL:        getEnv("PROVIDER2_URL", "http://localhost:3002/api/content"),
				RateLimit:  getEnvAsInt("PROVIDER2_RATE_LIMIT", 60),
				Timeout:    getEnvAsDuration("PROVIDER2_TIMEOUT", 5*time.Second),
				RetryCount: getEnvAsInt("PROVIDER2_RETRY_COUNT", 3),
				RetryDelay: getEnvAsDuration("PROVIDER2_RETRY_DELAY", 1*time.Second),
			},
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
