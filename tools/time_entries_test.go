package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterTimeEntryTools_NoPanic verifies that RegisterTimeEntryTools registers all
// two tools without panicking.
func TestRegisterTimeEntryTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterTimeEntryTools(s, client, mapper)
}

// TestSearchTimeEntriesHandler_ReturnsNoEntriesFound tests the empty-result case over wire.
func TestSearchTimeEntriesHandler_ReturnsNoEntriesFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_time_entries",
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
		t.Errorf("expected 0 returned entries, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchTimeEntriesHandler_ReturnsEntries tests that seeded entries are returned over wire.
func TestSearchTimeEntriesHandler_ReturnsEntries(t *testing.T) {
	entry := autotasktest.TimeEntryFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(entry))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_time_entries",
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
		t.Errorf("expected at least 1 returned time entry, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected time entry to contain 'id' field")
	}
}

// TestCreateTimeEntryHandler_Success tests creating a time entry over wire.
func TestCreateTimeEntryHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.TimeEntryFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_time_entry",
		Arguments: map[string]any{
			"resourceID":   5001,
			"dateWorked":   "2024-01-15",
			"hoursWorked":  2.5,
			"summaryNotes": "Fixed server issue",
			"ticketID":     3001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	notesStr, _ := m["summaryNotes"].(string)
	if !strings.Contains(notesStr, "Fixed server issue") {
		t.Errorf("expected summaryNotes to contain 'Fixed server issue', got %v", m["summaryNotes"])
	}
}

// TestCreateTimeEntryHandler_InvalidDate tests that an invalid date returns an error result over wire.
func TestCreateTimeEntryHandler_InvalidDate(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.TimeEntryFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_time_entry",
		Arguments: map[string]any{
			"resourceID":   5001,
			"dateWorked":   "not-a-date",
			"hoursWorked":  1.0,
			"summaryNotes": "test",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestSearchTimeEntriesHandler_WithFilters verifies filters execution over wire.
func TestSearchTimeEntriesHandler_WithFilters(t *testing.T) {
	entry := autotasktest.TimeEntryFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(entry))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_time_entries",
		Arguments: map[string]any{
			"resourceID":      5001,
			"ticketID":        3001,
			"dateWorkedAfter": "2024-01-01T00:00:00Z",
			"maxResults":      10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}
