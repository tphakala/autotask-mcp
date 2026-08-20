package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const version = "1.1.0"

func printUsage() {
	fmt.Printf("autotask-mcp v%s - Model Context Protocol server for Autotask PSA\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  autotask-mcp [command|flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  serve               Start the MCP server (default)")
	fmt.Println("  doctor              Run pre-flight connectivity and permissions diagnostics")
	fmt.Println("  config              Manage file configuration ($XDG_CONFIG_HOME/autotask-mcp/config.json)")
	fmt.Println("    config path       Print the active config file path")
	fmt.Println("    config get [key]  Get a config value or view current configuration")
	fmt.Println("    config set <k> <v> Set a configuration key securely (0600)")
	fmt.Println("    config unset <k>  Remove a configuration key")
	fmt.Println("  version             Print version")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --doctor            Run pre-flight diagnostics (same as 'doctor')")
	fmt.Println("  --version, -v       Print version")
	fmt.Println("  --help, -h          Print this help message")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("autotask-mcp v%s\n", version)
			return
		case "help", "--help", "-h":
			printUsage()
			return
		case "doctor", "--doctor":
			cfg := loadConfig()
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			if err := runDoctor(ctx, cfg, os.Stdout); err != nil {
				os.Exit(1)
			}
			return
		case "config":
			if err := handleConfigCommand(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "autotask-mcp config error:", err)
				os.Exit(1)
			}
			return
		case "serve":
			// continue to serve
		default:
			if strings.HasPrefix(os.Args[1], "-") {
				fmt.Fprintf(os.Stderr, "unknown flag %q; run with --help for usage\n", os.Args[1])
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "unknown command %q; run with --help for usage\n", os.Args[1])
			os.Exit(1)
		}
	}

	cfg := loadConfig()

	// Set up structured logging to stderr (protects stdio JSON-RPC channel)
	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, cfg, logger); err != nil {
		printActionableStartupError(err, cfg)
		os.Exit(1)
	}
}
