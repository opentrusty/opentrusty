package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Session       SessionConfig
	Observability ObservabilityConfig
	Security      SecurityConfig
	RateLimit     RateLimitConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// SessionConfig holds session management configuration
type SessionConfig struct {
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
	Lifetime       time.Duration
	IdleTimeout    time.Duration
}

// ObservabilityConfig holds logging and tracing configuration
type ObservabilityConfig struct {
	LogLevel       string
	LogFormat      string
	OTELEnabled    bool
	ServiceName    string
	ServiceVersion string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	Argon2Memory       uint32
	Argon2Iterations   uint32
	Argon2Parallelism  uint8
	Argon2SaltLength   uint32
	Argon2KeyLength    uint32
	LockoutMaxAttempts int
	LockoutDuration    time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  parseDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: parseDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  parseDuration("SERVER_IDLE_TIMEOUT", "60s"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "opentrusty"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "opentrusty"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    parseInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    parseInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: parseDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		Session: SessionConfig{
			CookieName:     getEnv("SESSION_COOKIE_NAME", "opentrusty_session"),
			CookieDomain:   getEnv("SESSION_COOKIE_DOMAIN", ""),
			CookiePath:     getEnv("SESSION_COOKIE_PATH", "/"),
			CookieSecure:   parseBool("SESSION_COOKIE_SECURE", false),
			CookieHTTPOnly: parseBool("SESSION_COOKIE_HTTP_ONLY", true),
			CookieSameSite: getEnv("SESSION_COOKIE_SAME_SITE", "Lax"),
			Lifetime:       parseDuration("SESSION_LIFETIME", "24h"),
			IdleTimeout:    parseDuration("SESSION_IDLE_TIMEOUT", "30m"),
		},
		Observability: ObservabilityConfig{
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			LogFormat:      getEnv("LOG_FORMAT", "json"),
			OTELEnabled:    parseBool("OTEL_ENABLED", false),
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "opentrusty"),
			ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "0.1.0"),
		},
		Security: SecurityConfig{
			Argon2Memory:       uint32(parseInt("ARGON2_MEMORY", 65536)),
			Argon2Iterations:   uint32(parseInt("ARGON2_ITERATIONS", 3)),
			Argon2Parallelism:  uint8(parseInt("ARGON2_PARALLELISM", 4)),
			Argon2SaltLength:   uint32(parseInt("ARGON2_SALT_LENGTH", 16)),
			Argon2KeyLength:    uint32(parseInt("ARGON2_KEY_LENGTH", 32)),
			LockoutMaxAttempts: parseInt("SECURITY_LOCKOUT_MAX_ATTEMPTS", 5),
			LockoutDuration:    parseDuration("SECURITY_LOCKOUT_DURATION", "15m"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: float64(parseInt("RATELIMIT_RPS", 10)),
			Burst:             parseInt("RATELIMIT_BURST", 20),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func parseDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	d, err := time.ParseDuration(value)
	if err != nil {
		// Fallback to default
		d, _ = time.ParseDuration(defaultValue)
	}
	return d
}
