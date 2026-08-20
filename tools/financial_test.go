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

// TestRegisterFinancialTools_NoPanic verifies registration does not panic.
func TestRegisterFinancialTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterFinancialTools(s, client, mapper)
}

// TestGetQuoteHandler_NotFound tests that a missing quote returns an error result over wire.
func TestGetQuoteHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_quote",
		Arguments: map[string]any{
			"quoteId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing quote")
	}
}

// TestGetQuoteHandler_Success tests retrieving a seeded quote over wire.
func TestGetQuoteHandler_Success(t *testing.T) {
	quote := &entities.Quote{
		ID:        autotask.Set(int64(6001)),
		CompanyID: autotask.Set(int64(1001)),
		Name:      autotask.Set("Standard Quote"),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(quote))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_quote",
		Arguments: map[string]any{
			"quoteId": 6001,
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
	if !strings.Contains(nameStr, "Quote") {
		t.Errorf("expected name to contain 'Quote', got %v", m["name"])
	}
}

// TestSearchQuotesHandler_NoResults tests the empty-result case over wire.
func TestSearchQuotesHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_quotes",
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
		t.Errorf("expected 0 returned quotes, got %d", resp.Summary.Returned)
	}
}

// TestCreateQuoteHandler_Success tests creating a quote over wire.
func TestCreateQuoteHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_quote",
		Arguments: map[string]any{
			"companyId": 1001,
			"name":      "Q-2024-001",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["name"].(string)
	if !strings.Contains(nameStr, "Q-2024-001") {
		t.Errorf("expected name to contain 'Q-2024-001', got %v", m["name"])
	}
}

// TestCreateQuoteHandler_InvalidEffectiveDate tests that an invalid date returns an error result over wire.
func TestCreateQuoteHandler_InvalidEffectiveDate(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_quote",
		Arguments: map[string]any{
			"companyId":     1001,
			"effectiveDate": "not-a-date",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateQuoteItemHandler_AutoDeterminesType tests creating a quote item over wire.
func TestCreateQuoteItemHandler_AutoDeterminesType(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_quote_item",
		Arguments: map[string]any{
			"quoteId":   101,
			"quantity":  2,
			"productID": 5001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result for product quote item, got IsError=true; content: %v", result.Content)
	}
}

// TestDeleteQuoteItemHandler_NotFound tests that deleting a missing quote item returns an error over wire.
func TestDeleteQuoteItemHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_delete_quote_item",
		Arguments: map[string]any{
			"quoteId":     101,
			"quoteItemId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing quote item")
	}
}

// TestGetOpportunityHandler_NotFound tests that a missing opportunity returns an error result over wire.
func TestGetOpportunityHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_opportunity",
		Arguments: map[string]any{
			"opportunityId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing opportunity")
	}
}

// TestSearchOpportunitiesHandler_NoResults tests the empty-result case over wire.
func TestSearchOpportunitiesHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_opportunities",
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
		t.Errorf("expected 0 returned opportunities, got %d", resp.Summary.Returned)
	}
}

// TestCreateOpportunityHandler_InvalidDate tests that an invalid date returns an error result over wire.
func TestCreateOpportunityHandler_InvalidDate(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_opportunity",
		Arguments: map[string]any{
			"title":              "Big Deal",
			"companyId":          1001,
			"ownerResourceId":    5001,
			"status":             1,
			"stage":              1,
			"projectedCloseDate": "not-a-date",
			"startDate":          "2024-01-01",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateOpportunityHandler_Success tests creating an opportunity over wire.
func TestCreateOpportunityHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_opportunity",
		Arguments: map[string]any{
			"title":              "Big Deal",
			"companyId":          1001,
			"ownerResourceId":    5001,
			"status":             1,
			"stage":              1,
			"projectedCloseDate": "2024-06-30",
			"startDate":          "2024-01-15",
			"amount":             50000,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	titleStr, _ := m["title"].(string)
	if !strings.Contains(titleStr, "Big Deal") {
		t.Errorf("expected title to contain 'Big Deal', got %v", m["title"])
	}
}

// TestSearchContractsHandler_NoResults tests the empty-result case over wire.
func TestSearchContractsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_contracts",
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
		t.Errorf("expected 0 returned contracts, got %d", resp.Summary.Returned)
	}
}

// TestSearchContractsHandler_WithResults tests that seeded contracts are returned over wire.
func TestSearchContractsHandler_WithResults(t *testing.T) {
	contract := autotasktest.ContractFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(contract))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_contracts",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 returned contract, got %d", resp.Summary.Returned)
	}
}

// TestSearchInvoicesHandler_NoResults tests the empty-result case for invoices over wire.
func TestSearchInvoicesHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_invoices",
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
		t.Errorf("expected 0 returned invoices, got %d", resp.Summary.Returned)
	}
}

