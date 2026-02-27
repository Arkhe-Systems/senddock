package config

import "os"

type Config struct {
	Port        string
	DatabaseUrl string
	RedisUrl    string
	JWTSecret   string
	FrontendURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", ""),
		RedisUrl:    getEnv("REDIS_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
