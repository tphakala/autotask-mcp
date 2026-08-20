package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterExpenseTools_NoPanic verifies registration does not panic.
func TestRegisterExpenseTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterExpenseTools(s, client)
}

// TestGetExpenseReportHandler_NotFound tests that a missing report returns an error result over wire.
func TestGetExpenseReportHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_expense_report",
		Arguments: map[string]any{
			"reportId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing expense report")
	}
}

// TestSearchExpenseReportsHandler_NoResults tests the empty-result case over wire.
func TestSearchExpenseReportsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_expense_reports",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result for empty search, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned reports, got %d", resp.Summary.Returned)
	}
}

// TestCreateExpenseReportHandler_InvalidDate tests that an invalid date returns an error result over wire.
func TestCreateExpenseReportHandler_InvalidDate(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_expense_report",
		Arguments: map[string]any{
			"name":           "Test Report",
			"submitterId":    5001,
			"weekEndingDate": "not-a-date",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateExpenseReportHandler_Success tests creating an expense report over wire.
func TestCreateExpenseReportHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_expense_report",
		Arguments: map[string]any{
			"name":           "Weekly Expense Report",
			"submitterId":    5001,
			"weekEndingDate": "2024-01-19",
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
	if !strings.Contains(nameStr, "Weekly Expense Report") {
		t.Errorf("expected name to contain 'Weekly Expense Report', got %v", m["name"])
	}
}

// TestCreateExpenseItemHandler_InvalidDate tests that an invalid expense date returns an error result over wire.
func TestCreateExpenseItemHandler_InvalidDate(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_expense_item",
		Arguments: map[string]any{
			"expenseReportId": 1,
			"description":     "Lunch",
			"expenseDate":     "bad-date",
			"expenseCategory": 1,
			"amount":          15.50,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestCreateExpenseItemHandler_Success tests creating an expense item over wire.
func TestCreateExpenseItemHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_expense_item",
		Arguments: map[string]any{
			"expenseReportId":     1,
			"description":         "Team lunch",
			"expenseDate":         "2024-01-15",
			"expenseCategory":     1,
			"amount":              45.00,
			"isBillableToCompany": true,
			"isReimbursable":      true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	descStr, _ := m["description"].(string)
	if !strings.Contains(descStr, "Team lunch") {
		t.Errorf("expected description to contain 'Team lunch', got %v", m["description"])
	}
}
