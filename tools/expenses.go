package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetExpenseReportInput defines the input parameters for getting an expense report.
type GetExpenseReportInput struct {
	ReportID int64 `json:"reportId" jsonschema:"Expense report ID to retrieve"`
}

// SearchExpenseReportsInput defines the input parameters for searching expense reports.
type SearchExpenseReportsInput struct {
	SubmitterID int64 `json:"submitterId,omitempty" jsonschema:"Filter by submitter resource ID"`
	Status      int   `json:"status,omitempty" jsonschema:"Filter by report status"`
	PageSize    int   `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// CreateExpenseReportInput defines the input parameters for creating an expense report.
type CreateExpenseReportInput struct {
	Name           string `json:"name" jsonschema:"Expense report name"`
	SubmitterID    int64  `json:"submitterId" jsonschema:"Resource ID of the submitter"`
	WeekEndingDate string `json:"weekEndingDate" jsonschema:"Week-ending date (YYYY-MM-DD or ISO format)"`
}

// CreateExpenseItemInput defines the input parameters for creating an expense item.
type CreateExpenseItemInput struct {
	ExpenseReportID     int64   `json:"expenseReportId" jsonschema:"Expense report ID to add the item to"`
	Description         string  `json:"description" jsonschema:"Expense item description"`
	ExpenseDate         string  `json:"expenseDate" jsonschema:"Date of expense (YYYY-MM-DD or ISO format)"`
	ExpenseCategory     int     `json:"expenseCategory" jsonschema:"Expense category ID"`
	Amount              float64 `json:"amount" jsonschema:"Expense amount"`
	CompanyID           int64   `json:"companyId,omitempty" jsonschema:"Company to bill"`
	HaveReceipt         bool    `json:"haveReceipt,omitempty" jsonschema:"Whether a receipt is attached"`
	IsBillableToCompany bool    `json:"isBillableToCompany,omitempty" jsonschema:"Whether to bill to company"`
	IsReimbursable      bool    `json:"isReimbursable,omitempty" jsonschema:"Whether the expense is reimbursable"`
	PaymentType         int     `json:"paymentType,omitempty" jsonschema:"Payment type ID"`
}

// RegisterExpenseTools registers all expense-related MCP tools with the server.
func RegisterExpenseTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_expense_report",
		Description: "Retrieve one expense report by its numeric reportId, returning its full field set. Use this to fetch a single known report; to locate reports by submitter or status use autotask_search_expense_reports instead. Read-only.",
		Annotations: readOnlyTool("Get expense report"),
	}, getExpenseReportHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_expense_reports",
		Description: "Find expense reports filtered by submitter resource or status, returning up to pageSize records (default 25, max 500). Use this to locate reports, then autotask_get_expense_report for one report by ID. Read-only.",
		Annotations: readOnlyTool("Search expense reports"),
	}, searchExpenseReportsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_expense_report",
		Description: "Create the header of an expense report (its name, submitter, and week-ending date) that acts as the container for expense line items. Requires name, submitterId, and weekEndingDate; returns the created report including its new ID. Add individual expenses to it with autotask_create_expense_item. Writes to Autotask.",
		Annotations: createTool("Create expense report"),
	}, createExpenseReportHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_expense_item",
		Description: "Add one expense line item (amount, category, date, and description) to an existing expense report identified by expenseReportId, with optional billing company, receipt, reimbursable, and payment-type fields. Requires expenseReportId, description, expenseDate, expenseCategory, and amount; the report must already exist, so create it first with autotask_create_expense_report. Writes to Autotask.",
		Annotations: createTool("Create expense item"),
	}, createExpenseItemHandler(client))
}

// getExpenseReportHandler returns a handler that retrieves a single expense report.
func getExpenseReportHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetExpenseReportInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetExpenseReportInput) (*mcp.CallToolResult, any, error) {
		report, err := autotask.Get[entities.ExpenseReport](ctx, client, in.ReportID)
		if err != nil {
			return errorResult("failed to get expense report %d: %v", in.ReportID, err)
		}

		m, err := entityToMap(report)
		if err != nil {
			return errorResult("failed to convert expense report: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		reports, err := autotask.List[entities.ExpenseReport](ctx, client, q)
		if err != nil {
			return errorResult("failed to search expense reports: %v", err)
		}

		if len(reports) == 0 {
			return textResult("No expense reports found")
		}

		maps, err := entitiesToMaps(reports)
		if err != nil {
			return errorResult("failed to convert expense reports: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
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

		entity := &entities.ExpenseReport{
			Name:        autotask.Set(in.Name),
			SubmitterID: autotask.Set(in.SubmitterID),
			WeekEnding:  autotask.Set(weekEnding),
		}
		created, err := autotask.Create[entities.ExpenseReport](ctx, client, entity)
		if err != nil {
			return errorResult("failed to create expense report: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created expense report: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		entity := &entities.ExpenseItem{
			ExpenseReportID:              autotask.Set(in.ExpenseReportID),
			Description:                  autotask.Set(in.Description),
			ExpenseDate:                  autotask.Set(expenseDate),
			ExpenseCategory:              autotask.Set(int64(in.ExpenseCategory)),
			ExpenseCurrencyExpenseAmount: autotask.Set(in.Amount),
		}
		if in.CompanyID != 0 {
			entity.CompanyID = autotask.Set(in.CompanyID)
		}
		if in.HaveReceipt {
			entity.HaveReceipt = autotask.Set(in.HaveReceipt)
		}
		if in.IsBillableToCompany {
			entity.IsBillableToCompany = autotask.Set(in.IsBillableToCompany)
		}
		if in.IsReimbursable {
			entity.IsReimbursable = autotask.Set(in.IsReimbursable)
		}
		if in.PaymentType != 0 {
			entity.PaymentType = autotask.Set(int64(in.PaymentType))
		}

		created, err := autotask.Create[entities.ExpenseItem](ctx, client, entity)
		if err != nil {
			return errorResult("failed to create expense item: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created expense item: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created expense item: %v", err)
		}

		return textResult("%s", string(data))
	}
}
