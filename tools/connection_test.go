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

// TestTestConnectionHandler_Success tests that connection succeeds against the mock server.
func TestTestConnectionHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := testConnectionHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if text.Text == "" {
		t.Error("expected non-empty connection result text")
	}
}
