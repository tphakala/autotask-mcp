package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchResourcesInput defines the input parameters for searching resources.
type SearchResourcesInput struct {
	SearchTerm   string `json:"searchTerm,omitempty" jsonschema:"Search term for resource name or email"`
	IsActive     *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	ResourceType string `json:"resourceType,omitempty" jsonschema:"Filter by resource type (Employee, Contractor, Temporary)"`
	MaxResults   int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 500)"`
}

// RegisterResourceTools registers all resource-related MCP tools with the server.
func RegisterResourceTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_resources",
		Description: "Find internal staff (employees, contractors, or temporary workers) of the Autotask account by name or email substring, active status, and resource type, returning a compact summary of matching records (up to maxResults, default 25, max 500). Resources are the people assigned to tickets and tasks; for client-side people at a company use autotask_search_contacts instead. Use the returned resource ID as assignedResourceID when creating or updating tickets. Read-only.",
		Annotations: readOnlyTool("Search resources"),
	}, searchResourcesHandler(client))
}

// searchResourcesHandler returns a handler that searches resources using the provided filters.
func searchResourcesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchResourcesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchResourcesInput) (*mcp.CallToolResult, any, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 500)

		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Or(
				autotask.Field("firstName", autotask.OpContains, in.SearchTerm),
				autotask.Field("lastName", autotask.OpContains, in.SearchTerm),
				autotask.Field("email", autotask.OpContains, in.SearchTerm),
			)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}
		if in.ResourceType != "" {
			q.Where("resourceType", autotask.OpEq, in.ResourceType)
		}

		resources, err := autotask.List[entities.Resource](ctx, client, q)
		if err != nil {
			return errorResult("failed to search resources: %v", err)
		}

		if len(resources) == 0 {
			return textResult("No resources found")
		}

		maps, err := entitiesToMaps(resources)
		if err != nil {
			return errorResult("failed to convert resources: %v", err)
		}

		return searchResult(ctx, nil, maps, "autotask_search_resources", maxResults)
	}
}
