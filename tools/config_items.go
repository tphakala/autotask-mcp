package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchConfigurationItemsInput defines the input parameters for searching configuration items.
type SearchConfigurationItemsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by reference title (partial match)"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	ProductID  int64  `json:"productID,omitempty" jsonschema:"Filter by product ID"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterConfigItemTools registers all configuration item MCP tools with the server.
func RegisterConfigItemTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_configuration_items",
		Description: "Find configuration items (CIs: assets or installed products tracked against a company) by reference-title substring, company, active status, or product, returning a compact summary of up to 25 (max 500). CIs link a company to the products it owns; filter by companyID to list one company's assets or by productID (from autotask_search_products) to find every install of a product. Read-only.",
		Annotations: readOnlyTool("Search configuration items"),
	}, searchConfigurationItemsHandler(client, mapper))
}

// searchConfigurationItemsHandler returns a handler that searches configuration items.
func searchConfigurationItemsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchConfigurationItemsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchConfigurationItemsInput) (*mcp.CallToolResult, any, error) {
		page := 1
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("referenceTitle", autotask.OpContains, in.SearchTerm)
		}
		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}
		if in.ProductID != 0 {
			q.Where("productID", autotask.OpEq, in.ProductID)
		}

		items, err := autotask.List[entities.ConfigurationItem](ctx, client, q)
		if err != nil {
			return errorResult("failed to search configuration items: %v", err)
		}

		if len(items) == 0 {
			return textResult("No configuration items found")
		}

		maps, err := entitiesToMaps(items)
		if err != nil {
			return errorResult("failed to convert configuration items: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_configuration_items", page, pageSize)
	}
}
