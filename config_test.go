package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}
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
	// Multi-byte UTF-8: masking must operate on runes, not bytes. 🛡️ is two code
	// points (U+1F6E1 + U+FE0F), so the 10-rune input keeps the first 2 and last 2
	// runes and masks the middle 6. Byte-slicing would corrupt the output here.
	if s := maskSecret("🔑🔒secret🛡️"); s != "🔑🔒******🛡️" {
		t.Errorf("maskSecret(UTF-8) = %q, want %q", s, "🔑🔒******🛡️")
	}
}

// TestGetConfigField_UnsetPointerKeys pins that pointer-backed keys (http_port,
// lazy_loading) return empty when unset instead of a false "unknown config key".
func TestGetConfigField_UnsetPointerKeys(t *testing.T) {
	fc := FileConfig{Username: "alice"}

	for _, key := range []string{"http_port", "httpport", "lazy_loading", "lazyloading"} {
		val, err := getConfigField(fc, key)
		if err != nil {
			t.Errorf("getConfigField(%q) unset: unexpected error %v", key, err)
		}
		if val != "" {
			t.Errorf("getConfigField(%q) unset: expected empty, got %q", key, val)
		}
	}

	// A genuinely unknown key must still error.
	if _, err := getConfigField(fc, "bogus"); err == nil {
		t.Error("expected error for unknown key 'bogus'")
	}

	// A set pointer value must still round-trip.
	port := 9090
	fc.HTTPPort = &port
	if val, err := getConfigField(fc, "http_port"); err != nil || val != "9090" {
		t.Errorf("getConfigField(http_port) set: got %q, err %v; want 9090", val, err)
	}
}

// TestValidateFileAPIURL pins the api_url allowlist: https to an autotask.net host
// (or subdomain) is accepted; everything else is rejected, so a tampered config
// file cannot redirect credential-bearing requests to an arbitrary endpoint.
func TestValidateFileAPIURL(t *testing.T) {
	valid := []string{
		"https://autotask.net",
		"https://webservices19.autotask.net/ATServicesRest",
		"https://WEBSERVICES2.AUTOTASK.NET/ATServicesRest",
	}
	for _, u := range valid {
		if err := validateFileAPIURL(u); err != nil {
			t.Errorf("validateFileAPIURL(%q): unexpected error %v", u, err)
		}
	}

	invalid := []string{
		"http://webservices19.autotask.net/ATServicesRest", // not https
		"https://evil.example.com",                         // wrong host
		"https://evil-autotask.net",                        // look-alike, no dot boundary
		"https://autotask.net.evil.com",                    // suffix trick
		"ftp://webservices19.autotask.net",                 // wrong scheme
		"not a url at all",                                 // unparseable as an absolute https URL
		"https://",                                         // no host
	}
	for _, u := range invalid {
		if err := validateFileAPIURL(u); err == nil {
			t.Errorf("validateFileAPIURL(%q): expected error, got nil", u)
		}
	}

	// The reject-path error must not echo URL userinfo (neither the password nor the
	// username): the message is printed to stderr. Use an http scheme so this hits the
	// scheme branch.
	err := validateFileAPIURL("http://secretuser:sup3rs3cret@webservices19.autotask.net")
	if err == nil {
		t.Fatal("expected error for an http api_url with userinfo")
	}
	if strings.Contains(err.Error(), "sup3rs3cret") {
		t.Errorf("error message leaked the api_url password: %q", err.Error())
	}
	if strings.Contains(err.Error(), "secretuser") {
		t.Errorf("error message leaked the api_url username: %q", err.Error())
	}
}

// TestLoadFileConfig_RefusesInsecurePerms verifies a group/world-accessible config
// file is refused (fail closed) rather than loaded, so credentials are never read
// from a file other users can access. Unix-only: Windows synthesizes perm bits.
func TestLoadFileConfig_RefusesInsecurePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")
	if err := saveFileConfig(cfgPath, FileConfig{Username: "u"}); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}
	if err := os.Chmod(cfgPath, 0644); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	fc, loaded, err := loadFileConfig(cfgPath)
	if err == nil {
		t.Error("expected error for group/world-readable config file")
	}
	if loaded {
		t.Error("expected loaded=false when refusing an insecure config file")
	}
	if fc.Username != "" {
		t.Errorf("expected no values from a refused config, got username %q", fc.Username)
	}
}

// TestLoadConfig_IgnoresBadFileAPIURL verifies the server config path drops a bad
// file api_url (so it cannot redirect credentials) while keeping the rest of the
// file, and never blocks startup. The config CLI can still read/repair such a file
// because validation lives on the server path, not in loadFileConfig.
func TestLoadConfig_IgnoresBadFileAPIURL(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	clearAllConfigEnv(t)

	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")
	if err := saveFileConfig(cfgPath, FileConfig{Username: "fileuser", APIURL: "http://evil.example.com"}); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	cfg := loadConfig()
	if cfg.APIURL != "" {
		t.Errorf("expected bad file api_url to be ignored, got %q", cfg.APIURL)
	}
	if cfg.Username != "fileuser" {
		t.Errorf("expected the rest of the file to still load, got username %q", cfg.Username)
	}

	// loadFileConfig itself no longer validates api_url, so the CLI can read it back.
	fc, loaded, err := loadFileConfig(cfgPath)
	if err != nil || !loaded {
		t.Fatalf("loadFileConfig should read a file with a bad api_url for CLI repair: err=%v loaded=%v", err, loaded)
	}
	if fc.APIURL != "http://evil.example.com" {
		t.Errorf("loadFileConfig should return the raw stored api_url, got %q", fc.APIURL)
	}
}

// TestSetConfigField_ValidatesAPIURL verifies api_url is validated on write, so the
// CLI cannot persist a value the server would refuse; an empty value clears it.
func TestSetConfigField_ValidatesAPIURL(t *testing.T) {
	var fc FileConfig
	if err := setConfigField(&fc, "api_url", "http://evil.example.com"); err == nil {
		t.Error("expected error setting a non-autotask.net api_url")
	}
	if fc.APIURL != "" {
		t.Errorf("bad api_url must not be stored, got %q", fc.APIURL)
	}
	const good = "https://webservices19.autotask.net/ATServicesRest"
	if err := setConfigField(&fc, "api_url", good); err != nil {
		t.Errorf("valid api_url rejected: %v", err)
	}
	if fc.APIURL != good {
		t.Errorf("api_url = %q, want %q", fc.APIURL, good)
	}
	if err := setConfigField(&fc, "api_url", ""); err != nil {
		t.Errorf("clearing api_url should succeed: %v", err)
	}
	if fc.APIURL != "" {
		t.Errorf("api_url should be cleared, got %q", fc.APIURL)
	}
}

// TestLoadFileConfig_AcceptsValidAPIURL verifies a legitimate https autotask.net
// api_url loads normally.
func TestLoadFileConfig_AcceptsValidAPIURL(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "autotask-mcp", "config.json")
	const apiURL = "https://webservices19.autotask.net/ATServicesRest"
	if err := saveFileConfig(cfgPath, FileConfig{Username: "u", APIURL: apiURL}); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	fc, loaded, err := loadFileConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !loaded {
		t.Error("expected loaded=true for a valid config file")
	}
	if fc.APIURL != apiURL {
		t.Errorf("api_url = %q, want %q", fc.APIURL, apiURL)
	}
}

// TestSaveFileConfig_TightensDirPermissions verifies saveFileConfig tightens an
// already-more-permissive config directory to 0700, since MkdirAll does not.
func TestSaveFileConfig_TightensDirPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}
	tempDir := t.TempDir()
	dir := filepath.Join(tempDir, "autotask-mcp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(dir, 0777); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	cfgPath := filepath.Join(dir, "config.json")
	if err := saveFileConfig(cfgPath, FileConfig{Username: "u"}); err != nil {
		t.Fatalf("saveFileConfig: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0700 {
		t.Errorf("config dir permissions = %#o, want 0700", perm)
	}
}
