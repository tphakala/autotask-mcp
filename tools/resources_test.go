package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterResourceTools_NoPanic verifies that RegisterResourceTools registers
// the tool without panicking.
func TestRegisterResourceTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterResourceTools(s, client)
}

// TestSearchResourcesHandler_ReturnsNoResourcesFound tests the empty-result case over wire.
func TestSearchResourcesHandler_ReturnsNoResourcesFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_resources",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned resources, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchResourcesHandler_ReturnsResources tests that seeded resources are returned over wire.
func TestSearchResourcesHandler_ReturnsResources(t *testing.T) {
	resource := autotasktest.ResourceFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(resource))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_resources",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 returned resource, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected resource to contain 'id' field")
	}
}

// TestSearchResourcesHandler_WithFilters verifies that filters can be applied over wire.
func TestSearchResourcesHandler_WithFilters(t *testing.T) {
	resource := autotasktest.ResourceFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(resource))
	ctx := context.Background()

	active := true
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_resources",
		Arguments: map[string]any{
			"searchTerm":   "John",
			"isActive":     active,
			"resourceType": "Employee",
			"maxResults":   10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

