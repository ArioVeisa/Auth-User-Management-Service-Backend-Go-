package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	RedisURL    string

	JWTSigningKey       string
	JWTAlgo             string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	RefreshTokenSecret  string

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	RateLimitRequests int
	RateLimitWindow   time.Duration
	MaxFailedLogins   int
	LockDuration      time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	rateLimitReqs, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "5"))

	return &Config{
		AppEnv:              getEnv("APP_ENV", "dev"),
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSigningKey:       getEnv("JWT_SIGNING_KEY", "your-super-secret-key-change-in-production"),
		JWTAlgo:             getEnv("JWT_ALGO", "HS256"),
		AccessTokenExpiry:   15 * time.Minute,
		RefreshTokenExpiry:  30 * 24 * time.Hour,
		RefreshTokenSecret:  getEnv("REFRESH_TOKEN_SECRET", "refresh-secret-key"),
		SMTPHost:            getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:            smtpPort,
		SMTPUser:            getEnv("SMTP_USER", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		RateLimitRequests:   rateLimitReqs,
		RateLimitWindow:     time.Second,
		MaxFailedLogins:     5,
		LockDuration:        15 * time.Minute,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
