package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterConnectionTools_NoPanic verifies registration does not panic.
func TestRegisterConnectionTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterConnectionTools(s, client)
}

// TestTestConnectionHandler_Direct tests handler logic directly.
func TestTestConnectionHandler_Direct(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := testConnectionHandler(client)
	ctx := context.Background()

	res, out, err := handler(ctx, nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil *CallToolResult for SDK inference, got %v", res)
	}
	if !out.Success || out.Entity != "Tickets" {
		t.Errorf("unexpected output: %+v", out)
	}
}

// TestTestConnectionHandler_Wire tests that connection succeeds against the mock server over wire protocol.
func TestTestConnectionHandler_Wire(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_test_connection",
	})
	if err != nil {
		t.Fatalf("wire call failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	out := parseStructuredContent[ConnectionStatusOut](t, result)
	if !out.Success {
		t.Errorf("expected Success=true, got %v", out.Success)
	}
	if out.Entity != "Tickets" {
		t.Errorf("expected Entity=Tickets, got %q", out.Entity)
	}
}
