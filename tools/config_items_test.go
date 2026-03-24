package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/autotask-mcp/services"
)

// TestRegisterConfigItemTools_NoPanic verifies registration does not panic.
func TestRegisterConfigItemTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterConfigItemTools(s, client, mapper)
}

// TestSearchConfigurationItemsHandler_NoResults tests the empty-result case.
func TestSearchConfigurationItemsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	handler := searchConfigurationItemsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchConfigurationItemsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestSearchConfigurationItemsHandler_WithResults tests that seeded CIs are returned.
func TestSearchConfigurationItemsHandler_WithResults(t *testing.T) {
	ci := autotasktest.ConfigurationItemFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ci))
	mapper := services.NewMappingCache(client)
	handler := searchConfigurationItemsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchConfigurationItemsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	var resp map[string]any
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatalf("expected 'items' array in response, got: %v", resp)
	}
	if len(items) == 0 {
		t.Error("expected at least one configuration item in results")
	}
}

// TestSearchConfigurationItemsHandler_WithFilters verifies filters can be applied.
func TestSearchConfigurationItemsHandler_WithFilters(t *testing.T) {
	ci := autotasktest.ConfigurationItemFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ci))
	mapper := services.NewMappingCache(client)
	handler := searchConfigurationItemsHandler(client, mapper)
	ctx := context.Background()

	active := true
	result, _, err := handler(ctx, nil, SearchConfigurationItemsInput{
		SearchTerm: "PROD",
		CompanyID:  1001,
		IsActive:   &active,
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
