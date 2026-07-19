package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DatabaseUrl        string
	RedisUrl           string
	JWTSecret          string
	FrontendURL        string
	PublicURL          string
	DeploymentMode     string
	RateLimitPerMinute int64
	WatchtowerURL      string
	WatchtowerToken    string
	TemplateLibraryURL string
}

func Load() Config {
	frontendURL := strings.TrimRight(getEnv("FRONTEND_URL", ""), "/")
	publicURL := strings.TrimRight(getEnv("PUBLIC_URL", ""), "/")

	mode := getEnv("DEPLOYMENT_MODE", "self-hosted")
	if mode != "cloud" && strings.HasPrefix(publicURL, "http://localhost") {
		log.Println("WARNING: the public URL points at localhost. Unsubscribe and tracking links in outgoing emails will not work outside this machine. Set it under Instance in the dashboard.")
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseUrl:        getEnv("DATABASE_URL", ""),
		RedisUrl:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		FrontendURL:        frontendURL,
		PublicURL:          publicURL,
		DeploymentMode:     mode,
		RateLimitPerMinute: getEnvInt64("RATE_LIMIT_PER_MINUTE", 600),
		WatchtowerURL:      strings.TrimSpace(getEnv("SENDDOCK_WATCHTOWER_URL", "")),
		WatchtowerToken:    strings.TrimSpace(getEnv("SENDDOCK_WATCHTOWER_TOKEN", "")),
		TemplateLibraryURL: strings.TrimSpace(getEnv("TEMPLATE_LIBRARY_URL", "https://raw.githubusercontent.com/Arkhe-Systems/senddock-templates/main/index.json")),
	}
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		log.Printf("Invalid %s=%q, using default %d", key, value, fallback)
		return fallback
	}
	return n
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
