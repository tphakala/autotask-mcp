package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterProjectTools_NoPanic verifies that RegisterProjectTools registers all
// two tools without panicking.
func TestRegisterProjectTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterProjectTools(s, client, mapper)
}

// TestSearchProjectsHandler_ReturnsNoProjectsFound tests the empty-result case.
func TestSearchProjectsHandler_ReturnsNoProjectsFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchProjectsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchProjectsInput{})
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
	if text.Text != "No projects found" {
		t.Errorf("expected 'No projects found', got %q", text.Text)
	}
}

// TestSearchProjectsHandler_ReturnsProjects tests that seeded projects are returned.
func TestSearchProjectsHandler_ReturnsProjects(t *testing.T) {
	project := autotasktest.ProjectFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(project))
	mapper := newTestMapper(client)

	handler := searchProjectsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchProjectsInput{})
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
		t.Error("expected at least one project in results")
	}
}

// TestCreateProjectHandler_Success tests that a project can be created.
func TestCreateProjectHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.ProjectFixture()),
	)

	handler := createProjectHandler(client)
	ctx := context.Background()

	in := CreateProjectInput{
		CompanyID:   1001,
		ProjectName: "New Infrastructure Project",
		Status:      1,
		Description: "A test project",
		StartDate:   "2024-01-15",
		EndDate:     "2024-06-30",
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

// TestCreateProjectHandler_InvalidDate tests that an invalid date returns an error result.
func TestCreateProjectHandler_InvalidDate(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.ProjectFixture()),
	)

	handler := createProjectHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateProjectInput{
		CompanyID:   1001,
		ProjectName: "Test",
		Status:      1,
		StartDate:   "not-a-date",
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

// TestSearchProjectsHandler_WithFilters verifies that multiple filters can be applied.
func TestSearchProjectsHandler_WithFilters(t *testing.T) {
	project := autotasktest.ProjectFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(project))
	mapper := newTestMapper(client)

	handler := searchProjectsHandler(client, mapper)
	ctx := context.Background()

	in := SearchProjectsInput{
		SearchTerm: "Infrastructure",
		CompanyID:  1001,
		Status:     1,
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
