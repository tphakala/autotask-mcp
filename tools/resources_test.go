package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterResourceTools_NoPanic verifies that RegisterResourceTools registers
// the tool without panicking.
func TestRegisterResourceTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterResourceTools(s, client)
}

// TestSearchResourcesHandler_ReturnsNoResourcesFound tests the empty-result case.
func TestSearchResourcesHandler_ReturnsNoResourcesFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)

	handler := searchResourcesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchResourcesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if text.Text != "No resources found" {
		t.Errorf("expected 'No resources found', got %q", text.Text)
	}
}

// TestSearchResourcesHandler_ReturnsResources tests that seeded resources are returned.
func TestSearchResourcesHandler_ReturnsResources(t *testing.T) {
	resource := autotasktest.ResourceFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(resource))

	handler := searchResourcesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchResourcesInput{})
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

	// Result should be a compact JSON response with summary + items.
	var resp map[string]any
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}
	items, ok := resp["items"].([]any)
	if !ok || len(items) == 0 {
		t.Error("expected at least one resource in results")
	}
}

// TestSearchResourcesHandler_WithFilters verifies that filters can be applied.
func TestSearchResourcesHandler_WithFilters(t *testing.T) {
	resource := autotasktest.ResourceFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(resource))

	handler := searchResourcesHandler(client)
	ctx := context.Background()

	active := true
	in := SearchResourcesInput{
		SearchTerm:   "John",
		IsActive:     &active,
		ResourceType: 1,
		Page:         1,
		PageSize:     10,
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	_ = result
}
