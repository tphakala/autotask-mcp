package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/autotask-mcp/services"
)

// GetFieldInfoInput defines the input parameters for getting field info.
type GetFieldInfoInput struct {
	EntityType string `json:"entityType" jsonschema:"Entity type name (e.g. Tickets, Companies)"`
	FieldName  string `json:"fieldName,omitempty" jsonschema:"Specific field name to retrieve info for"`
}

// RegisterPicklistTools registers all picklist-related MCP tools with the server.
func RegisterPicklistTools(s *mcp.Server, client *autotask.Client, picklist *services.PicklistCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_queues",
		Description: "List all ticket queue picklist values.",
	}, listQueuesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_ticket_statuses",
		Description: "List all ticket status picklist values.",
	}, listTicketStatusesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_ticket_priorities",
		Description: "List all ticket priority picklist values.",
	}, listTicketPrioritiesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_field_info",
		Description: "Get field metadata for an Autotask entity type. Optionally filter to a specific field.",
	}, getFieldInfoHandler(picklist))
}

// listQueuesHandler returns a handler that lists all queue picklist values.
func listQueuesHandler(picklist *services.PicklistCache) func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		values, err := picklist.GetPicklistValues(ctx, "Tickets", "queueID")
		if err != nil {
			return errorResult("failed to get queue picklist values: %v", err)
		}

		data, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			return errorResult("failed to marshal queue values: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// listTicketStatusesHandler returns a handler that lists all ticket status picklist values.
func listTicketStatusesHandler(picklist *services.PicklistCache) func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		values, err := picklist.GetPicklistValues(ctx, "Tickets", "status")
		if err != nil {
			return errorResult("failed to get ticket status picklist values: %v", err)
		}

		data, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket status values: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// listTicketPrioritiesHandler returns a handler that lists all ticket priority picklist values.
func listTicketPrioritiesHandler(picklist *services.PicklistCache) func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		values, err := picklist.GetPicklistValues(ctx, "Tickets", "priority")
		if err != nil {
			return errorResult("failed to get ticket priority picklist values: %v", err)
		}

		data, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket priority values: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getFieldInfoHandler returns a handler that retrieves field metadata for an entity type.
func getFieldInfoHandler(picklist *services.PicklistCache) func(ctx context.Context, req *mcp.CallToolRequest, in GetFieldInfoInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetFieldInfoInput) (*mcp.CallToolResult, any, error) {
		fields, err := picklist.GetFields(ctx, in.EntityType)
		if err != nil {
			return errorResult("failed to get fields for entity %q: %v", in.EntityType, err)
		}

		// Filter to a specific field if requested.
		if in.FieldName != "" {
			for _, f := range fields {
				if f.Name == in.FieldName {
					data, err := json.MarshalIndent(f, "", "  ")
					if err != nil {
						return errorResult("failed to marshal field info: %v", err)
					}
					return textResult("%s", string(data))
				}
			}
			return textResult("Field %q not found on entity %q", in.FieldName, in.EntityType)
		}

		data, err := json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return errorResult("failed to marshal fields: %v", err)
		}

		return textResult("%s", string(data))
	}
}
