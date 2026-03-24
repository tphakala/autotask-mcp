package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterTimeEntryTools_NoPanic verifies that RegisterTimeEntryTools registers all
// two tools without panicking.
func TestRegisterTimeEntryTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterTimeEntryTools(s, client, mapper)
}

// TestSearchTimeEntriesHandler_ReturnsNoEntriesFound tests the empty-result case.
func TestSearchTimeEntriesHandler_ReturnsNoEntriesFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchTimeEntriesHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTimeEntriesInput{})
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
	if text.Text != "No time entries found" {
		t.Errorf("expected 'No time entries found', got %q", text.Text)
	}
}

// TestSearchTimeEntriesHandler_ReturnsEntries tests that seeded entries are returned.
func TestSearchTimeEntriesHandler_ReturnsEntries(t *testing.T) {
	entry := autotasktest.TimeEntryFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(entry))
	mapper := newTestMapper(client)

	handler := searchTimeEntriesHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTimeEntriesInput{})
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
		t.Error("expected at least one time entry in results")
	}
}

// TestCreateTimeEntryHandler_Success tests that a time entry can be created.
func TestCreateTimeEntryHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.TimeEntryFixture()),
	)

	handler := createTimeEntryHandler(client)
	ctx := context.Background()

	in := CreateTimeEntryInput{
		ResourceID:   5001,
		DateWorked:   "2024-01-15",
		HoursWorked:  2.5,
		SummaryNotes: "Fixed server issue",
		TicketID:     3001,
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

// TestCreateTimeEntryHandler_InvalidDate tests that an invalid date returns an error result.
func TestCreateTimeEntryHandler_InvalidDate(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.TimeEntryFixture()),
	)

	handler := createTimeEntryHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateTimeEntryInput{
		ResourceID:   5001,
		DateWorked:   "not-a-date",
		HoursWorked:  1.0,
		SummaryNotes: "test",
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

// TestSearchTimeEntriesHandler_WithFilters verifies that filters can be applied.
func TestSearchTimeEntriesHandler_WithFilters(t *testing.T) {
	entry := autotasktest.TimeEntryFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(entry))
	mapper := newTestMapper(client)

	handler := searchTimeEntriesHandler(client, mapper)
	ctx := context.Background()

	in := SearchTimeEntriesInput{
		ResourceID:      5001,
		TicketID:        3001,
		DateWorkedAfter: "2024-01-01T00:00:00Z",
		Page:            1,
		PageSize:        10,
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
