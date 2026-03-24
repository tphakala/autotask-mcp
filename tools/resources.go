package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchResourcesInput defines the input parameters for searching resources.
type SearchResourcesInput struct {
	SearchTerm   string `json:"searchTerm,omitempty" jsonschema:"Search term for resource name or email"`
	IsActive     *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	ResourceType int    `json:"resourceType,omitempty" jsonschema:"Filter by resource type (1=Employee, 2=Contractor, 3=Temporary)"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize     int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterResourceTools registers all resource-related MCP tools with the server.
func RegisterResourceTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_resources",
		Description: "Search for resources (employees/contractors) in Autotask. Returns 25 results per page by default.",
	}, searchResourcesHandler(client))
}

// searchResourcesHandler returns a handler that searches resources using the provided filters.
func searchResourcesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchResourcesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchResourcesInput) (*mcp.CallToolResult, any, error) {
		pageSize := defaultPageSize(in.PageSize, 25, 500)

		q := autotask.NewQuery().Limit(pageSize)

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
		if in.ResourceType != 0 {
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

		data, err := json.Marshal(maps)
		if err != nil {
			return errorResult("failed to marshal resources: %v", err)
		}

		return textResult("%s", string(data))
	}
}
