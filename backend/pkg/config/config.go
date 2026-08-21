package config

import (
	"log/slog"
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
	IsCloud            bool
	RateLimitPerMinute int64
	WatchtowerURL      string
	WatchtowerToken    string
	TemplateLibraryURL string
}

func Load() Config {
	frontendURL := strings.TrimRight(getEnv("FRONTEND_URL", ""), "/")
	publicURL := strings.TrimRight(getEnv("PUBLIC_URL", ""), "/")

	isCloud := resolveCloud()
	if !isCloud && strings.HasPrefix(publicURL, "http://localhost") {
		slog.Warn("the public URL points at localhost; unsubscribe and tracking links in outgoing emails will not work outside this machine. Set it under Instance in the dashboard.")
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseUrl:        getEnv("DATABASE_URL", ""),
		RedisUrl:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		FrontendURL:        frontendURL,
		PublicURL:          publicURL,
		IsCloud:            isCloud,
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
		slog.Warn("invalid environment value, using default", "key", key, "value", value, "default", fallback)
		return fallback
	}
	return n
}

func (c Config) IsSelfHosted() bool {
	return !c.IsCloud
}

func (c Config) DeploymentModeName() string {
	if c.IsCloud {
		return "cloud"
	}
	return "self-hosted"
}

func resolveCloud() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CLOUD"))) {
	case "true", "1", "yes":
		return true
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv("DEPLOYMENT_MODE")), "cloud") {
		slog.Warn("DEPRECATION: DEPLOYMENT_MODE=cloud has been replaced by CLOUD=true and will stop being read in v0.9.")
		return true
	}

	return false
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
