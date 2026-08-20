package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
)

func TestRunDoctor_MissingCredentials(t *testing.T) {
	cfg := Config{}
	var buf bytes.Buffer

	err := runDoctor(context.Background(), cfg, &buf)
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}

	out := buf.String()
	if !strings.Contains(out, "ERROR: Missing required Autotask credentials") {
		t.Errorf("expected error message in output, got: %s", out)
	}
	if !strings.Contains(out, "autotask-mcp config set") {
		t.Errorf("expected config set advice in output, got: %s", out)
	}
}

func TestRunDoctor_SuccessWithMockServer(t *testing.T) {
	username := "api-user@example.com"
	secret := "api-secret-123"
	code := "INTEG-CODE"

	ts, _ := autotasktest.NewServer(t, autotasktest.WithAuth(username, secret, code))

	cfg := Config{
		Username:        username,
		Secret:          secret,
		IntegrationCode: code,
		APIURL:          ts.URL,
		Transport:       "stdio",
		AuthMode:        "env",
	}

	var buf bytes.Buffer
	err := runDoctor(context.Background(), cfg, &buf)
	if err != nil {
		t.Fatalf("runDoctor failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Connection & Authentication: SUCCESS") {
		t.Errorf("expected success message in output, got: %s", out)
	}
	if !strings.Contains(out, "Tickets") || !strings.Contains(out, "Companies") {
		t.Errorf("expected entity list in output, got: %s", out)
	}
	if !strings.Contains(out, "Doctor check completed successfully") {
		t.Errorf("expected completion message in output, got: %s", out)
	}
}

func TestPrintActionableStartupError(t *testing.T) {
	// Should not panic
	printActionableStartupError(context.Canceled, Config{})
	printActionableStartupError(context.Canceled, Config{
		Username:        "u",
		Secret:          "s",
		IntegrationCode: "c",
	})
}
