package main

import (
	"os"
	"path/filepath"
	"testing"
)

func clearAllConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"AUTOTASK_USERNAME",
		"AUTOTASK_SECRET",
		"AUTOTASK_INTEGRATION_CODE",
		"AUTOTASK_API_URL",
		"MCP_SERVER_NAME",
		"MCP_TRANSPORT",
		"MCP_HTTP_HOST",
		"MCP_HTTP_PORT",
		"LOG_LEVEL",
		"AUTH_MODE",
		"LAZY_LOADING",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Point to empty temp dir for config home to prevent reading local files during test
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	clearAllConfigEnv(t)

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
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	clearAllConfigEnv(t)

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

func TestLoadConfig_FileConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	clearAllConfigEnv(t)

	port := 9191
	lazy := true
	fc := FileConfig{
		Username:        "fileuser@example.com",
		Secret:          "filesecret123",
		IntegrationCode: "FILECODE",
		HTTPPort:        &port,
		LogLevel:        "warn",
		LazyLoading:     &lazy,
	}

	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")
	if err := saveFileConfig(cfgPath, fc); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	cfg := loadConfig()
	if cfg.Username != "fileuser@example.com" {
		t.Errorf("username = %q, want fileuser@example.com", cfg.Username)
	}
	if cfg.Secret != "filesecret123" {
		t.Errorf("secret = %q, want filesecret123", cfg.Secret)
	}
	if cfg.IntegrationCode != "FILECODE" {
		t.Errorf("integration code = %q, want FILECODE", cfg.IntegrationCode)
	}
	if cfg.HTTPPort != 9191 {
		t.Errorf("http port = %d, want 9191", cfg.HTTPPort)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("log level = %q, want warn", cfg.LogLevel)
	}
	if !cfg.LazyLoading {
		t.Errorf("lazy loading = %v, want true", cfg.LazyLoading)
	}
	if cfg.ConfigFile != cfgPath {
		t.Errorf("config file = %q, want %q", cfg.ConfigFile, cfgPath)
	}
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("AUTOTASK_SECRET", "")
	t.Setenv("AUTOTASK_INTEGRATION_CODE", "")

	fc := FileConfig{
		Username: "fileuser",
		Secret:   "filesecret",
	}
	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")
	if err := saveFileConfig(cfgPath, fc); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	// Override Username via env, leave Secret from file
	t.Setenv("AUTOTASK_USERNAME", "envuser")

	cfg := loadConfig()
	if cfg.Username != "envuser" {
		t.Errorf("username = %q, want envuser (overridden)", cfg.Username)
	}
	if cfg.Secret != "filesecret" {
		t.Errorf("secret = %q, want filesecret (from file)", cfg.Secret)
	}
}

func TestSaveFileConfig_Permissions(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")

	fc := FileConfig{
		Username: "secureuser",
	}
	if err := saveFileConfig(cfgPath, fc); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = 0%o, want 0600", perm)
	}
}

func TestHandleConfigCommand(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	// 1. Path command
	if err := handleConfigCommand([]string{"path"}); err != nil {
		t.Errorf("config path failed: %v", err)
	}

	// 2. Set commands
	if err := handleConfigCommand([]string{"set", "username", "admin@corp.com"}); err != nil {
		t.Errorf("config set username failed: %v", err)
	}
	if err := handleConfigCommand([]string{"set", "secret", "mySuperSecretPassword"}); err != nil {
		t.Errorf("config set secret failed: %v", err)
	}
	if err := handleConfigCommand([]string{"set", "http_port", "9999"}); err != nil {
		t.Errorf("config set http_port failed: %v", err)
	}

	// 3. Get commands
	if err := handleConfigCommand([]string{"get"}); err != nil {
		t.Errorf("config get failed: %v", err)
	}
	if err := handleConfigCommand([]string{"get", "username"}); err != nil {
		t.Errorf("config get username failed: %v", err)
	}

	// 4. Unset command
	if err := handleConfigCommand([]string{"unset", "http_port"}); err != nil {
		t.Errorf("config unset http_port failed: %v", err)
	}

	// Verify boolean handling
	if err := handleConfigCommand([]string{"set", "lazy_loading", "true"}); err != nil {
		t.Errorf("config set lazy_loading true failed: %v", err)
	}
	if err := handleConfigCommand([]string{"set", "lazy_loading", "invalid_bool"}); err == nil {
		t.Errorf("expected error setting invalid boolean value, got nil")
	}

	// Verify loaded file config
	fc, loaded, err := loadFileConfig("")
	if err != nil || !loaded {
		t.Fatalf("loadFileConfig failed: %v", err)
	}
	if fc.Username != "admin@corp.com" {
		t.Errorf("username = %q, want admin@corp.com", fc.Username)
	}
	if fc.Secret != "mySuperSecretPassword" {
		t.Errorf("secret = %q, want mySuperSecretPassword", fc.Secret)
	}
	if fc.HTTPPort != nil {
		t.Errorf("http_port should be nil after unset, got: %v", *fc.HTTPPort)
	}
	if fc.LazyLoading == nil || !*fc.LazyLoading {
		t.Errorf("lazy_loading should be true, got: %v", fc.LazyLoading)
	}
}

func TestMaskSecret(t *testing.T) {
	if s := maskSecret(""); s != "" {
		t.Errorf("maskSecret(\"\") = %q, want empty", s)
	}
	if s := maskSecret("abc"); s != "****" {
		t.Errorf("maskSecret(\"abc\") = %q, want ****", s)
	}
	if s := maskSecret("12345678"); s != "12****78" {
		t.Errorf("maskSecret(\"12345678\") = %q, want 12****78", s)
	}
	// Multi-byte UTF-8 test
	if s := maskSecret("🔑🔒secret🛡️"); s != "🔑🔒****🛡️" && len(s) == 0 {
		t.Errorf("maskSecret with UTF-8 failed, got: %s", s)
	}
}
