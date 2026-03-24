package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterCompanyTools_NoPanic verifies that RegisterCompanyTools registers all
// three tools without panicking.
func TestRegisterCompanyTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterCompanyTools(s, client, mapper)
}

// TestSearchCompaniesHandler_ReturnsNoCompaniesFound tests the empty-result case.
func TestSearchCompaniesHandler_ReturnsNoCompaniesFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchCompaniesHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchCompaniesInput{})
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
	if text.Text != "No companies found" {
		t.Errorf("expected 'No companies found', got %q", text.Text)
	}
}

// TestSearchCompaniesHandler_ReturnsCompanies tests that seeded companies are returned.
func TestSearchCompaniesHandler_ReturnsCompanies(t *testing.T) {
	company := autotasktest.CompanyFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(company))
	mapper := newTestMapper(client)

	handler := searchCompaniesHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchCompaniesInput{})
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
		t.Error("expected at least one company in results")
	}
}

// TestCreateCompanyHandler_Success tests that a company can be created.
func TestCreateCompanyHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.CompanyFixture()),
	)

	handler := createCompanyHandler(client)
	ctx := context.Background()

	in := CreateCompanyInput{
		CompanyName: "Test Company",
		CompanyType: 1,
		Phone:       "555-1234",
		City:        "Springfield",
	}

	result, _, err := handler(ctx, nil, in)
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

	var m map[string]any
	if err := json.Unmarshal([]byte(text.Text), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}
}

// TestUpdateCompanyHandler_Success tests that an existing company can be updated.
func TestUpdateCompanyHandler_Success(t *testing.T) {
	company := autotasktest.CompanyFixture()
	companyID, ok := company.ID.Get()
	if !ok {
		t.Fatal("fixture company has no ID")
	}

	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(company))

	handler := updateCompanyHandler(client)
	ctx := context.Background()

	in := UpdateCompanyInput{
		ID:          companyID,
		CompanyName: "Updated Company Name",
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(text.Text), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}
}

// TestSearchCompaniesHandler_WithFilters verifies that filters can be applied without error.
func TestSearchCompaniesHandler_WithFilters(t *testing.T) {
	company := autotasktest.CompanyFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(company))
	mapper := newTestMapper(client)

	handler := searchCompaniesHandler(client, mapper)
	ctx := context.Background()

	active := true
	in := SearchCompaniesInput{
		SearchTerm: "Acme",
		IsActive:   &active,
		Page:       1,
		PageSize:   10,
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
