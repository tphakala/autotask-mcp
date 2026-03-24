package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterExpenseTools_NoPanic verifies registration does not panic.
func TestRegisterExpenseTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterExpenseTools(s, client)
}

// TestGetExpenseReportHandler_NotFound tests that a missing report returns an error result.
func TestGetExpenseReportHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getExpenseReportHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetExpenseReportInput{ReportID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing expense report")
	}
}

// TestSearchExpenseReportsHandler_NoResults tests the empty-result case.
func TestSearchExpenseReportsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchExpenseReportsHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchExpenseReportsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result for empty search, got IsError=true")
	}
}

// TestCreateExpenseReportHandler_InvalidDate tests that an invalid date returns an error result.
func TestCreateExpenseReportHandler_InvalidDate(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createExpenseReportHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateExpenseReportInput{
		Name:           "Test Report",
		SubmitterID:    5001,
		WeekEndingDate: "not-a-date",
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

// TestCreateExpenseReportHandler_Success tests that an expense report can be created.
func TestCreateExpenseReportHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createExpenseReportHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateExpenseReportInput{
		Name:           "Weekly Expense Report",
		SubmitterID:    5001,
		WeekEndingDate: "2024-01-19",
		Description:    "Test report",
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

// TestCreateExpenseItemHandler_InvalidDate tests that an invalid expense date returns an error result.
func TestCreateExpenseItemHandler_InvalidDate(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createExpenseItemHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateExpenseItemInput{
		ExpenseReportID: 1,
		Description:     "Lunch",
		ExpenseDate:     "bad-date",
		ExpenseCategory: 1,
		Amount:          15.50,
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

// TestCreateExpenseItemHandler_Success tests that an expense item can be created.
func TestCreateExpenseItemHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := createExpenseItemHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateExpenseItemInput{
		ExpenseReportID:     1,
		Description:         "Team lunch",
		ExpenseDate:         "2024-01-15",
		ExpenseCategory:     1,
		Amount:              45.00,
		IsBillableToCompany: true,
		IsReimbursable:      true,
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
