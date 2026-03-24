package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
)

// GetExpenseReportInput defines the input parameters for getting an expense report.
type GetExpenseReportInput struct {
	ReportID int64 `json:"reportId" jsonschema:"Expense report ID to retrieve"`
}

// SearchExpenseReportsInput defines the input parameters for searching expense reports.
type SearchExpenseReportsInput struct {
	SubmitterID int64  `json:"submitterId,omitempty" jsonschema:"Filter by submitter resource ID"`
	Status      int    `json:"status,omitempty" jsonschema:"Filter by report status"`
	PageSize    int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// CreateExpenseReportInput defines the input parameters for creating an expense report.
type CreateExpenseReportInput struct {
	Name          string `json:"name" jsonschema:"Expense report name"`
	SubmitterID   int64  `json:"submitterId" jsonschema:"Resource ID of the submitter"`
	WeekEndingDate string `json:"weekEndingDate" jsonschema:"Week-ending date (YYYY-MM-DD or ISO format)"`
	Description   string `json:"description,omitempty" jsonschema:"Report description"`
}

// CreateExpenseItemInput defines the input parameters for creating an expense item.
type CreateExpenseItemInput struct {
	ExpenseReportID    int64   `json:"expenseReportId" jsonschema:"Expense report ID to add the item to"`
	Description        string  `json:"description" jsonschema:"Expense item description"`
	ExpenseDate        string  `json:"expenseDate" jsonschema:"Date of expense (YYYY-MM-DD or ISO format)"`
	ExpenseCategory    int     `json:"expenseCategory" jsonschema:"Expense category ID"`
	Amount             float64 `json:"amount" jsonschema:"Expense amount"`
	CompanyID          int64   `json:"companyId,omitempty" jsonschema:"Company to bill"`
	HaveReceipt        bool    `json:"haveReceipt,omitempty" jsonschema:"Whether a receipt is attached"`
	IsBillableToCompany bool   `json:"isBillableToCompany,omitempty" jsonschema:"Whether to bill to company"`
	IsReimbursable     bool    `json:"isReimbursable,omitempty" jsonschema:"Whether the expense is reimbursable"`
	PaymentType        int     `json:"paymentType,omitempty" jsonschema:"Payment type ID"`
}

// RegisterExpenseTools registers all expense-related MCP tools with the server.
func RegisterExpenseTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_expense_report",
		Description: "Get a specific expense report by ID.",
	}, getExpenseReportHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_expense_reports",
		Description: "Search for expense reports in Autotask.",
	}, searchExpenseReportsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_expense_report",
		Description: "Create a new expense report.",
	}, createExpenseReportHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_expense_item",
		Description: "Create a new expense item within an expense report.",
	}, createExpenseItemHandler(client))
}

// getExpenseReportHandler returns a handler that retrieves a single expense report.
func getExpenseReportHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetExpenseReportInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetExpenseReportInput) (*mcp.CallToolResult, any, error) {
		report, err := autotask.GetRaw(ctx, client, "ExpenseReports", in.ReportID)
		if err != nil {
			return errorResult("failed to get expense report %d: %v", in.ReportID, err)
		}

		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return errorResult("failed to marshal expense report: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchExpenseReportsHandler returns a handler that searches expense reports.
func searchExpenseReportsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchExpenseReportsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchExpenseReportsInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SubmitterID != 0 {
			q.Where("submitterID", autotask.OpEq, in.SubmitterID)
		}
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		}

		reports, err := autotask.ListRaw(ctx, client, "ExpenseReports", q)
		if err != nil {
			return errorResult("failed to search expense reports: %v", err)
		}

		if len(reports) == 0 {
			return textResult("No expense reports found")
		}

		data, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			return errorResult("failed to marshal expense reports: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createExpenseReportHandler returns a handler that creates a new expense report.
func createExpenseReportHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateExpenseReportInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateExpenseReportInput) (*mcp.CallToolResult, any, error) {
		weekEnding, err := parseDate(in.WeekEndingDate)
		if err != nil {
			return errorResult("invalid weekEndingDate format (expected YYYY-MM-DD or ISO format): %v", err)
		}

		payload := map[string]any{
			"name":           in.Name,
			"submitterID":    in.SubmitterID,
			"weekEndingDate": weekEnding.Format("2006-01-02"),
		}
		if in.Description != "" {
			payload["description"] = in.Description
		}

		created, err := autotask.CreateRaw(ctx, client, "ExpenseReports", payload)
		if err != nil {
			return errorResult("failed to create expense report: %v", err)
		}

		data, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created expense report: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createExpenseItemHandler returns a handler that creates a new expense item.
func createExpenseItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateExpenseItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateExpenseItemInput) (*mcp.CallToolResult, any, error) {
		expenseDate, err := parseDate(in.ExpenseDate)
		if err != nil {
			return errorResult("invalid expenseDate format (expected YYYY-MM-DD or ISO format): %v", err)
		}

		payload := map[string]any{
			"expenseReportID": in.ExpenseReportID,
			"description":     in.Description,
			"expenseDate":     expenseDate.Format("2006-01-02"),
			"expenseCategory": in.ExpenseCategory,
			"amount":          in.Amount,
		}
		if in.CompanyID != 0 {
			payload["companyID"] = in.CompanyID
		}
		if in.HaveReceipt {
			payload["haveReceipt"] = in.HaveReceipt
		}
		if in.IsBillableToCompany {
			payload["isBillableToCompany"] = in.IsBillableToCompany
		}
		if in.IsReimbursable {
			payload["isReimbursable"] = in.IsReimbursable
		}
		if in.PaymentType != 0 {
			payload["paymentType"] = in.PaymentType
		}

		created, err := autotask.CreateRaw(ctx, client, "ExpenseItems", payload)
		if err != nil {
			return errorResult("failed to create expense item: %v", err)
		}

		data, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created expense item: %v", err)
		}

		return textResult("%s", string(data))
	}
}
