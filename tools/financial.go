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
	Title                 string  `json:"title" jsonschema:"Opportunity title"`
	CompanyID             int64   `json:"companyId" jsonschema:"Company ID"`
	OwnerResourceID       int64   `json:"ownerResourceId" jsonschema:"Owner resource ID"`
	Status                int     `json:"status" jsonschema:"Status ID"`
	Stage                 int     `json:"stage" jsonschema:"Stage ID"`
	ProjectedCloseDate    string  `json:"projectedCloseDate" jsonschema:"Projected close date (YYYY-MM-DD or ISO format)"`
	StartDate             string  `json:"startDate" jsonschema:"Start date (YYYY-MM-DD or ISO format)"`
	Probability           int     `json:"probability,omitempty" jsonschema:"Win probability percentage"`
	Amount                float64 `json:"amount,omitempty" jsonschema:"Opportunity amount"`
	Cost                  float64 `json:"cost,omitempty" jsonschema:"Opportunity cost"`
	UseQuoteTotals        bool    `json:"useQuoteTotals,omitempty" jsonschema:"Whether to use quote totals"`
	TotalAmountMonths     int     `json:"totalAmountMonths,omitempty" jsonschema:"Total amount months"`
	ContactID             int64   `json:"contactId,omitempty" jsonschema:"Contact ID"`
	Description           string  `json:"description,omitempty" jsonschema:"Description"`
	OpportunityCategoryID int     `json:"opportunityCategoryID,omitempty" jsonschema:"Opportunity category ID"`
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
		quote, err := autotask.Get[entities.Quote](ctx, client, in.QuoteID)
		if err != nil {
			return errorResult("failed to get quote %d: %v", in.QuoteID, err)
		}

		m, err := entityToMap(quote)
		if err != nil {
			return errorResult("failed to convert quote: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		quotes, err := autotask.List[entities.Quote](ctx, client, q)
		if err != nil {
			return errorResult("failed to search quotes: %v", err)
		}

		if len(quotes) == 0 {
			return textResult("No quotes found")
		}

		maps, err := entitiesToMaps(quotes)
		if err != nil {
			return errorResult("failed to convert quotes: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quotes: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createQuoteHandler returns a handler that creates a new quote.
func createQuoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteInput) (*mcp.CallToolResult, any, error) {
		entity := &entities.Quote{
			CompanyID: autotask.Set(in.CompanyID),
		}
		if in.Name != "" {
			entity.Name = autotask.Set(in.Name)
		}
		if in.Description != "" {
			entity.Description = autotask.Set(in.Description)
		}
		if in.ContactID != 0 {
			entity.ContactID = autotask.Set(in.ContactID)
		}
		if in.OpportunityID != 0 {
			entity.OpportunityID = autotask.Set(in.OpportunityID)
		}
		if in.EffectiveDate != "" {
			t, err := parseDate(in.EffectiveDate)
			if err != nil {
				return errorResult("invalid effectiveDate format: %v", err)
			}
			entity.EffectiveDate = autotask.Set(t)
		}
		if in.ExpirationDate != "" {
			t, err := parseDate(in.ExpirationDate)
			if err != nil {
				return errorResult("invalid expirationDate format: %v", err)
			}
			entity.ExpirationDate = autotask.Set(t)
		}

		created, err := autotask.Create[entities.Quote](ctx, client, entity)
		if err != nil {
			return errorResult("failed to create quote: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created quote: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created quote: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getQuoteItemHandler returns a handler that retrieves a single quote item.
func getQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetQuoteItemInput) (*mcp.CallToolResult, any, error) {
		item, err := autotask.Get[entities.QuoteItem](ctx, client, in.QuoteItemID)
		if err != nil {
			return errorResult("failed to get quote item %d: %v", in.QuoteItemID, err)
		}

		m, err := entityToMap(item)
		if err != nil {
			return errorResult("failed to convert quote item: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		items, err := autotask.List[entities.QuoteItem](ctx, client, q)
		if err != nil {
			return errorResult("failed to search quote items: %v", err)
		}

		if len(items) == 0 {
			return textResult("No quote items found")
		}

		maps, err := entitiesToMaps(items)
		if err != nil {
			return errorResult("failed to convert quote items: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal quote items: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createQuoteItemHandler returns a handler that creates a new quote item.
func createQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateQuoteItemInput) (*mcp.CallToolResult, any, error) {
		entity := &entities.QuoteItem{
			QuoteID:  autotask.Set(in.QuoteID),
			Quantity: autotask.Set(in.Quantity),
		}

		if in.Name != "" {
			entity.Name = autotask.Set(in.Name)
		}
		if in.Description != "" {
			entity.Description = autotask.Set(in.Description)
		}
		if in.UnitPrice != 0 {
			entity.UnitPrice = autotask.Set(in.UnitPrice)
		}
		if in.UnitCost != 0 {
			entity.UnitCost = autotask.Set(in.UnitCost)
		}
		if in.UnitDiscount != 0 {
			entity.UnitDiscount = autotask.Set(in.UnitDiscount)
		}
		if in.LineDiscount != 0 {
			entity.LineDiscount = autotask.Set(in.LineDiscount)
		}
		if in.PercentageDiscount != 0 {
			entity.PercentageDiscount = autotask.Set(in.PercentageDiscount)
		}
		if in.IsOptional != nil {
			entity.IsOptional = autotask.Set(*in.IsOptional)
		}
		if in.ProductID != 0 {
			entity.ProductID = autotask.Set(in.ProductID)
		}
		if in.ServiceID != 0 {
			entity.ServiceID = autotask.Set(in.ServiceID)
		}
		if in.ServiceBundleID != 0 {
			entity.ServiceBundleID = autotask.Set(in.ServiceBundleID)
		}
		if in.SortOrderID != 0 {
			entity.SortOrderID = autotask.Set(int64(in.SortOrderID))
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
			entity.QuoteItemType = autotask.Set(int64(itemType))
		}

		created, err := autotask.Create[entities.QuoteItem](ctx, client, entity)
		if err != nil {
			return errorResult("failed to create quote item: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created quote item: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created quote item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// updateQuoteItemHandler returns a handler that updates an existing quote item.
func updateQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in UpdateQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UpdateQuoteItemInput) (*mcp.CallToolResult, any, error) {
		entity := &entities.QuoteItem{
			ID: autotask.Set(in.QuoteItemID),
		}

		if in.Quantity != nil {
			entity.Quantity = autotask.Set(*in.Quantity)
		}
		if in.UnitPrice != nil {
			entity.UnitPrice = autotask.Set(*in.UnitPrice)
		}
		if in.UnitDiscount != nil {
			entity.UnitDiscount = autotask.Set(*in.UnitDiscount)
		}
		if in.LineDiscount != nil {
			entity.LineDiscount = autotask.Set(*in.LineDiscount)
		}
		if in.PercentageDiscount != nil {
			entity.PercentageDiscount = autotask.Set(*in.PercentageDiscount)
		}
		if in.IsOptional != nil {
			entity.IsOptional = autotask.Set(*in.IsOptional)
		}
		if in.SortOrderID != nil {
			entity.SortOrderID = autotask.Set(int64(*in.SortOrderID))
		}

		updated, err := autotask.Update[entities.QuoteItem](ctx, client, entity)
		if err != nil {
			return errorResult("failed to update quote item %d: %v", in.QuoteItemID, err)
		}

		m, err := entityToMap(updated)
		if err != nil {
			return errorResult("failed to convert updated quote item: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal updated quote item: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// deleteQuoteItemHandler returns a handler that deletes a quote item.
func deleteQuoteItemHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in DeleteQuoteItemInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in DeleteQuoteItemInput) (*mcp.CallToolResult, any, error) {
		if err := autotask.Delete[entities.QuoteItem](ctx, client, in.QuoteItemID); err != nil {
			return errorResult("failed to delete quote item %d: %v", in.QuoteItemID, err)
		}

		return textResult("Quote item %d deleted successfully", in.QuoteItemID)
	}
}

// getOpportunityHandler returns a handler that retrieves a single opportunity.
func getOpportunityHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetOpportunityInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetOpportunityInput) (*mcp.CallToolResult, any, error) {
		opp, err := autotask.Get[entities.Opportunity](ctx, client, in.OpportunityID)
		if err != nil {
			return errorResult("failed to get opportunity %d: %v", in.OpportunityID, err)
		}

		m, err := entityToMap(opp)
		if err != nil {
			return errorResult("failed to convert opportunity: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		opps, err := autotask.List[entities.Opportunity](ctx, client, q)
		if err != nil {
			return errorResult("failed to search opportunities: %v", err)
		}

		if len(opps) == 0 {
			return textResult("No opportunities found")
		}

		maps, err := entitiesToMaps(opps)
		if err != nil {
			return errorResult("failed to convert opportunities: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
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

		entity := &entities.Opportunity{
			Title:              autotask.Set(in.Title),
			CompanyID:          autotask.Set(in.CompanyID),
			OwnerResourceID:    autotask.Set(in.OwnerResourceID),
			Status:             autotask.Set(int64(in.Status)),
			Stage:              autotask.Set(int64(in.Stage)),
			ProjectedCloseDate: autotask.Set(projectedClose),
			StartDate:          autotask.Set(startDate),
		}

		if in.Probability != 0 {
			entity.Probability = autotask.Set(int64(in.Probability))
		}
		if in.Amount != 0 {
			entity.Amount = autotask.Set(in.Amount)
		}
		if in.Cost != 0 {
			entity.Cost = autotask.Set(in.Cost)
		}
		if in.UseQuoteTotals {
			entity.UseQuoteTotals = autotask.Set(in.UseQuoteTotals)
		}
		if in.TotalAmountMonths != 0 {
			entity.TotalAmountMonths = autotask.Set(int64(in.TotalAmountMonths))
		}
		if in.ContactID != 0 {
			entity.ContactID = autotask.Set(in.ContactID)
		}
		if in.Description != "" {
			entity.Description = autotask.Set(in.Description)
		}
		if in.OpportunityCategoryID != 0 {
			entity.OpportunityCategoryID = autotask.Set(int64(in.OpportunityCategoryID))
		}

		created, err := autotask.Create[entities.Opportunity](ctx, client, entity)
		if err != nil {
			return errorResult("failed to create opportunity: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created opportunity: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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

		invoices, err := autotask.List[entities.Invoice](ctx, client, q)
		if err != nil {
			return errorResult("failed to search invoices: %v", err)
		}

		if len(invoices) == 0 {
			return textResult("No invoices found")
		}

		maps, err := entitiesToMaps(invoices)
		if err != nil {
			return errorResult("failed to convert invoices: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
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
