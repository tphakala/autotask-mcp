package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterProjectTools_NoPanic verifies that RegisterProjectTools registers all
// two tools without panicking.
func TestRegisterProjectTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterProjectTools(s, client, mapper)
}

// TestSearchProjectsHandler_ReturnsNoProjectsFound tests the empty-result case over wire.
func TestSearchProjectsHandler_ReturnsNoProjectsFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_projects",
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
		t.Errorf("expected 0 returned projects, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchProjectsHandler_ReturnsProjects tests that seeded projects are returned over wire.
func TestSearchProjectsHandler_ReturnsProjects(t *testing.T) {
	project := autotasktest.ProjectFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(project))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_projects",
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
		t.Errorf("expected at least 1 returned project, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected project to contain 'id' field")
	}
}

// TestCreateProjectHandler_Success tests creating a project over wire.
func TestCreateProjectHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.ProjectFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_project",
		Arguments: map[string]any{
			"companyID":   1001,
			"projectName": "New Infrastructure Project",
			"status":      1,
			"description": "A test project",
			"startDate":   "2024-01-15",
			"endDate":     "2024-06-30",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["projectName"].(string)
	if !strings.Contains(nameStr, "New Infrastructure Project") {
		t.Errorf("expected projectName to contain 'New Infrastructure Project', got %v", m["projectName"])
	}
}

// TestCreateProjectHandler_InvalidDate tests that an invalid date returns an error result over wire.
func TestCreateProjectHandler_InvalidDate(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.ProjectFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_project",
		Arguments: map[string]any{
			"companyID":   1001,
			"projectName": "Test",
			"status":      1,
			"startDate":   "not-a-date",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestSearchProjectsHandler_WithFilters verifies filters execution over wire.
func TestSearchProjectsHandler_WithFilters(t *testing.T) {
	project := autotasktest.ProjectFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(project))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_projects",
		Arguments: map[string]any{
			"searchTerm": "Infrastructure",
			"companyID":  1001,
			"status":     1,
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
