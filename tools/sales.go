package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetProductInput defines the input parameters for getting a product.
type GetProductInput struct {
	ProductID int64 `json:"productId" jsonschema:"Product ID to retrieve"`
}

// SearchProductsInput defines the input parameters for searching products.
type SearchProductsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by product name (partial match)"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// GetServiceInput defines the input parameters for getting a service.
type GetServiceInput struct {
	ServiceID int64 `json:"serviceId" jsonschema:"Service ID to retrieve"`
}

// SearchServicesInput defines the input parameters for searching services.
type SearchServicesInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by service name (partial match)"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// GetServiceBundleInput defines the input parameters for getting a service bundle.
type GetServiceBundleInput struct {
	ServiceBundleID int64 `json:"serviceBundleId" jsonschema:"Service bundle ID to retrieve"`
}

// SearchServiceBundlesInput defines the input parameters for searching service bundles.
type SearchServiceBundlesInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by service bundle name (partial match)"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// RegisterSalesTools registers all sales-related MCP tools with the server.
func RegisterSalesTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_product",
		Description: "Retrieve one catalog product by its numeric ID, returning the full field set. A product is a one-off sellable item such as hardware, a software license, or materials. Use when you already have a product ID; to find products by name or active status use autotask_search_products instead. Read-only.",
		Annotations: readOnlyTool("Get product"),
	}, getProductHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_products",
		Description: "Find catalog products (one-off sellable items such as hardware, licenses, or materials) by name substring and active status, returning up to maxResults (default 25, max 500). Use this to locate a product, then autotask_get_product for the full field set of one by ID. Distinct from autotask_search_services, which lists recurring billable services rather than one-off products. Read-only.",
		Annotations: readOnlyTool("Search products"),
	}, searchProductsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_service",
		Description: "Retrieve one catalog service by its numeric ID, returning the full field set. A service is a recurring, periodically billed offering, unlike a one-off product. Use when you already have a service ID; to find services by name or active status use autotask_search_services instead. Read-only.",
		Annotations: readOnlyTool("Get service"),
	}, getServiceHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_services",
		Description: "Find recurring billable services by name substring and active status, returning up to maxResults (default 25, max 500). Services are periodically billed offerings, unlike one-off products (autotask_search_products) or grouped service bundles (autotask_search_service_bundles). Use this to locate a service, then autotask_get_service for the full field set of one by ID. Read-only.",
		Annotations: readOnlyTool("Search services"),
	}, searchServicesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_service_bundle",
		Description: "Retrieve one service bundle by its numeric ID, returning the full field set. A service bundle groups several individual services sold together as a single line item. Use when you already have a bundle ID; to find bundles by name or active status use autotask_search_service_bundles instead. Read-only.",
		Annotations: readOnlyTool("Get service bundle"),
	}, getServiceBundleHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_service_bundles",
		Description: "Find service bundles (groups of individual services sold together as one line item) by name substring and active status, returning up to maxResults (default 25, max 500). Use this to locate a bundle, then autotask_get_service_bundle for the full field set of one by ID; for the individual services that make up bundles see autotask_search_services. Read-only.",
		Annotations: readOnlyTool("Search service bundles"),
	}, searchServiceBundlesHandler(client))
}

// getProductHandler returns a handler that retrieves a single product.
func getProductHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetProductInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetProductInput) (*mcp.CallToolResult, map[string]any, error) {
		product, err := autotask.Get[entities.Product](ctx, client, in.ProductID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(product)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchProductsHandler returns a handler that searches products.
func searchProductsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProductsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProductsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)
		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Where("name", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		products, err := autotask.List[entities.Product](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(products) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(products)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, nil, maps, "autotask_search_products", maxResults)
	}
}

// getServiceHandler returns a handler that retrieves a single service.
func getServiceHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceInput) (*mcp.CallToolResult, map[string]any, error) {
		service, err := autotask.Get[entities.Service](ctx, client, in.ServiceID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(service)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchServicesHandler returns a handler that searches services.
func searchServicesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchServicesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchServicesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)
		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Where("serviceName", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		svcs, err := autotask.List[entities.Service](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(svcs) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(svcs)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, nil, maps, "autotask_search_services", maxResults)
	}
}

// getServiceBundleHandler returns a handler that retrieves a single service bundle.
func getServiceBundleHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceBundleInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceBundleInput) (*mcp.CallToolResult, map[string]any, error) {
		bundle, err := autotask.Get[entities.ServiceBundle](ctx, client, in.ServiceBundleID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(bundle)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchServiceBundlesHandler returns a handler that searches service bundles.
func searchServiceBundlesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchServiceBundlesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchServiceBundlesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)
		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Where("serviceBundleName", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		bundles, err := autotask.List[entities.ServiceBundle](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(bundles) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(bundles)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, nil, maps, "autotask_search_service_bundles", maxResults)
	}
}
