package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// --- Quote inputs ---

// GetQuoteInput defines the input parameters for getting a quote.
type GetQuoteInput struct {
	QuoteID int64 `json:"quoteId" jsonschema:"Quote ID to retrieve"`
}

// SearchQuotesInput defines the input parameters for searching quotes.
type SearchQuotesInput struct {
	CompanyID     int64  `json:"companyId,omitempty" jsonschema:"Filter by company ID"`
	ContactID     int64  `json:"contactId,omitempty" jsonschema:"Filter by contact ID"`
	OpportunityID int64  `json:"opportunityId,omitempty" jsonschema:"Filter by opportunity ID"`
	SearchTerm    string `json:"searchTerm,omitempty" jsonschema:"Search by quote name (partial match)"`
	PageSize      int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// CreateQuoteInput defines the input parameters for creating a quote.
type CreateQuoteInput struct {
	CompanyID      int64  `json:"companyId" jsonschema:"Company ID for the quote"`
	Name           string `json:"name,omitempty" jsonschema:"Quote name"`
	Description    string `json:"description,omitempty" jsonschema:"Quote description"`
	ContactID      int64  `json:"contactId,omitempty" jsonschema:"Contact ID"`
	OpportunityID  int64  `json:"opportunityId,omitempty" jsonschema:"Opportunity ID"`
	EffectiveDate  string `json:"effectiveDate,omitempty" jsonschema:"Effective date (YYYY-MM-DD or ISO format)"`
	ExpirationDate string `json:"expirationDate,omitempty" jsonschema:"Expiration date (YYYY-MM-DD or ISO format)"`
}

// --- Quote Item inputs ---

// GetQuoteItemInput defines the input parameters for getting a quote item.
type GetQuoteItemInput struct {
	QuoteItemID int64 `json:"quoteItemId" jsonschema:"Quote item ID to retrieve"`
}

// SearchQuoteItemsInput defines the input parameters for searching quote items.
type SearchQuoteItemsInput struct {
	QuoteID    int64  `json:"quoteId,omitempty" jsonschema:"Filter by quote ID"`
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by item name (partial match)"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// CreateQuoteItemInput defines the input parameters for creating a quote item.
type CreateQuoteItemInput struct {
	QuoteID            int64   `json:"quoteId" jsonschema:"Quote ID to add item to"`
	Quantity           float64 `json:"quantity" jsonschema:"Item quantity"`
	Name               string  `json:"name,omitempty" jsonschema:"Item name"`
	Description        string  `json:"description,omitempty" jsonschema:"Item description"`
	UnitPrice          float64 `json:"unitPrice,omitempty" jsonschema:"Unit price"`
	UnitCost           float64 `json:"unitCost,omitempty" jsonschema:"Unit cost"`
	UnitDiscount       float64 `json:"unitDiscount,omitempty" jsonschema:"Unit discount amount"`
	LineDiscount       float64 `json:"lineDiscount,omitempty" jsonschema:"Line discount amount"`
	PercentageDiscount float64 `json:"percentageDiscount,omitempty" jsonschema:"Percentage discount"`
	IsOptional         *bool   `json:"isOptional,omitempty" jsonschema:"Whether the item is optional"`
	ServiceID          int64   `json:"serviceID,omitempty" jsonschema:"Service ID (sets quoteItemType=11)"`
	ProductID          int64   `json:"productID,omitempty" jsonschema:"Product ID (sets quoteItemType=1)"`
	ServiceBundleID    int64   `json:"serviceBundleID,omitempty" jsonschema:"Service bundle ID (sets quoteItemType=12)"`
	SortOrderID        int     `json:"sortOrderID,omitempty" jsonschema:"Sort order"`
	QuoteItemType      int     `json:"quoteItemType,omitempty" jsonschema:"Quote item type ID (auto-determined from IDs if not set)"`
}

// UpdateQuoteItemInput defines the input parameters for updating a quote item.
type UpdateQuoteItemInput struct {
	QuoteItemID        int64    `json:"quoteItemId" jsonschema:"Quote item ID to update"`
	Quantity           *float64 `json:"quantity,omitempty" jsonschema:"Updated quantity"`
	UnitPrice          *float64 `json:"unitPrice,omitempty" jsonschema:"Updated unit price"`
	UnitDiscount       *float64 `json:"unitDiscount,omitempty" jsonschema:"Updated unit discount"`
	LineDiscount       *float64 `json:"lineDiscount,omitempty" jsonschema:"Updated line discount"`
	PercentageDiscount *float64 `json:"percentageDiscount,omitempty" jsonschema:"Updated percentage discount"`
	IsOptional         *bool    `json:"isOptional,omitempty" jsonschema:"Updated optional flag"`
	SortOrderID        *int     `json:"sortOrderID,omitempty" jsonschema:"Updated sort order"`
}

// DeleteQuoteItemInput defines the input parameters for deleting a quote item.
type DeleteQuoteItemInput struct {
	QuoteID     int64 `json:"quoteId" jsonschema:"Quote ID that owns the item"`
	QuoteItemID int64 `json:"quoteItemId" jsonschema:"Quote item ID to delete"`
}

// --- Opportunity inputs ---

// GetOpportunityInput defines the input parameters for getting an opportunity.
type GetOpportunityInput struct {
	OpportunityID int64 `json:"opportunityId" jsonschema:"Opportunity ID to retrieve"`
}

// SearchOpportunitiesInput defines the input parameters for searching opportunities.
type SearchOpportunitiesInput struct {
	CompanyID  int64  `json:"companyId,omitempty" jsonschema:"Filter by company ID"`
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by opportunity title (partial match)"`
	Status     int    `json:"status,omitempty" jsonschema:"Filter by status ID"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// CreateOpportunityInput defines the input parameters for creating an opportunity.
type CreateOpportunityInput struct {
	Title                  string  `json:"title" jsonschema:"Opportunity title"`
	CompanyID              int64   `json:"companyId" jsonschema:"Company ID"`
	OwnerResourceID        int64   `json:"ownerResourceId" jsonschema:"Owner resource ID"`
	Status                 int     `json:"status" jsonschema:"Status ID"`
	Stage                  int     `json:"stage" jsonschema:"Stage ID"`
	ProjectedCloseDate     string  `json:"projectedCloseDate" jsonschema:"Projected close date (YYYY-MM-DD or ISO format)"`
	StartDate              string  `json:"startDate" jsonschema:"Start date (YYYY-MM-DD or ISO format)"`
	Probability            int     `json:"probability,omitempty" jsonschema:"Win probability percentage"`
	Amount                 float64 `json:"amount,omitempty" jsonschema:"Opportunity amount"`
	Cost                   float64 `json:"cost,omitempty" jsonschema:"Opportunity cost"`
	UseQuoteTotals         bool    `json:"useQuoteTotals,omitempty" jsonschema:"Whether to use quote totals"`
	TotalAmountMonths      int     `json:"totalAmountMonths,omitempty" jsonschema:"Total amount months"`
	ContactID              int64   `json:"contactId,omitempty" jsonschema:"Contact ID"`
	Description            string  `json:"description,omitempty" jsonschema:"Description"`
	OpportunityCategoryID  int     `json:"opportunityCategoryID,omitempty" jsonschema:"Opportunity category ID"`
}

// --- Invoice inputs ---

// SearchInvoicesInput defines the input parameters for searching invoices.
type SearchInvoicesInput struct {
	CompanyID     int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	InvoiceNumber string `json:"invoiceNumber,omitempty" jsonschema:"Filter by invoice number"`
	IsVoided      *bool  `json:"isVoided,omitempty" jsonschema:"Filter by voided status"`
	PageSize      int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// --- Contract inputs ---

// SearchContractsInput defines the input parameters for searching contracts.
type SearchContractsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by contract name (partial match)"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	Status     int    `json:"status,omitempty" jsonschema:"Filter by status ID"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterFinancialTools registers all financial-related MCP tools with the server.
func RegisterFinancialTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	// Quotes
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_quote",
		Description: "Get a specific quote by ID.",
	}, getQuoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_quotes",
		Description: "Search for quotes in Autotask.",
	}, searchQuotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_quote",
		Description: "Create a new quote in Autotask.",
	}, createQuoteHandler(client))

	// Quote Items
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_quote_item",
		Description: "Get a specific quote item by ID.",
	}, getQuoteItemHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_quote_items",
		Description: "Search for quote items in Autotask.",
	}, searchQuoteItemsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_quote_item",
		Description: "Create a new item on a quote. Quote item type is auto-determined from productID, serviceID, or serviceBundleID if not specified.",
	}, createQuoteItemHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_update_quote_item",
		Description: "Update an existing quote item. Only provided fields are changed.",
	}, updateQuoteItemHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_delete_quote_item",
		Description: "Delete a quote item.",
	}, deleteQuoteItemHandler(client))

	// Opportunities
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_opportunity",
		Description: "Get a specific opportunity by ID.",
	}, getOpportunityHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_opportunities",
		Description: "Search for opportunities in Autotask.",
	}, searchOpportunitiesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_opportunity",
		Description: "Create a new opportunity in Autotask.",
	}, createOpportunityHandler(client))

	// Invoices
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_invoices",
		Description: "Search for invoices in Autotask.",
	}, searchInvoicesHandler(client))

	// Contracts
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_contracts",
		Description: "Search for contracts in Autotask.",
	}, searchContractsHandler(client, mapper))
}

// getQuoteHandler returns a handler that retrieves a single quote.
func getQuoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteInput) (*mcp.CallToolResult, any, error) {
		quote, err := autotask.GetRaw(ctx, client, "Quotes", in.QuoteID)
		if err != nil {
			return errorResult("failed to get quote %d: %v", in.QuoteID, err)
		}

		data, err := json.MarshalIndent(quote, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quote: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchQuotesHandler returns a handler that searches quotes.
func searchQuotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchQuotesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchQuotesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.ContactID != 0 {
			q.Where("contactID", autotask.OpEq, in.ContactID)
		}
		if in.OpportunityID != 0 {
			q.Where("opportunityID", autotask.OpEq, in.OpportunityID)
		}
		if in.SearchTerm != "" {
			q.Where("name", autotask.OpContains, in.SearchTerm)
		}

		quotes, err := autotask.ListRaw(ctx, client, "Quotes", q)
		if err != nil {
			return errorResult("failed to search quotes: %v", err)
		}

		if len(quotes) == 0 {
			return textResult("No quotes found")
		}

		data, err := json.MarshalIndent(quotes, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quotes: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createQuoteHandler returns a handler that creates a new quote.
func createQuoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteInput) (*mcp.CallToolResult, any, error) {
		payload := map[string]any{
			"companyID": in.CompanyID,
		}
		if in.Name != "" {
			payload["name"] = in.Name
		}
		if in.Description != "" {
			payload["description"] = in.Description
		}
		if in.ContactID != 0 {
			payload["contactID"] = in.ContactID
		}
		if in.OpportunityID != 0 {
			payload["opportunityID"] = in.OpportunityID
		}
		if in.EffectiveDate != "" {
			t, err := parseDate(in.EffectiveDate)
			if err != nil {
				return errorResult("invalid effectiveDate format: %v", err)
			}
			payload["effectiveDate"] = t.Format("2006-01-02")
		}
		if in.ExpirationDate != "" {
			t, err := parseDate(in.ExpirationDate)
			if err != nil {
				return errorResult("invalid expirationDate format: %v", err)
			}
			payload["expirationDate"] = t.Format("2006-01-02")
		}

		created, err := autotask.CreateRaw(ctx, client, "Quotes", payload)
		if err != nil {
			return errorResult("failed to create quote: %v", err)
		}

		data, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created quote: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getQuoteItemHandler returns a handler that retrieves a single quote item.
func getQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteItemInput) (*mcp.CallToolResult, any, error) {
		item, err := autotask.GetRaw(ctx, client, "QuoteItems", in.QuoteItemID)
		if err != nil {
			return errorResult("failed to get quote item %d: %v", in.QuoteItemID, err)
		}

		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quote item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchQuoteItemsHandler returns a handler that searches quote items.
func searchQuoteItemsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchQuoteItemsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchQuoteItemsInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.QuoteID != 0 {
			q.Where("quoteID", autotask.OpEq, in.QuoteID)
		}
		if in.SearchTerm != "" {
			q.Where("name", autotask.OpContains, in.SearchTerm)
		}

		items, err := autotask.ListRaw(ctx, client, "QuoteItems", q)
		if err != nil {
			return errorResult("failed to search quote items: %v", err)
		}

		if len(items) == 0 {
			return textResult("No quote items found")
		}

		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quote items: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createQuoteItemHandler returns a handler that creates a new quote item.
func createQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteItemInput) (*mcp.CallToolResult, any, error) {
		payload := map[string]any{
			"quoteID":  in.QuoteID,
			"quantity": in.Quantity,
		}

		if in.Name != "" {
			payload["name"] = in.Name
		}
		if in.Description != "" {
			payload["description"] = in.Description
		}
		if in.UnitPrice != 0 {
			payload["unitPrice"] = in.UnitPrice
		}
		if in.UnitCost != 0 {
			payload["unitCost"] = in.UnitCost
		}
		if in.UnitDiscount != 0 {
			payload["unitDiscount"] = in.UnitDiscount
		}
		if in.LineDiscount != 0 {
			payload["lineDiscount"] = in.LineDiscount
		}
		if in.PercentageDiscount != 0 {
			payload["percentageDiscount"] = in.PercentageDiscount
		}
		if in.IsOptional != nil {
			payload["isOptional"] = *in.IsOptional
		}
		if in.ProductID != 0 {
			payload["productID"] = in.ProductID
		}
		if in.ServiceID != 0 {
			payload["serviceID"] = in.ServiceID
		}
		if in.ServiceBundleID != 0 {
			payload["serviceBundleID"] = in.ServiceBundleID
		}
		if in.SortOrderID != 0 {
			payload["sortOrderID"] = in.SortOrderID
		}

		// Auto-determine quoteItemType if not provided.
		// Autotask quote item type constants.
		const (
			quoteItemProduct       = 1
			quoteItemService       = 11
			quoteItemServiceBundle = 12
		)
		itemType := in.QuoteItemType
		if itemType == 0 {
			switch {
			case in.ProductID != 0:
				itemType = quoteItemProduct
			case in.ServiceID != 0:
				itemType = quoteItemService
			case in.ServiceBundleID != 0:
				itemType = quoteItemServiceBundle
			}
		}
		if itemType != 0 {
			payload["quoteItemType"] = itemType
		}

		created, err := autotask.CreateRaw(ctx, client, "QuoteItems", payload)
		if err != nil {
			return errorResult("failed to create quote item: %v", err)
		}

		data, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created quote item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// updateQuoteItemHandler returns a handler that updates an existing quote item.
func updateQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in UpdateQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UpdateQuoteItemInput) (*mcp.CallToolResult, any, error) {
		payload := map[string]any{
			"id": in.QuoteItemID,
		}

		if in.Quantity != nil {
			payload["quantity"] = *in.Quantity
		}
		if in.UnitPrice != nil {
			payload["unitPrice"] = *in.UnitPrice
		}
		if in.UnitDiscount != nil {
			payload["unitDiscount"] = *in.UnitDiscount
		}
		if in.LineDiscount != nil {
			payload["lineDiscount"] = *in.LineDiscount
		}
		if in.PercentageDiscount != nil {
			payload["percentageDiscount"] = *in.PercentageDiscount
		}
		if in.IsOptional != nil {
			payload["isOptional"] = *in.IsOptional
		}
		if in.SortOrderID != nil {
			payload["sortOrderID"] = *in.SortOrderID
		}

		updated, err := autotask.UpdateRaw(ctx, client, "QuoteItems", payload)
		if err != nil {
			return errorResult("failed to update quote item %d: %v", in.QuoteItemID, err)
		}

		data, err := json.MarshalIndent(updated, "", "  ")
		if err != nil {
			return errorResult("failed to marshal updated quote item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// deleteQuoteItemHandler returns a handler that deletes a quote item.
func deleteQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in DeleteQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in DeleteQuoteItemInput) (*mcp.CallToolResult, any, error) {
		if err := autotask.DeleteRaw(ctx, client, "QuoteItems", in.QuoteItemID); err != nil {
			return errorResult("failed to delete quote item %d: %v", in.QuoteItemID, err)
		}

		return textResult("Quote item %d deleted successfully", in.QuoteItemID)
	}
}

// getOpportunityHandler returns a handler that retrieves a single opportunity.
func getOpportunityHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetOpportunityInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetOpportunityInput) (*mcp.CallToolResult, any, error) {
		opp, err := autotask.GetRaw(ctx, client, "Opportunities", in.OpportunityID)
		if err != nil {
			return errorResult("failed to get opportunity %d: %v", in.OpportunityID, err)
		}

		data, err := json.MarshalIndent(opp, "", "  ")
		if err != nil {
			return errorResult("failed to marshal opportunity: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchOpportunitiesHandler returns a handler that searches opportunities.
func searchOpportunitiesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchOpportunitiesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchOpportunitiesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.SearchTerm != "" {
			q.Where("title", autotask.OpContains, in.SearchTerm)
		}
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		}

		opps, err := autotask.ListRaw(ctx, client, "Opportunities", q)
		if err != nil {
			return errorResult("failed to search opportunities: %v", err)
		}

		if len(opps) == 0 {
			return textResult("No opportunities found")
		}

		data, err := json.MarshalIndent(opps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal opportunities: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createOpportunityHandler returns a handler that creates a new opportunity.
func createOpportunityHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateOpportunityInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateOpportunityInput) (*mcp.CallToolResult, any, error) {
		projectedClose, err := parseDate(in.ProjectedCloseDate)
		if err != nil {
			return errorResult("invalid projectedCloseDate format: %v", err)
		}
		startDate, err := parseDate(in.StartDate)
		if err != nil {
			return errorResult("invalid startDate format: %v", err)
		}

		payload := map[string]any{
			"title":               in.Title,
			"companyID":           in.CompanyID,
			"ownerResourceID":     in.OwnerResourceID,
			"status":              in.Status,
			"stage":               in.Stage,
			"projectedCloseDate":  projectedClose.Format("2006-01-02"),
			"startDate":           startDate.Format("2006-01-02"),
		}

		if in.Probability != 0 {
			payload["probability"] = in.Probability
		}
		if in.Amount != 0 {
			payload["amount"] = in.Amount
		}
		if in.Cost != 0 {
			payload["cost"] = in.Cost
		}
		if in.UseQuoteTotals {
			payload["useQuoteTotals"] = in.UseQuoteTotals
		}
		if in.TotalAmountMonths != 0 {
			payload["totalAmountMonths"] = in.TotalAmountMonths
		}
		if in.ContactID != 0 {
			payload["contactID"] = in.ContactID
		}
		if in.Description != "" {
			payload["description"] = in.Description
		}
		if in.OpportunityCategoryID != 0 {
			payload["opportunityCategoryID"] = in.OpportunityCategoryID
		}

		created, err := autotask.CreateRaw(ctx, client, "Opportunities", payload)
		if err != nil {
			return errorResult("failed to create opportunity: %v", err)
		}

		data, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created opportunity: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchInvoicesHandler returns a handler that searches invoices.
func searchInvoicesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchInvoicesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchInvoicesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.InvoiceNumber != "" {
			q.Where("invoiceNumber", autotask.OpEq, in.InvoiceNumber)
		}
		if in.IsVoided != nil {
			q.Where("isVoided", autotask.OpEq, *in.IsVoided)
		}

		invoices, err := autotask.ListRaw(ctx, client, "Invoices", q)
		if err != nil {
			return errorResult("failed to search invoices: %v", err)
		}

		if len(invoices) == 0 {
			return textResult("No invoices found")
		}

		data, err := json.MarshalIndent(invoices, "", "  ")
		if err != nil {
			return errorResult("failed to marshal invoices: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchContractsHandler returns a handler that searches contracts.
func searchContractsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchContractsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchContractsInput) (*mcp.CallToolResult, any, error) {
		page := 1
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("contractName", autotask.OpContains, in.SearchTerm)
		}
		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		}

		contracts, err := autotask.List[entities.Contract](ctx, client, q)
		if err != nil {
			return errorResult("failed to search contracts: %v", err)
		}

		if len(contracts) == 0 {
			return textResult("No contracts found")
		}

		maps, err := entitiesToMaps(contracts)
		if err != nil {
			return errorResult("failed to convert contracts: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_contracts", page, pageSize)
	}
}
