package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/metadata"
)

// runDoctor executes comprehensive configuration, connectivity, and entity permission checks.
func runDoctor(ctx context.Context, cfg Config, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	_, _ = fmt.Fprintln(out, "============================================================")
	_, _ = fmt.Fprintln(out, " autotask-mcp doctor: system and connectivity diagnostics   ")
	_, _ = fmt.Fprintln(out, "============================================================")
	_, _ = fmt.Fprintln(out)

	// Step 1: Configuration Diagnostic
	if err := checkConfigDoctor(cfg, out); err != nil {
		return err
	}

	// Step 2: API Connectivity and Zone Resolution
	client, ticketInfo, err := checkConnectivityDoctor(ctx, cfg, out)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck

	// Step 3: Core Entity Permissions Diagnostic
	allPermsOk := checkPermissionsDoctor(ctx, client, ticketInfo, out)

	if allPermsOk {
		_, _ = fmt.Fprintln(out, "============================================================")
		_, _ = fmt.Fprintln(out, " Doctor check completed successfully. Server is ready.     ")
		_, _ = fmt.Fprintln(out, "============================================================")
	} else {
		_, _ = fmt.Fprintln(out, "============================================================")
		_, _ = fmt.Fprintln(out, " Doctor check completed with warnings. Review entity checks. ")
		_, _ = fmt.Fprintln(out, "============================================================")
	}
	return nil
}

func checkConfigDoctor(cfg Config, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "[1/3] Configuration & Credentials Check")
	_, _ = fmt.Fprintln(out, "------------------------------------------------------------")

	cfgPath := defaultConfigPath()
	_, _ = fmt.Fprintf(out, "Config File Path: %s\n", cfgPath)
	if cfg.ConfigFile != "" {
		_, _ = fmt.Fprintln(out, "Config File Status: Loaded successfully")
		if info, err := os.Stat(cfg.ConfigFile); err == nil {
			perm := info.Mode().Perm()
			if perm&0077 == 0 {
				_, _ = fmt.Fprintf(out, "File Permissions:   0%o (Secure - owner read/write only)\n", perm)
			} else {
				_, _ = fmt.Fprintf(out, "File Permissions:   0%o (WARNING: Recommend 0600)\n", perm)
			}
		}
	} else {
		_, _ = fmt.Fprintln(out, "Config File Status: Not present (using env vars / defaults)")
	}

	printCredStatus := func(name, val, envKey string) {
		if val == "" {
			_, _ = fmt.Fprintf(out, "  - %-24s [MISSING]\n", name+":")
		} else if os.Getenv(envKey) != "" {
			_, _ = fmt.Fprintf(out, "  - %-24s [OK] (from env $%s)\n", name+":", envKey)
		} else {
			_, _ = fmt.Fprintf(out, "  - %-24s [OK] (from config file)\n", name+":")
		}
	}

	printCredStatus("AUTOTASK_USERNAME", cfg.Username, "AUTOTASK_USERNAME")
	printCredStatus("AUTOTASK_SECRET", cfg.Secret, "AUTOTASK_SECRET")
	printCredStatus("AUTOTASK_INTEGRATION_CODE", cfg.IntegrationCode, "AUTOTASK_INTEGRATION_CODE")
	if cfg.APIURL != "" {
		_, _ = fmt.Fprintf(out, "  - %-24s %s\n", "AUTOTASK_API_URL (override):", cfg.APIURL)
	}
	_, _ = fmt.Fprintf(out, "  - %-24s %s\n", "Transport:", cfg.Transport)
	_, _ = fmt.Fprintf(out, "  - %-24s %s\n", "Auth Mode:", cfg.AuthMode)
	_, _ = fmt.Fprintf(out, "  - %-24s %v\n", "Lazy Loading:", cfg.LazyLoading)
	_, _ = fmt.Fprintln(out)

	if cfg.Username == "" || cfg.Secret == "" || cfg.IntegrationCode == "" {
		_, _ = fmt.Fprintln(out, "❌ ERROR: Missing required Autotask credentials.")
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, "To configure credentials, you can either:")
		_, _ = fmt.Fprintln(out, "  1. Set environment variables:")
		_, _ = fmt.Fprintln(out, "     export AUTOTASK_USERNAME=\"your-api-user@example.com\"")
		_, _ = fmt.Fprintln(out, "     export AUTOTASK_SECRET=\"your-api-secret\"")
		_, _ = fmt.Fprintln(out, "     export AUTOTASK_INTEGRATION_CODE=\"your-integration-code\"")
		_, _ = fmt.Fprintln(out, "  2. Or save them in the secure XDG configuration file:")
		_, _ = fmt.Fprintln(out, "     autotask-mcp config set username \"your-api-user@example.com\"")
		_, _ = fmt.Fprintln(out, "     autotask-mcp config set secret \"your-api-secret\"")
		_, _ = fmt.Fprintln(out, "     autotask-mcp config set integration_code \"your-integration-code\"")
		_, _ = fmt.Fprintln(out)
		return fmt.Errorf("missing required credentials")
	}
	return nil
}

func checkConnectivityDoctor(ctx context.Context, cfg Config, out io.Writer) (*autotask.Client, *metadata.EntityInfo, error) {
	_, _ = fmt.Fprintln(out, "[2/3] API Connectivity & Authentication Check")
	_, _ = fmt.Fprintln(out, "------------------------------------------------------------")

	authCfg := autotask.AuthConfig{
		Username:        cfg.Username,
		Secret:          cfg.Secret,
		IntegrationCode: cfg.IntegrationCode,
	}
	clientOpts := []autotask.ClientOption{
		autotask.WithMaxConcurrency(2),
		autotask.WithRateLimiter(),
		autotask.WithCircuitBreaker(),
	}
	if cfg.APIURL != "" {
		clientOpts = append(clientOpts, autotask.WithBaseURL(cfg.APIURL))
	}

	connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	startTime := time.Now()
	client, err := autotask.NewClient(connectCtx, authCfg, clientOpts...)
	if err != nil {
		_, _ = fmt.Fprintf(out, "❌ Connection failed: %v\n", err)
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, "Troubleshooting advice:")
		_, _ = fmt.Fprintln(out, "  - Verify that the username and secret match an active API User in Autotask.")
		_, _ = fmt.Fprintln(out, "  - Verify that the tracking identifier / integration code is registered with Datto/Autotask.")
		_, _ = fmt.Fprintln(out, "  - If your network requires a proxy or custom gateway, specify AUTOTASK_API_URL.")
		_, _ = fmt.Fprintln(out)
		return nil, nil, fmt.Errorf("connection failed: %w", err)
	}

	// Verify authentication by fetching Tickets metadata
	ticketInfo, err := metadata.GetEntityInfo(connectCtx, client, "Tickets")
	latency := time.Since(startTime)
	if err != nil {
		_, _ = fmt.Fprintf(out, "❌ Authentication/Query failed: %v\n", err)
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, "Troubleshooting advice:")
		_, _ = fmt.Fprintln(out, "  - The API user authenticated, but failed to fetch entity metadata.")
		_, _ = fmt.Fprintln(out, "  - Check that the API User's Security Level in Autotask has API permissions enabled.")
		_, _ = fmt.Fprintln(out)
		_ = client.Close()
		return nil, nil, fmt.Errorf("metadata query failed: %w", err)
	}

	_, _ = fmt.Fprintf(out, "✓ Connection & Authentication: SUCCESS (Latency: %s)\n", latency.Round(time.Millisecond))
	_, _ = fmt.Fprintln(out)
	return client, ticketInfo, nil
}

func checkPermissionsDoctor(ctx context.Context, client *autotask.Client, ticketInfo *metadata.EntityInfo, out io.Writer) bool {
	_, _ = fmt.Fprintln(out, "[3/3] Core Entity Permissions Check")
	_, _ = fmt.Fprintln(out, "------------------------------------------------------------")
	_, _ = fmt.Fprintf(out, "%-16s %-12s %-12s %-12s %-12s\n", "Entity", "canCreate", "canUpdate", "canQuery", "canDelete")
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 60))

	coreEntities := []string{"Tickets", "Companies", "Contacts", "TimeEntries", "Resources"}
	allPermsOk := true

	for _, entity := range coreEntities {
		var (
			info *metadata.EntityInfo
			err  error
		)
		if entity == "Tickets" && ticketInfo != nil {
			info = ticketInfo
		} else {
			info, err = metadata.GetEntityInfo(ctx, client, entity)
			if err != nil || info == nil {
				_, _ = fmt.Fprintf(out, "%-16s [ERROR: %v]\n", entity, err)
				allPermsOk = false
				continue
			}
		}
		_, _ = fmt.Fprintf(out, "%-16s %-12v %-12v %-12v %-12v\n",
			entity, info.CanCreate, info.CanUpdate, info.CanQuery, info.CanDelete)
	}
	_, _ = fmt.Fprintln(out)

	if !allPermsOk {
		_, _ = fmt.Fprintln(out, "⚠️ Some entity permissions could not be queried.")
	} else {
		_, _ = fmt.Fprintln(out, "✓ All core entity permissions verified successfully.")
	}
	_, _ = fmt.Fprintln(out)
	return allPermsOk
}

// printActionableStartupError writes actionable guidance to stderr when the server fails to start.
func printActionableStartupError(err error, cfg Config) {
	_, _ = fmt.Fprintln(os.Stderr, "autotask-mcp: startup failed:", err)
	_, _ = fmt.Fprintln(os.Stderr)
	if cfg.Username == "" || cfg.Secret == "" || cfg.IntegrationCode == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Actionable Advice:")
		_, _ = fmt.Fprintln(os.Stderr, "  Missing Autotask credentials. Configure them using either:")
		_, _ = fmt.Fprintln(os.Stderr, "    1. Environment variables:")
		_, _ = fmt.Fprintln(os.Stderr, "       export AUTOTASK_USERNAME=\"api-user@example.com\"")
		_, _ = fmt.Fprintln(os.Stderr, "       export AUTOTASK_SECRET=\"api-secret\"")
		_, _ = fmt.Fprintln(os.Stderr, "       export AUTOTASK_INTEGRATION_CODE=\"integration-code\"")
		_, _ = fmt.Fprintln(os.Stderr, "    2. Secure file configuration:")
		_, _ = fmt.Fprintln(os.Stderr, "       autotask-mcp config set username \"api-user@example.com\"")
		_, _ = fmt.Fprintln(os.Stderr, "       autotask-mcp config set secret \"api-secret\"")
		_, _ = fmt.Fprintln(os.Stderr, "       autotask-mcp config set integration_code \"integration-code\"")
		_, _ = fmt.Fprintln(os.Stderr, "  Run 'autotask-mcp --doctor' to diagnose connectivity.")
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Actionable Advice:")
		_, _ = fmt.Fprintln(os.Stderr, "  Run 'autotask-mcp --doctor' to diagnose credentials, zone routing, and permissions.")
	}
	_, _ = fmt.Fprintln(os.Stderr)
}
