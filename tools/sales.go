package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// GetServiceInput defines the input parameters for getting a service.
type GetServiceInput struct {
	ServiceID int64 `json:"serviceId" jsonschema:"Service ID to retrieve"`
}

// SearchServicesInput defines the input parameters for searching services.
type SearchServicesInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by service name (partial match)"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// GetServiceBundleInput defines the input parameters for getting a service bundle.
type GetServiceBundleInput struct {
	ServiceBundleID int64 `json:"serviceBundleId" jsonschema:"Service bundle ID to retrieve"`
}

// SearchServiceBundlesInput defines the input parameters for searching service bundles.
type SearchServiceBundlesInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search by service bundle name (partial match)"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
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
		Description: "Find catalog products (one-off sellable items such as hardware, licenses, or materials) by name substring and active status, returning up to 25 per page (max 500). Use this to locate a product, then autotask_get_product for the full field set of one by ID. Distinct from autotask_search_services, which lists recurring billable services rather than one-off products. Read-only.",
		Annotations: readOnlyTool("Search products"),
	}, searchProductsHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_service",
		Description: "Retrieve one catalog service by its numeric ID, returning the full field set. A service is a recurring, periodically billed offering, unlike a one-off product. Use when you already have a service ID; to find services by name or active status use autotask_search_services instead. Read-only.",
		Annotations: readOnlyTool("Get service"),
	}, getServiceHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_services",
		Description: "Find recurring billable services by name substring and active status, returning up to 25 per page (max 500). Services are periodically billed offerings, unlike one-off products (autotask_search_products) or grouped service bundles (autotask_search_service_bundles). Use this to locate a service, then autotask_get_service for the full field set of one by ID. Read-only.",
		Annotations: readOnlyTool("Search services"),
	}, searchServicesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_service_bundle",
		Description: "Retrieve one service bundle by its numeric ID, returning the full field set. A service bundle groups several individual services sold together as a single line item. Use when you already have a bundle ID; to find bundles by name or active status use autotask_search_service_bundles instead. Read-only.",
		Annotations: readOnlyTool("Get service bundle"),
	}, getServiceBundleHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_service_bundles",
		Description: "Find service bundles (groups of individual services sold together as one line item) by name substring and active status, returning up to 25 per page (max 500). Use this to locate a bundle, then autotask_get_service_bundle for the full field set of one by ID; for the individual services that make up bundles see autotask_search_services. Read-only.",
		Annotations: readOnlyTool("Search service bundles"),
	}, searchServiceBundlesHandler(client))
}

// getProductHandler returns a handler that retrieves a single product.
func getProductHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetProductInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetProductInput) (*mcp.CallToolResult, any, error) {
		product, err := autotask.Get[entities.Product](ctx, client, in.ProductID)
		if err != nil {
			return errorResult("failed to get product %d: %v", in.ProductID, err)
		}

		m, err := entityToMap(product)
		if err != nil {
			return errorResult("failed to convert product: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal product: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchProductsHandler returns a handler that searches products.
func searchProductsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProductsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProductsInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("name", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		products, err := autotask.List[entities.Product](ctx, client, q)
		if err != nil {
			return errorResult("failed to search products: %v", err)
		}

		if len(products) == 0 {
			return textResult("No products found")
		}

		maps, err := entitiesToMaps(products)
		if err != nil {
			return errorResult("failed to convert products: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal products: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getServiceHandler returns a handler that retrieves a single service.
func getServiceHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceInput) (*mcp.CallToolResult, any, error) {
		service, err := autotask.Get[entities.Service](ctx, client, in.ServiceID)
		if err != nil {
			return errorResult("failed to get service %d: %v", in.ServiceID, err)
		}

		m, err := entityToMap(service)
		if err != nil {
			return errorResult("failed to convert service: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal service: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchServicesHandler returns a handler that searches services.
func searchServicesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchServicesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchServicesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("serviceName", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		services, err := autotask.List[entities.Service](ctx, client, q)
		if err != nil {
			return errorResult("failed to search services: %v", err)
		}

		if len(services) == 0 {
			return textResult("No services found")
		}

		maps, err := entitiesToMaps(services)
		if err != nil {
			return errorResult("failed to convert services: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal services: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getServiceBundleHandler returns a handler that retrieves a single service bundle.
func getServiceBundleHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceBundleInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetServiceBundleInput) (*mcp.CallToolResult, any, error) {
		bundle, err := autotask.Get[entities.ServiceBundle](ctx, client, in.ServiceBundleID)
		if err != nil {
			return errorResult("failed to get service bundle %d: %v", in.ServiceBundleID, err)
		}

		m, err := entityToMap(bundle)
		if err != nil {
			return errorResult("failed to convert service bundle: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal service bundle: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchServiceBundlesHandler returns a handler that searches service bundles.
func searchServiceBundlesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchServiceBundlesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchServiceBundlesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)
		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("serviceBundleName", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		bundles, err := autotask.List[entities.ServiceBundle](ctx, client, q)
		if err != nil {
			return errorResult("failed to search service bundles: %v", err)
		}

		if len(bundles) == 0 {
			return textResult("No service bundles found")
		}

		maps, err := entitiesToMaps(bundles)
		if err != nil {
			return errorResult("failed to convert service bundles: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal service bundles: %v", err)
		}

		return textResult("%s", string(data))
	}
}
