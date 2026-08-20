package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterConfigItemTools_NoPanic verifies registration does not panic.
func TestRegisterConfigItemTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterConfigItemTools(s, client, mapper)
}

// TestSearchConfigurationItemsHandler_NoResults tests the empty-result case over wire.
func TestSearchConfigurationItemsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_configuration_items",
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
		t.Errorf("expected 0 returned items, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchConfigurationItemsHandler_WithResults tests that seeded CIs are returned over wire.
func TestSearchConfigurationItemsHandler_WithResults(t *testing.T) {
	ci := autotasktest.ConfigurationItemFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(ci))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_configuration_items",
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
		t.Errorf("expected at least 1 item, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected CI to contain 'id' field")
	}
}

// TestSearchConfigurationItemsHandler_WithFilters verifies filters can be applied over wire.
func TestSearchConfigurationItemsHandler_WithFilters(t *testing.T) {
	ci := autotasktest.ConfigurationItemFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(ci))
	ctx := context.Background()

	active := true
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_configuration_items",
		Arguments: map[string]any{
			"searchTerm": "PROD",
			"companyID":  1001,
			"isActive":   active,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}
