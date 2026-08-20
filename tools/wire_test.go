package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
)

// setupWireTest sets up a full in-memory MCP server with all tools registered and connected to an MCP client.
func setupWireTest(t *testing.T, opts ...autotasktest.ServerOption) (*mcp.ClientSession, *autotask.Client) {
	t.Helper()
	ctx := context.Background()

	_, client := autotasktest.NewServer(t, opts...)
	mapper := services.NewMappingCache(client)
	picklist := services.NewPicklistCache(client)

	server := mcp.NewServer(&mcp.Implementation{Name: "autotask-mcp-test", Version: "v0.0.1"}, nil)
	RegisterAll(server, client, mapper, picklist)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession, client
}

// setupLazyWireTest sets up a lazy-loading in-memory MCP server with meta-tools registered.
func setupLazyWireTest(t *testing.T, opts ...autotasktest.ServerOption) (*mcp.ClientSession, *autotask.Client) {
	t.Helper()
	ctx := context.Background()

	_, client := autotasktest.NewServer(t, opts...)
	mapper := services.NewMappingCache(client)
	picklist := services.NewPicklistCache(client)

	server := mcp.NewServer(&mcp.Implementation{Name: "autotask-lazy-test", Version: "v0.0.1"}, nil)
	RegisterLazyTools(server, client, mapper, picklist)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession, client
}

// parseStructuredContent decodes the structured content from a CallToolResult into target struct.
func parseStructuredContent[T any](t *testing.T, result *mcp.CallToolResult) T {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil CallToolResult")
	}
	if result.StructuredContent == nil {
		t.Fatalf("expected non-nil StructuredContent in result: %+v", result)
	}
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("failed to marshal StructuredContent: %v", err)
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to unmarshal StructuredContent into %T: %v", out, err)
	}
	return out
}
