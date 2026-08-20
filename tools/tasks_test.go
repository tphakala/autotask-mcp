package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterTaskTools_NoPanic verifies that RegisterTaskTools registers all
// two tools without panicking.
func TestRegisterTaskTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterTaskTools(s, client, mapper)
}

// TestSearchTasksHandler_ReturnsNoTasksFound tests the empty-result case over wire.
func TestSearchTasksHandler_ReturnsNoTasksFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_tasks",
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
		t.Errorf("expected 0 returned tasks, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchTasksHandler_ReturnsTasks tests that seeded tasks are returned over wire.
func TestSearchTasksHandler_ReturnsTasks(t *testing.T) {
	task := autotasktest.TaskFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(task))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_tasks",
		Arguments: map[string]any{
			"status": 1,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 returned task, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected task to contain 'id' field")
	}
}

// TestCreateTaskHandler_Success tests creating a task over wire.
func TestCreateTaskHandler_Success(t *testing.T) {
	proj := autotasktest.ProjectFixture()
	projID, ok := proj.ID.Get()
	if !ok {
		t.Fatal("fixture project has no ID")
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(proj))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_task",
		Arguments: map[string]any{
			"projectID":   projID,
			"title":       "New task",
			"status":      1,
			"description": "A test task",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	titleStr, _ := m["title"].(string)
	if !strings.Contains(titleStr, "New task") {
		t.Errorf("expected title to contain 'New task', got %v", m["title"])
	}
}

// TestSearchTasksHandler_WithFilters verifies filters execution over wire.
func TestSearchTasksHandler_WithFilters(t *testing.T) {
	task := autotasktest.TaskFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(task))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_tasks",
		Arguments: map[string]any{
			"searchTerm":         "firmware",
			"projectID":          4001,
			"status":             1,
			"assignedResourceID": 5001,
			"maxResults":         10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}


