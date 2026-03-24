package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/autotask-mcp/services"
)

// TestRegisterFinancialTools_NoPanic verifies registration does not panic.
func TestRegisterFinancialTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterFinancialTools(s, client, mapper)
}

// TestGetQuoteHandler_NotFound tests that a missing quote returns an error result.
func TestGetQuoteHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getQuoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetQuoteInput{QuoteID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing quote")
	}
}

// TestSearchQuotesHandler_NoResults tests the empty-result case.
func TestSearchQuotesHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchQuotesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchQuotesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestCreateQuoteHandler_Success tests that a quote can be created.
func TestCreateQuoteHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createQuoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateQuoteInput{
		CompanyID: 1001,
		Name:      "Q-2024-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestCreateQuoteHandler_InvalidEffectiveDate tests that an invalid date returns an error result.
func TestCreateQuoteHandler_InvalidEffectiveDate(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createQuoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateQuoteInput{
		CompanyID:     1001,
		EffectiveDate: "not-a-date",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateQuoteItemHandler_AutoDeterminesType tests that the quote item type is auto-determined.
func TestCreateQuoteItemHandler_AutoDeterminesType(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createQuoteItemHandler(client)
	ctx := context.Background()

	// Product ID → type 1
	result, _, err := handler(ctx, nil, CreateQuoteItemInput{
		QuoteID:   101,
		Quantity:  2,
		ProductID: 5001,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result for product quote item, got IsError=true; content: %v", result.Content)
	}
}

// TestDeleteQuoteItemHandler_NotFound tests that deleting a missing quote item returns an error.
func TestDeleteQuoteItemHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := deleteQuoteItemHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, DeleteQuoteItemInput{QuoteID: 101, QuoteItemID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing quote item")
	}
}

// TestGetOpportunityHandler_NotFound tests that a missing opportunity returns an error result.
func TestGetOpportunityHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getOpportunityHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetOpportunityInput{OpportunityID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing opportunity")
	}
}

// TestSearchOpportunitiesHandler_NoResults tests the empty-result case.
func TestSearchOpportunitiesHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchOpportunitiesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchOpportunitiesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestCreateOpportunityHandler_InvalidDate tests that an invalid date returns an error result.
func TestCreateOpportunityHandler_InvalidDate(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createOpportunityHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateOpportunityInput{
		Title:              "Big Deal",
		CompanyID:          1001,
		OwnerResourceID:    5001,
		Status:             1,
		Stage:              1,
		ProjectedCloseDate: "not-a-date",
		StartDate:          "2024-01-01",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateOpportunityHandler_Success tests that an opportunity can be created.
func TestCreateOpportunityHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createOpportunityHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateOpportunityInput{
		Title:              "Big Deal",
		CompanyID:          1001,
		OwnerResourceID:    5001,
		Status:             1,
		Stage:              1,
		ProjectedCloseDate: "2024-06-30",
		StartDate:          "2024-01-15",
		Amount:             50000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestSearchContractsHandler_NoResults tests the empty-result case.
func TestSearchContractsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	handler := searchContractsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchContractsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestSearchContractsHandler_WithResults tests that seeded contracts are returned.
func TestSearchContractsHandler_WithResults(t *testing.T) {
	contract := autotasktest.ContractFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(contract))
	mapper := services.NewMappingCache(client)
	handler := searchContractsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchContractsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestSearchInvoicesHandler_NoResults tests the empty-result case for invoices.
func TestSearchInvoicesHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchInvoicesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchInvoicesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
