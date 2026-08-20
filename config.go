package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// ConfigFile source path (if loaded from file)
	ConfigFile string
}

// FileConfig defines the JSON structure for the on-disk config file (~/.config/autotask-mcp/config.json).
type FileConfig struct {
	Username        string `json:"username,omitempty"`
	Secret          string `json:"secret,omitempty"`
	IntegrationCode string `json:"integration_code,omitempty"`
	APIURL          string `json:"api_url,omitempty"`
	ServerName      string `json:"server_name,omitempty"`
	Transport       string `json:"transport,omitempty"`
	HTTPHost        string `json:"http_host,omitempty"`
	HTTPPort        *int   `json:"http_port,omitempty"`
	LogLevel        string `json:"log_level,omitempty"`
	AuthMode        string `json:"auth_mode,omitempty"`
	LazyLoading     *bool  `json:"lazy_loading,omitempty"`
}

// defaultConfigPath returns the standard XDG path for the config file.
func defaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "autotask-mcp", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "autotask-mcp", "config.json")
	}
	return filepath.Join(home, ".config", "autotask-mcp", "config.json")
}

// checkSecureFilePermissions checks whether file permissions are 0600 or stricter.
// If permissions are more permissive than 0600 on unix, it prints a warning to os.Stderr.
func checkSecureFilePermissions(path string, perm os.FileMode) {
	if perm&0077 != 0 {
		fmt.Fprintf(os.Stderr, "autotask-mcp: warning: config file %s has insecure permissions %#o (expected 0600)\n", path, perm)
	}
}

// loadFileConfig reads and parses the JSON config file from disk.
// Returns an empty FileConfig without error if the file does not exist.
func loadFileConfig(path string) (FileConfig, bool, error) {
	if path == "" {
		path = defaultConfigPath()
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return FileConfig{}, false, nil
	}
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("stat config file %s: %w", path, err)
	}
	if info.IsDir() {
		return FileConfig{}, false, fmt.Errorf("config path %s is a directory, expected JSON file", path)
	}

	checkSecureFilePermissions(path, info.Mode().Perm())

	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return FileConfig{}, false, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	return fc, true, nil
}

// saveFileConfig saves the FileConfig struct to disk with 0600 permissions and 0700 dir permissions.
func saveFileConfig(path string, fc FileConfig) error {
	if path == "" {
		path = defaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	tmpFile, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp config file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) //nolint:errcheck // clean up if rename fails

	if err := tmpFile.Chmod(0600); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("setting secure permissions on temp file: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing temp config file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("syncing temp config file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomically replacing config file %s: %w", path, err)
	}

	return nil
}

// loadConfig merges defaults, file-based config, and environment variables.
func loadConfig() Config {
	cfgPath := defaultConfigPath()
	fileCfg, loaded, err := loadFileConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "autotask-mcp: error loading config file %s: %v\n", cfgPath, err)
	}

	cfg := Config{
		ServerName:  "autotask-mcp",
		Transport:   "stdio",
		HTTPHost:    "0.0.0.0",
		HTTPPort:    8080,
		LogLevel:    "info",
		AuthMode:    "env",
		LazyLoading: false,
	}
	if loaded {
		cfg.ConfigFile = cfgPath
	}

	applyFileConfig(&cfg, fileCfg)
	applyEnvOverrides(&cfg)
	return cfg
}

func applyFileConfig(cfg *Config, fileCfg FileConfig) {
	if fileCfg.Username != "" {
		cfg.Username = fileCfg.Username
	}
	if fileCfg.Secret != "" {
		cfg.Secret = fileCfg.Secret
	}
	if fileCfg.IntegrationCode != "" {
		cfg.IntegrationCode = fileCfg.IntegrationCode
	}
	if fileCfg.APIURL != "" {
		cfg.APIURL = fileCfg.APIURL
	}
	if fileCfg.ServerName != "" {
		cfg.ServerName = fileCfg.ServerName
	}
	if fileCfg.Transport != "" {
		cfg.Transport = fileCfg.Transport
	}
	if fileCfg.HTTPHost != "" {
		cfg.HTTPHost = fileCfg.HTTPHost
	}
	if fileCfg.HTTPPort != nil && *fileCfg.HTTPPort > 0 && *fileCfg.HTTPPort <= 65535 {
		cfg.HTTPPort = *fileCfg.HTTPPort
	}
	if fileCfg.LogLevel != "" {
		cfg.LogLevel = fileCfg.LogLevel
	}
	if fileCfg.AuthMode != "" {
		cfg.AuthMode = fileCfg.AuthMode
	}
	if fileCfg.LazyLoading != nil {
		cfg.LazyLoading = *fileCfg.LazyLoading
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AUTOTASK_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("AUTOTASK_SECRET"); v != "" {
		cfg.Secret = v
	}
	if v := os.Getenv("AUTOTASK_INTEGRATION_CODE"); v != "" {
		cfg.IntegrationCode = v
	}
	if v := os.Getenv("AUTOTASK_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("MCP_SERVER_NAME"); v != "" {
		cfg.ServerName = v
	}
	if v := os.Getenv("MCP_TRANSPORT"); v != "" {
		cfg.Transport = v
	}
	if v := os.Getenv("MCP_HTTP_HOST"); v != "" {
		cfg.HTTPHost = v
	}
	if v := os.Getenv("MCP_HTTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			cfg.HTTPPort = p
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("AUTH_MODE"); v != "" {
		cfg.AuthMode = v
	}
	if v := os.Getenv("LAZY_LOADING"); v != "" {
		cfg.LazyLoading = strings.ToLower(v) == "true" || v == "1"
	}
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= 4 {
		return "****"
	}
	return string(runes[:2]) + strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-2:])
}

// handleConfigCommand handles CLI subcommands for 'autotask-mcp config ...'.
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: autotask-mcp config <path|get|set|unset> [arguments...]")
	}

	cfgPath := defaultConfigPath()

	switch args[0] {
	case "path":
		fmt.Println(cfgPath)
		return nil
	case "get":
		return handleConfigGet(cfgPath, args[1:])
	case "set":
		return handleConfigSet(cfgPath, args[1:])
	case "unset":
		return handleConfigUnset(cfgPath, args[1:])
	default:
		return fmt.Errorf("unknown config subcommand %q; expected path, get, set, or unset", args[0])
	}
}

func handleConfigGet(cfgPath string, args []string) error {
	fc, loaded, err := loadFileConfig(cfgPath)
	if err != nil {
		return err
	}
	if !loaded {
		fmt.Printf("Config file does not exist at %s\n", cfgPath)
		return nil
	}

	if len(args) > 0 {
		val, err := getConfigField(fc, args[0])
		if err != nil {
			return err
		}
		if val != "" {
			fmt.Println(val)
		}
		return nil
	}

	display := fc
	if display.Secret != "" {
		display.Secret = maskSecret(display.Secret)
	}
	data, err := json.MarshalIndent(display, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Config from %s:\n%s\n", cfgPath, string(data))
	return nil
}

func getConfigField(fc FileConfig, key string) (string, error) {
	fields := map[string]string{
		"username":         fc.Username,
		"secret":           maskSecret(fc.Secret),
		"integration_code": fc.IntegrationCode,
		"integrationcode":  fc.IntegrationCode,
		"api_url":          fc.APIURL,
		"apiurl":           fc.APIURL,
		"server_name":      fc.ServerName,
		"servername":       fc.ServerName,
		"transport":        fc.Transport,
		"http_host":        fc.HTTPHost,
		"httphost":         fc.HTTPHost,
		"log_level":        fc.LogLevel,
		"loglevel":         fc.LogLevel,
		"auth_mode":        fc.AuthMode,
		"authmode":         fc.AuthMode,
	}
	// Register the pointer-backed keys unconditionally so an UNSET value returns
	// empty (like every string field) rather than a false "unknown config key".
	fields["http_port"], fields["httpport"] = "", ""
	if fc.HTTPPort != nil {
		fields["http_port"] = strconv.Itoa(*fc.HTTPPort)
		fields["httpport"] = strconv.Itoa(*fc.HTTPPort)
	}
	fields["lazy_loading"], fields["lazyloading"] = "", ""
	if fc.LazyLoading != nil {
		fields["lazy_loading"] = strconv.FormatBool(*fc.LazyLoading)
		fields["lazyloading"] = strconv.FormatBool(*fc.LazyLoading)
	}

	val, ok := fields[strings.ToLower(key)]
	if !ok {
		return "", fmt.Errorf("unknown config key %q", key)
	}
	return val, nil
}

func handleConfigSet(cfgPath string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: autotask-mcp config set <key> <value>")
	}
	key, val := strings.ToLower(args[0]), args[1]

	fc, _, err := loadFileConfig(cfgPath)
	if err != nil {
		return err
	}

	if err := setConfigField(&fc, key, val); err != nil {
		return err
	}

	if err := saveFileConfig(cfgPath, fc); err != nil {
		return err
	}
	fmt.Printf("Successfully updated %s in %s\n", key, cfgPath)
	return nil
}

func setConfigField(fc *FileConfig, key, val string) error {
	switch strings.ToLower(key) {
	case "username":
		fc.Username = val
	case "secret":
		fc.Secret = val
	case "integration_code", "integrationcode":
		fc.IntegrationCode = val
	case "api_url", "apiurl":
		fc.APIURL = val
	case "server_name", "servername":
		fc.ServerName = val
	case "transport":
		fc.Transport = val
	case "http_host", "httphost":
		fc.HTTPHost = val
	case "log_level", "loglevel":
		fc.LogLevel = val
	case "auth_mode", "authmode":
		fc.AuthMode = val
	case "http_port", "httpport":
		if val == "" {
			fc.HTTPPort = nil
			return nil
		}
		p, err := strconv.Atoi(val)
		if err != nil || p <= 0 || p > 65535 {
			return fmt.Errorf("invalid port %q: must be integer between 1 and 65535", val)
		}
		fc.HTTPPort = &p
	case "lazy_loading", "lazyloading":
		if val == "" {
			fc.LazyLoading = nil
			return nil
		}
		switch strings.ToLower(val) {
		case "true", "1":
			b := true
			fc.LazyLoading = &b
		case "false", "0":
			b := false
			fc.LazyLoading = &b
		default:
			return fmt.Errorf("invalid boolean value %q for %s: must be true or false", val, key)
		}
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

func handleConfigUnset(cfgPath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: autotask-mcp config unset <key>")
	}
	key := strings.ToLower(args[0])

	fc, loaded, err := loadFileConfig(cfgPath)
	if err != nil {
		return err
	}
	if !loaded {
		fmt.Printf("Config file does not exist at %s\n", cfgPath)
		return nil
	}

	if err := setConfigField(&fc, key, ""); err != nil {
		return err
	}

	if err := saveFileConfig(cfgPath, fc); err != nil {
		return err
	}
	fmt.Printf("Successfully unset %s in %s\n", key, cfgPath)
	return nil
}
