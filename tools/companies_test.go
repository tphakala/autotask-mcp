package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterCompanyTools_NoPanic verifies that RegisterCompanyTools registers all
// three tools without panicking.
func TestRegisterCompanyTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterCompanyTools(s, client, mapper)
}

// TestSearchCompaniesHandler_ReturnsNoCompaniesFound tests the empty-result case over wire.
func TestSearchCompaniesHandler_ReturnsNoCompaniesFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_companies",
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
		t.Errorf("expected 0 returned companies, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchCompaniesHandler_ReturnsCompanies tests that seeded companies are returned over wire.
func TestSearchCompaniesHandler_ReturnsCompanies(t *testing.T) {
	company := autotasktest.CompanyFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(company))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_companies",
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
		t.Errorf("expected at least 1 returned company, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected company to contain 'id' field")
	}
}

// TestCreateCompanyHandler_Success tests creating a company over wire.
func TestCreateCompanyHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.CompanyFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_company",
		Arguments: map[string]any{
			"companyName": "Test Company",
			"companyType": 1,
			"phone":       "555-1234",
			"city":        "Springfield",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["companyName"].(string)
	if !strings.Contains(nameStr, "Test Company") {
		t.Errorf("expected companyName to contain 'Test Company', got %v", m["companyName"])
	}
}

// TestUpdateCompanyHandler_Success tests updating an existing company over wire.
func TestUpdateCompanyHandler_Success(t *testing.T) {
	company := autotasktest.CompanyFixture()
	companyID, ok := company.ID.Get()
	if !ok {
		t.Fatal("fixture company has no ID")
	}

	cs, _ := setupWireTest(t, autotasktest.WithEntity(company))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_update_company",
		Arguments: map[string]any{
			"id":          companyID,
			"companyName": "Updated Company Name",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["companyName"].(string)
	if !strings.Contains(nameStr, "Updated Company Name") {
		t.Errorf("expected companyName to contain 'Updated Company Name', got %v", m["companyName"])
	}
}

// TestSearchCompaniesHandler_WithFilters verifies filters execution over wire.
func TestSearchCompaniesHandler_WithFilters(t *testing.T) {
	company := autotasktest.CompanyFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(company))
	ctx := context.Background()

	active := true
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_companies",
		Arguments: map[string]any{
			"searchTerm": "Acme",
			"isActive":   active,
			"maxResults": 10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}
