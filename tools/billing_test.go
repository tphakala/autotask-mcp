package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/entities"
)

// TestRegisterBillingTools_NoPanic verifies registration does not panic.
func TestRegisterBillingTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterBillingTools(s, client, mapper)
}

// TestGetBillingItemHandler_NotFound tests that a missing billing item returns an error result over wire.
func TestGetBillingItemHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_billing_item",
		Arguments: map[string]any{
			"billingItemId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing billing item")
	}
}

// TestSearchBillingItemsHandler_NoResults tests the empty-result case over wire.
func TestSearchBillingItemsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_billing_items",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected IsError=false for no-results case; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned billing items, got %d", resp.Summary.Returned)
	}
}

// TestSearchBillingItemApprovalLevelsHandler_NoResults tests the empty-result case over wire.
func TestSearchBillingItemApprovalLevelsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_billing_item_approval_levels",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected IsError=false for no-results case; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned approval levels, got %d", resp.Summary.Returned)
	}
}

// TestGetBillingItemHandler_MissingId tests that billingItemId=0 returns a validation error.
func TestGetBillingItemHandler_MissingId(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_billing_item",
		Arguments: map[string]any{
			"billingItemId": 0,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true when billingItemId is 0")
	}
}

// TestGetBillingItemHandler_Success tests retrieving a seeded billing item over wire.
func TestGetBillingItemHandler_Success(t *testing.T) {
	item := &entities.BillingItem{
		ID:          autotask.Set(int64(9001)),
		CompanyID:   autotask.Set(int64(1001)),
		Description: autotask.Set("Monthly Support"),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(item))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_billing_item",
		Arguments: map[string]any{
			"billingItemId": 9001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	m := parseStructuredContent[map[string]any](t, result)
	idVal, ok := m["id"]
	if !ok {
		t.Fatal("expected 'id' field in billing item response")
	}
	fVal, isFloat := idVal.(float64)
	if !isFloat || int64(fVal) != 9001 {
		t.Errorf("expected id=9001, got %v", idVal)
	}
}
