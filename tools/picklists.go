package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
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
		Description: "Return the id and label of every ticket queue picklist value from the Tickets entity queueID field. Use these ids to resolve a queue by name when routing or filtering tickets; for ticket status or priority values call autotask_list_ticket_statuses or autotask_list_ticket_priorities instead. Served from cached Autotask field metadata. Read-only.",
		Annotations: readOnlyTool("List queues"),
	}, listQueuesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_ticket_statuses",
		Description: "Return the id and label of every ticket status picklist value from the Tickets entity status field. Use these ids for the status filter of autotask_search_tickets or the status field of autotask_create_ticket and autotask_update_ticket, which all take a numeric status ID; status 5 is the completed state that autotask_search_tickets excludes by default. For queues or priorities call autotask_list_queues or autotask_list_ticket_priorities. Read-only.",
		Annotations: readOnlyTool("List ticket statuses"),
	}, listTicketStatusesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_ticket_priorities",
		Description: "Return the id and label of every ticket priority picklist value from the Tickets entity priority field. Use these ids to set the priority field on autotask_create_ticket or autotask_update_ticket, which require a numeric priority ID; for statuses or queues call autotask_list_ticket_statuses or autotask_list_queues. Read-only.",
		Annotations: readOnlyTool("List ticket priorities"),
	}, listTicketPrioritiesHandler(picklist))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_field_info",
		Description: "Return field metadata for one Autotask entity type such as Tickets or Companies, including each field's data type, requiredness, and picklist options, for every field or for a single field when fieldName is supplied. Use this to discover valid fields and allowed values on any entity before searching, creating, or updating it; for the common ticket picklists call autotask_list_ticket_statuses, autotask_list_ticket_priorities, or autotask_list_queues directly. Requires entityType and returns all fields unless fieldName matches one. Read-only.",
		Annotations: readOnlyTool("Get field info"),
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
