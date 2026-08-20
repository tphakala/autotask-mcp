package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/entities"
)

// TestRegisterSalesTools_NoPanic verifies registration does not panic.
func TestRegisterSalesTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSalesTools(s, client)
}

// TestGetProductHandler_NotFound tests that a missing product returns an error result over wire.
func TestGetProductHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_product",
		Arguments: map[string]any{
			"productId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing product")
	}
}

// TestGetProductHandler_Success tests retrieving a seeded product over wire.
func TestGetProductHandler_Success(t *testing.T) {
	prod := &entities.Product{
		ID:          autotask.Set(int64(8001)),
		Name:        autotask.Set("Standard Monitor"),
		Description: autotask.Set("24-inch 1080p Monitor"),
		IsActive:    autotask.Set(true),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(prod))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_product",
		Arguments: map[string]any{
			"productId": 8001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["name"].(string)
	if !strings.Contains(nameStr, "Standard Monitor") {
		t.Errorf("expected name to contain 'Standard Monitor', got %v", m["name"])
	}
}

// TestSearchProductsHandler_NoResults tests the empty-result case over wire.
func TestSearchProductsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_products",
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
		t.Errorf("expected 0 returned products, got %d", resp.Summary.Returned)
	}
}

// TestSearchProductsHandler_WithFilters verifies product filtering over wire.
func TestSearchProductsHandler_WithFilters(t *testing.T) {
	prod := &entities.Product{
		ID:       autotask.Set(int64(8001)),
		Name:     autotask.Set("Standard Monitor"),
		IsActive: autotask.Set(true),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(prod))
	ctx := context.Background()

	active := true
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_products",
		Arguments: map[string]any{
			"searchTerm": "Monitor",
			"isActive":   active,
			"maxResults": 10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 product returned, got %d", resp.Summary.Returned)
	}
}

// TestGetServiceHandler_NotFound tests that a missing service returns an error result over wire.
func TestGetServiceHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_service",
		Arguments: map[string]any{
			"serviceId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing service")
	}
}

// TestSearchServicesHandler_NoResults tests the empty-result case over wire.
func TestSearchServicesHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_services",
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
		t.Errorf("expected 0 returned services, got %d", resp.Summary.Returned)
	}
}

// TestGetServiceBundleHandler_NotFound tests that a missing service bundle returns an error result over wire.
func TestGetServiceBundleHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_service_bundle",
		Arguments: map[string]any{
			"serviceBundleId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing service bundle")
	}
}

// TestSearchServiceBundlesHandler_NoResults tests the empty-result case over wire.
func TestSearchServiceBundlesHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_service_bundles",
		Arguments: map[string]any{
			"searchTerm": "acme",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned bundles, got %d", resp.Summary.Returned)
	}
}

