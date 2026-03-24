package main

import (
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := loadConfig()
	if cfg.Transport != "stdio" {
		t.Errorf("default transport = %q, want stdio", cfg.Transport)
	}
	if cfg.HTTPPort != 8080 {
		t.Errorf("default port = %d, want 8080", cfg.HTTPPort)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log level = %q, want info", cfg.LogLevel)
	}
	if cfg.AuthMode != "env" {
		t.Errorf("default auth mode = %q, want env", cfg.AuthMode)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("AUTOTASK_USERNAME", "testuser")
	t.Setenv("AUTOTASK_SECRET", "testsecret")
	t.Setenv("AUTOTASK_INTEGRATION_CODE", "TESTCODE")
	t.Setenv("MCP_TRANSPORT", "http")
	t.Setenv("MCP_HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	cfg := loadConfig()
	if cfg.Username != "testuser" {
		t.Errorf("username = %q, want testuser", cfg.Username)
	}
	if cfg.Secret != "testsecret" {
		t.Errorf("secret = %q, want testsecret", cfg.Secret)
	}
	if cfg.IntegrationCode != "TESTCODE" {
		t.Errorf("integration code = %q, want TESTCODE", cfg.IntegrationCode)
	}
	if cfg.Transport != "http" {
		t.Errorf("transport = %q, want http", cfg.Transport)
	}
	if cfg.HTTPPort != 9090 {
		t.Errorf("port = %d, want 9090", cfg.HTTPPort)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log level = %q, want debug", cfg.LogLevel)
	}
}
