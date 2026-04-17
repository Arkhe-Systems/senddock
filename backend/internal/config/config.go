package config

import "os"

type Config struct {
	Port               string
	DatabaseUrl        string
	RedisUrl           string
	JWTSecret          string
	FrontendURL        string
	DeploymentMode     string
	WaitlistProjectID  string
	WaitlistTemplateID string
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseUrl:        getEnv("DATABASE_URL", ""),
		RedisUrl:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		DeploymentMode:     getEnv("DEPLOYMENT_MODE", "self-hosted"),
		WaitlistProjectID:  getEnv("WAITLIST_PROJECT_ID", ""),
		WaitlistTemplateID: getEnv("WAITLIST_TEMPLATE_ID", ""),
	}
}

func (c Config) IsSelfHosted() bool {
	return c.DeploymentMode != "cloud"
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
