package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/autotask-mcp/resources"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/autotask-mcp/tools"
)

const serverInstructions = "Autotask PSA MCP Server. Provides tools for managing tickets, companies, contacts, projects, time entries, billing, and more. Use autotask_search_* tools to find entities, autotask_get_* for details, and autotask_create_*/autotask_update_* for modifications. Use picklist tools to discover valid field values."

// buildServer creates and configures an MCP server with all tool handlers registered.
// When lazyLoading is true, only 4 meta-tools are registered for progressive discovery.
func buildServer(client *autotask.Client, lazyLoading bool) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "autotask-mcp", Version: version},
		&mcp.ServerOptions{Instructions: serverInstructions},
	)

	if lazyLoading {
		tools.RegisterLazyTools(s)
	} else {
		mapper := services.NewMappingCache(client)
		picklist := services.NewPicklistCache(client)
		tools.RegisterAll(s, client, mapper, picklist)
	}
	resources.RegisterAll(s, client)

	return s
}

// run is the main entry point for the server. It replaces the stub in main.go.
func run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	logger.Info("autotask-mcp starting", "version", version, "transport", cfg.Transport)

	switch cfg.Transport {
	case "stdio":
		return runStdio(ctx, cfg, logger)
	case "http":
		return runHTTP(ctx, cfg, logger)
	default:
		return fmt.Errorf("unknown transport %q: expected \"stdio\" or \"http\"", cfg.Transport)
	}
}

// runStdio starts the MCP server on stdin/stdout.
func runStdio(ctx context.Context, cfg Config, logger *slog.Logger) error {
	if cfg.Username == "" || cfg.Secret == "" || cfg.IntegrationCode == "" {
		return fmt.Errorf("missing Autotask credentials: set AUTOTASK_USERNAME, AUTOTASK_SECRET, and AUTOTASK_INTEGRATION_CODE")
	}

	authCfg := autotask.AuthConfig{
		Username:        cfg.Username,
		Secret:          cfg.Secret,
		IntegrationCode: cfg.IntegrationCode,
	}

	clientOpts := []autotask.ClientOption{
		autotask.WithLogger(logger),
		autotask.WithRateLimiter(),
		autotask.WithCircuitBreaker(),
	}
	if cfg.APIURL != "" {
		clientOpts = append(clientOpts, autotask.WithBaseURL(cfg.APIURL))
	}

	client, err := autotask.NewClient(ctx, authCfg, clientOpts...)
	if err != nil {
		return fmt.Errorf("creating autotask client: %w", err)
	}
	defer client.Close() //nolint:errcheck

	s := buildServer(client, cfg.LazyLoading)
	logger.Info("autotask-mcp ready", "transport", "stdio", "lazyLoading", cfg.LazyLoading)
	return s.Run(ctx, &mcp.StdioTransport{})
}

// runHTTP starts the MCP server over HTTP with streamable transport.
func runHTTP(ctx context.Context, cfg Config, logger *slog.Logger) error {
	var sharedClient *autotask.Client

	if cfg.AuthMode == "env" {
		// Validate credentials
		if cfg.Username == "" || cfg.Secret == "" || cfg.IntegrationCode == "" {
			return fmt.Errorf("missing Autotask credentials for env auth mode: set AUTOTASK_USERNAME, AUTOTASK_SECRET, and AUTOTASK_INTEGRATION_CODE")
		}

		authCfg := autotask.AuthConfig{
			Username:        cfg.Username,
			Secret:          cfg.Secret,
			IntegrationCode: cfg.IntegrationCode,
		}
		clientOpts := []autotask.ClientOption{
			autotask.WithLogger(logger),
			autotask.WithRateLimiter(),
			autotask.WithCircuitBreaker(),
		}
		if cfg.APIURL != "" {
			clientOpts = append(clientOpts, autotask.WithBaseURL(cfg.APIURL))
		}

		var err error
		sharedClient, err = autotask.NewClient(ctx, authCfg, clientOpts...)
		if err != nil {
			return fmt.Errorf("creating autotask client: %w", err)
		}
		defer sharedClient.Close() //nolint:errcheck
	}

	// Factory function returns an *mcp.Server for each request.
	getServer := func(r *http.Request) *mcp.Server {
		if cfg.AuthMode == "env" {
			return buildServer(sharedClient, cfg.LazyLoading)
		}

		// Gateway mode: extract credentials from request headers.
		apiKey := r.Header.Get("X-API-Key")
		apiSecret := r.Header.Get("X-API-Secret")
		integrationCode := r.Header.Get("X-Integration-Code")
		if apiKey == "" || apiSecret == "" || integrationCode == "" {
			return nil
		}

		authCfg := autotask.AuthConfig{
			Username:        apiKey,
			Secret:          apiSecret,
			IntegrationCode: integrationCode,
		}
		// Note: WithRateLimiter and WithCircuitBreaker are middleware wrappers with
		// no background goroutines; they do not require explicit cleanup via Close().
		// We deliberately omit WithThresholdMonitor here to avoid background goroutine
		// leaks on per-session clients. If background monitoring is needed, a
		// session-lifecycle hook from the MCP SDK would be required.
		// TODO: Close per-session clients when session ends once the SDK exposes
		//       a session-end callback.
		clientOpts := []autotask.ClientOption{
			autotask.WithLogger(logger),
			autotask.WithRateLimiter(),
			autotask.WithCircuitBreaker(),
		}
		if cfg.APIURL != "" {
			clientOpts = append(clientOpts, autotask.WithBaseURL(cfg.APIURL))
		}

		client, err := autotask.NewClient(r.Context(), authCfg, clientOpts...)
		if err != nil {
			logger.Error("failed to create autotask client for request", "error", err)
			return nil
		}
		return buildServer(client, cfg.LazyLoading)
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
		Logger: logger,
	})

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"transport": "http",
			"authMode":  cfg.AuthMode,
			"version":   version,
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown: watch for context cancellation.
	go func() {
		<-ctx.Done()
		_ = httpServer.Close()
	}()

	logger.Info("autotask-mcp HTTP server listening", "addr", addr, "authMode", cfg.AuthMode)
	if err := httpServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}
