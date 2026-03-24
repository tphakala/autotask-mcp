package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterTaskTools_NoPanic verifies that RegisterTaskTools registers all
// two tools without panicking.
func TestRegisterTaskTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterTaskTools(s, client, mapper)
}

// TestSearchTasksHandler_ReturnsNoTasksFound tests the empty-result case.
func TestSearchTasksHandler_ReturnsNoTasksFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchTasksHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTasksInput{})
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
	if text.Text != "No tasks found" {
		t.Errorf("expected 'No tasks found', got %q", text.Text)
	}
}

// TestSearchTasksHandler_ReturnsTasks tests that seeded tasks are returned.
func TestSearchTasksHandler_ReturnsTasks(t *testing.T) {
	task := autotasktest.TaskFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(task))
	mapper := newTestMapper(client)

	handler := searchTasksHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTasksInput{Status: 1})
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
		t.Error("expected at least one task in results")
	}
}

// TestCreateTaskHandler_Success tests that a task can be created.
func TestCreateTaskHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.TaskFixture()),
	)

	handler := createTaskHandler(client)
	ctx := context.Background()

	in := CreateTaskInput{
		ProjectID:   4001,
		Title:       "New task",
		Status:      1,
		Description: "A test task",
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

// TestSearchTasksHandler_WithFilters verifies that multiple filters can be applied.
func TestSearchTasksHandler_WithFilters(t *testing.T) {
	task := autotasktest.TaskFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(task))
	mapper := newTestMapper(client)

	handler := searchTasksHandler(client, mapper)
	ctx := context.Background()

	in := SearchTasksInput{
		SearchTerm:         "firmware",
		ProjectID:          4001,
		Status:             1,
		AssignedResourceID: 5001,
		Page:               1,
		PageSize:           10,
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
