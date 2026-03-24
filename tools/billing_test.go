package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/autotask-mcp/services"
)

// TestRegisterBillingTools_NoPanic verifies registration does not panic.
func TestRegisterBillingTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterBillingTools(s, client, mapper)
}

// TestGetBillingItemHandler_NotFound tests that a missing billing item returns an error result.
func TestGetBillingItemHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getBillingItemHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetBillingItemInput{BillingItemID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing billing item")
	}
}

// TestSearchBillingItemsHandler_NoResults tests the empty-result case.
func TestSearchBillingItemsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	handler := searchBillingItemsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchBillingItemsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestSearchBillingItemApprovalLevelsHandler_NoResults tests the empty-result case.
func TestSearchBillingItemApprovalLevelsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchBillingItemApprovalLevelsHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchBillingItemApprovalLevelsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
