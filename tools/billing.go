package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/autotask-mcp/services"
)

// GetBillingItemInput defines the input parameters for getting a billing item.
type GetBillingItemInput struct {
	BillingItemID int64 `json:"billingItemId" jsonschema:"Billing item ID to retrieve"`
}

// SearchBillingItemsInput defines the input parameters for searching billing items.
type SearchBillingItemsInput struct {
	CompanyID    int64  `json:"companyId,omitempty" jsonschema:"Filter by company ID"`
	TicketID     int64  `json:"ticketId,omitempty" jsonschema:"Filter by ticket ID"`
	ProjectID    int64  `json:"projectId,omitempty" jsonschema:"Filter by project ID"`
	ContractID   int64  `json:"contractId,omitempty" jsonschema:"Filter by contract ID"`
	InvoiceID    int64  `json:"invoiceId,omitempty" jsonschema:"Filter by invoice ID"`
	PostedAfter  string `json:"postedAfter,omitempty" jsonschema:"Filter items posted on or after this date (ISO format)"`
	PostedBefore string `json:"postedBefore,omitempty" jsonschema:"Filter items posted on or before this date (ISO format)"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize     int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// SearchBillingItemApprovalLevelsInput defines the input for searching billing item approval levels.
type SearchBillingItemApprovalLevelsInput struct {
	TimeEntryID        int64  `json:"timeEntryId,omitempty" jsonschema:"Filter by time entry ID"`
	ApprovalResourceID int64  `json:"approvalResourceId,omitempty" jsonschema:"Filter by approving resource ID"`
	ApprovalLevel      int    `json:"approvalLevel,omitempty" jsonschema:"Filter by approval level"`
	ApprovedAfter      string `json:"approvedAfter,omitempty" jsonschema:"Filter items approved on or after this date (ISO format)"`
	ApprovedBefore     string `json:"approvedBefore,omitempty" jsonschema:"Filter items approved on or before this date (ISO format)"`
	Page               int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize           int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterBillingTools registers all billing-related MCP tools with the server.
func RegisterBillingTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_billing_item",
		Description: "Get a specific billing item by ID.",
	}, getBillingItemHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_billing_items",
		Description: "Search for billing items in Autotask.",
	}, searchBillingItemsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_billing_item_approval_levels",
		Description: "Search for billing item approval levels in Autotask.",
	}, searchBillingItemApprovalLevelsHandler(client))
}

// getBillingItemHandler returns a handler that retrieves a single billing item.
func getBillingItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetBillingItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetBillingItemInput) (*mcp.CallToolResult, any, error) {
		if in.BillingItemID == 0 {
			return errorResult("billingItemId is required")
		}
		item, err := autotask.GetRaw(ctx, client, "BillingItems", in.BillingItemID)
		if err != nil {
			return errorResult("failed to get billing item %d: %v", in.BillingItemID, err)
		}

		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return errorResult("failed to marshal billing item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchBillingItemsHandler returns a handler that searches billing items.
func searchBillingItemsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemsInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.TicketID != 0 {
			q.Where("ticketID", autotask.OpEq, in.TicketID)
		}
		if in.ProjectID != 0 {
			q.Where("projectID", autotask.OpEq, in.ProjectID)
		}
		if in.ContractID != 0 {
			q.Where("contractID", autotask.OpEq, in.ContractID)
		}
		if in.InvoiceID != 0 {
			q.Where("invoiceID", autotask.OpEq, in.InvoiceID)
		}
		if in.PostedAfter != "" {
			q.Where("postedDate", autotask.OpGte, in.PostedAfter)
		}
		if in.PostedBefore != "" {
			q.Where("postedDate", autotask.OpLte, in.PostedBefore)
		}

		items, err := autotask.ListRaw(ctx, client, "BillingItems", q)
		if err != nil {
			return errorResult("failed to search billing items: %v", err)
		}

		if len(items) == 0 {
			return textResult("No billing items found")
		}

		return searchResult(ctx, mapper, items, "autotask_search_billing_items", page, pageSize)
	}
}

// searchBillingItemApprovalLevelsHandler returns a handler that searches billing item approval levels.
func searchBillingItemApprovalLevelsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemApprovalLevelsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemApprovalLevelsInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.TimeEntryID != 0 {
			q.Where("timeEntryID", autotask.OpEq, in.TimeEntryID)
		}
		if in.ApprovalResourceID != 0 {
			q.Where("approvalResourceID", autotask.OpEq, in.ApprovalResourceID)
		}
		if in.ApprovalLevel != 0 {
			q.Where("approvalLevel", autotask.OpEq, in.ApprovalLevel)
		}
		if in.ApprovedAfter != "" {
			q.Where("approvedDate", autotask.OpGte, in.ApprovedAfter)
		}
		if in.ApprovedBefore != "" {
			q.Where("approvedDate", autotask.OpLte, in.ApprovedBefore)
		}

		levels, err := autotask.ListRaw(ctx, client, "BillingItemApprovalLevels", q)
		if err != nil {
			return errorResult("failed to search billing item approval levels: %v", err)
		}

		if len(levels) == 0 {
			return textResult("No billing item approval levels found")
		}

		return searchResult(ctx, nil, levels, "autotask_search_billing_item_approval_levels", page, pageSize)
	}
}
