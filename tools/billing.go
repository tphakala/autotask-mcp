package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
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
	MaxResults   int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// SearchBillingItemApprovalLevelsInput defines the input for searching billing item approval levels.
type SearchBillingItemApprovalLevelsInput struct {
	TimeEntryID        int64  `json:"timeEntryId,omitempty" jsonschema:"Filter by time entry ID"`
	ApprovalResourceID int64  `json:"approvalResourceId,omitempty" jsonschema:"Filter by approving resource ID"`
	ApprovalLevel      int    `json:"approvalLevel,omitempty" jsonschema:"Filter by approval level"`
	ApprovedAfter      string `json:"approvedAfter,omitempty" jsonschema:"Filter items approved on or after this date (ISO format)"`
	ApprovedBefore     string `json:"approvedBefore,omitempty" jsonschema:"Filter items approved on or before this date (ISO format)"`
	MaxResults         int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// RegisterBillingTools registers all billing-related MCP tools with the server.
func RegisterBillingTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_billing_item",
		Description: "Retrieve one billing item by its numeric billingItemId, returning its full field set. Use this to fetch a single known item; to find items by company, ticket, project, contract, invoice, or posted-date range use autotask_search_billing_items instead. Read-only.",
		Annotations: readOnlyTool("Get billing item"),
	}, getBillingItemHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_billing_items",
		Description: "Find billing items by company, ticket, project, contract, invoice, or posted-date range, returning a compact summary of matching records (up to maxResults, default 25, max 500). Use this to locate billing items, then autotask_get_billing_item for the full field set of one item. Read-only.",
		Annotations: readOnlyTool("Search billing items"),
	}, searchBillingItemsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_billing_item_approval_levels",
		Description: "Find billing-item approval-level records by time entry, approving resource, approval level, or approved-date range, returning a compact summary of matching records (up to maxResults, default 25, max 500). Read-only.",
		Annotations: readOnlyTool("Search billing item approval levels"),
	}, searchBillingItemApprovalLevelsHandler(client))
}

// getBillingItemHandler returns a handler that retrieves a single billing item.
func getBillingItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetBillingItemInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetBillingItemInput) (*mcp.CallToolResult, map[string]any, error) {
		if in.BillingItemID == 0 {
			return nil, nil, fmt.Errorf("billingItemId is required")
		}
		item, err := autotask.Get[entities.BillingItem](ctx, client, in.BillingItemID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(item)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchBillingItemsHandler returns a handler that searches billing items.
func searchBillingItemsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)
		q := autotask.NewQuery().Limit(maxResults + 1)

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

		items, err := autotask.List[entities.BillingItem](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(items) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(items)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, mapper, maps, "autotask_search_billing_items", maxResults)
	}
}

// searchBillingItemApprovalLevelsHandler returns a handler that searches billing item approval levels.
func searchBillingItemApprovalLevelsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemApprovalLevelsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchBillingItemApprovalLevelsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)
		q := autotask.NewQuery().Limit(maxResults + 1)

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

		levels, err := autotask.List[entities.BillingItemApprovalLevel](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(levels) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(levels)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, nil, maps, "autotask_search_billing_item_approval_levels", maxResults)
	}
}
