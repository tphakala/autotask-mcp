package main

import (
	"os"
	"strconv"
)

type Config struct {
	// Autotask credentials
	Username        string
	Secret          string
	IntegrationCode string
	APIURL          string // optional override

	// Server
	ServerName string
	Transport  string // "stdio" or "http"
	HTTPPort   int
	HTTPHost   string

	// Logging
	LogLevel string

	// Auth mode
	AuthMode    string // "env" or "gateway"
	LazyLoading bool
}

func loadConfig() Config {
	cfg := Config{
		Username:        os.Getenv("AUTOTASK_USERNAME"),
		Secret:          os.Getenv("AUTOTASK_SECRET"),
		IntegrationCode: os.Getenv("AUTOTASK_INTEGRATION_CODE"),
		APIURL:          os.Getenv("AUTOTASK_API_URL"),
		ServerName:      envOr("MCP_SERVER_NAME", "autotask-mcp"),
		Transport:       envOr("MCP_TRANSPORT", "stdio"),
		HTTPHost:        envOr("MCP_HTTP_HOST", "0.0.0.0"),
		LogLevel:        envOr("LOG_LEVEL", "info"),
		AuthMode:        envOr("AUTH_MODE", "env"),
		LazyLoading:     os.Getenv("LAZY_LOADING") == "true",
	}

	if port := os.Getenv("MCP_HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.HTTPPort = p
		}
	}
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}

	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
