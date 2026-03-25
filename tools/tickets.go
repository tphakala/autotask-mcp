package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// SearchTicketsInput defines the input parameters for searching tickets.
type SearchTicketsInput struct {
	SearchTerm         string `json:"searchTerm,omitempty" jsonschema:"Search by ticket number prefix"`
	CompanyID          int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	Status             int    `json:"status,omitempty" jsonschema:"Filter by ticket status ID (omit for all open tickets)"`
	AssignedResourceID int64  `json:"assignedResourceID,omitempty" jsonschema:"Filter by assigned resource ID"`
	Unassigned         bool   `json:"unassigned,omitempty" jsonschema:"Set to true to find unassigned tickets"`
	CreatedAfter       string `json:"createdAfter,omitempty" jsonschema:"Filter tickets created on or after this date (ISO format)"`
	CreatedBefore      string `json:"createdBefore,omitempty" jsonschema:"Filter tickets created on or before this date (ISO format)"`
	LastActivityAfter  string `json:"lastActivityAfter,omitempty" jsonschema:"Filter tickets with activity on or after this date (ISO format)"`
	Page               int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize           int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// GetTicketDetailsInput defines the input parameters for retrieving a single ticket.
type GetTicketDetailsInput struct {
	TicketID int64 `json:"ticketID" jsonschema:"Ticket ID to retrieve"`
	// FullDetails is included for API surface consistency; all fields are always returned.
	FullDetails bool `json:"fullDetails,omitempty" jsonschema:"Whether to return full ticket details (default false)"`
}

// CreateTicketInput defines the input parameters for creating a new ticket.
type CreateTicketInput struct {
	CompanyID              int64  `json:"companyID" jsonschema:"Company ID for the ticket"`
	Title                  string `json:"title" jsonschema:"Ticket title"`
	Description            string `json:"description" jsonschema:"Ticket description"`
	Status                 int    `json:"status,omitempty" jsonschema:"Ticket status ID"`
	Priority               int    `json:"priority,omitempty" jsonschema:"Ticket priority ID"`
	AssignedResourceID     int64  `json:"assignedResourceID,omitempty" jsonschema:"Assigned resource ID"`
	AssignedResourceRoleID int64  `json:"assignedResourceRoleID,omitempty" jsonschema:"Role ID for the assigned resource"`
	ContactID              int64  `json:"contactID,omitempty" jsonschema:"Contact ID for the ticket"`
}

// UpdateTicketInput defines the input parameters for updating an existing ticket.
type UpdateTicketInput struct {
	TicketID               int64  `json:"ticketId" jsonschema:"The ID of the ticket to update"`
	Title                  string `json:"title,omitempty" jsonschema:"Ticket title"`
	Description            string `json:"description,omitempty" jsonschema:"Ticket description"`
	Status                 int    `json:"status,omitempty" jsonschema:"Ticket status ID"`
	Priority               int    `json:"priority,omitempty" jsonschema:"Ticket priority ID"`
	AssignedResourceID     int64  `json:"assignedResourceID,omitempty" jsonschema:"Assigned resource ID"`
	AssignedResourceRoleID int64  `json:"assignedResourceRoleID,omitempty" jsonschema:"Role ID for the assigned resource"`
	DueDateTime            string `json:"dueDateTime,omitempty" jsonschema:"Due date/time in ISO 8601 format"`
	ContactID              int64  `json:"contactID,omitempty" jsonschema:"Contact ID for the ticket"`
}

// RegisterTicketTools registers all ticket-related MCP tools with the server.
func RegisterTicketTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_tickets",
		Description: "Search for tickets in Autotask. Returns 25 results per page. Use get_ticket_details for full data.",
	}, searchTicketsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_details",
		Description: "Get detailed information for a specific ticket by ID.",
	}, getTicketDetailsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_ticket",
		Description: "Create a new ticket in Autotask.",
	}, createTicketHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_update_ticket",
		Description: "Update an existing ticket. Only provided fields are changed.",
	}, updateTicketHandler(client))
}

// searchTicketsHandler returns a handler that searches tickets using the provided filters.
func searchTicketsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketsInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 500)

		// Autotask uses cursor-based pagination (NextPageURL), not offset-based.
		// pageSize controls how many records to fetch; page is used only for
		// metadata in the compact response (so callers can track which page they are on).
		q := autotask.NewQuery().Limit(pageSize)

		// Default: exclude completed tickets (status 5) unless a specific status is requested.
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		} else {
			q.Where("status", autotask.OpNotEq, 5)
		}

		if in.SearchTerm != "" {
			q.Where("ticketNumber", autotask.OpBeginsWith, in.SearchTerm)
		}
		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.AssignedResourceID != 0 {
			q.Where("assignedResourceID", autotask.OpEq, in.AssignedResourceID)
		}
		if in.Unassigned {
			q.Where("assignedResourceID", autotask.OpNotExist, nil)
		}
		if in.CreatedAfter != "" {
			q.Where("createDate", autotask.OpGte, in.CreatedAfter)
		}
		if in.CreatedBefore != "" {
			q.Where("createDate", autotask.OpLte, in.CreatedBefore)
		}
		if in.LastActivityAfter != "" {
			q.Where("lastActivityDate", autotask.OpGte, in.LastActivityAfter)
		}

		tickets, err := autotask.List[entities.Ticket](ctx, client, q)
		if err != nil {
			return errorResult("failed to search tickets: %v", err)
		}

		if len(tickets) == 0 {
			return textResult("No tickets found")
		}

		maps, err := entitiesToMaps(tickets)
		if err != nil {
			return errorResult("failed to convert tickets: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_tickets", page, pageSize)
	}
}

// getTicketDetailsHandler returns a handler that retrieves a single ticket by ID.
func getTicketDetailsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketDetailsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketDetailsInput) (*mcp.CallToolResult, any, error) {
		ticket, err := autotask.Get[entities.Ticket](ctx, client, in.TicketID)
		if err != nil {
			return errorResult("failed to get ticket %d: %v", in.TicketID, err)
		}

		m, err := entityToMap(ticket)
		if err != nil {
			return errorResult("failed to convert ticket: %v", err)
		}

		if mapper != nil {
			mapper.EnhanceItems(ctx, []map[string]any{m})
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createTicketHandler returns a handler that creates a new ticket.
func createTicketHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketInput) (*mcp.CallToolResult, any, error) {
		ticket := &entities.Ticket{
			CompanyID:   autotask.Set(in.CompanyID),
			Title:       autotask.Set(in.Title),
			Description: autotask.Set(in.Description),
		}

		if in.Status != 0 {
			ticket.Status = autotask.Set(int64(in.Status))
		}
		if in.Priority != 0 {
			ticket.Priority = autotask.Set(int64(in.Priority))
		}
		if in.AssignedResourceID != 0 {
			ticket.AssignedResourceID = autotask.Set(in.AssignedResourceID)
		}
		if in.AssignedResourceRoleID != 0 {
			ticket.AssignedResourceRoleID = autotask.Set(in.AssignedResourceRoleID)
		}
		if in.ContactID != 0 {
			ticket.ContactID = autotask.Set(in.ContactID)
		}

		created, err := autotask.Create[entities.Ticket](ctx, client, ticket)
		if err != nil {
			return errorResult("failed to create ticket: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created ticket: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created ticket: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// updateTicketHandler returns a handler that updates an existing ticket.
func updateTicketHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in UpdateTicketInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UpdateTicketInput) (*mcp.CallToolResult, any, error) {
		ticket := &entities.Ticket{
			ID: autotask.Set(in.TicketID),
		}

		if in.Title != "" {
			ticket.Title = autotask.Set(in.Title)
		}
		if in.Description != "" {
			ticket.Description = autotask.Set(in.Description)
		}
		if in.Status != 0 {
			ticket.Status = autotask.Set(int64(in.Status))
		}
		if in.Priority != 0 {
			ticket.Priority = autotask.Set(int64(in.Priority))
		}
		if in.AssignedResourceID != 0 {
			ticket.AssignedResourceID = autotask.Set(in.AssignedResourceID)
		}
		if in.AssignedResourceRoleID != 0 {
			ticket.AssignedResourceRoleID = autotask.Set(in.AssignedResourceRoleID)
		}
		if in.DueDateTime != "" {
			t, err := parseDate(in.DueDateTime)
			if err != nil {
				return errorResult("invalid dueDateTime format (expected YYYY-MM-DD or RFC3339): %v", err)
			}
			ticket.DueDateTime = autotask.Set(t)
		}
		if in.ContactID != 0 {
			ticket.ContactID = autotask.Set(in.ContactID)
		}

		updated, err := autotask.Update[entities.Ticket](ctx, client, ticket)
		if err != nil {
			return errorResult("failed to update ticket %d: %v", in.TicketID, err)
		}

		m, err := entityToMap(updated)
		if err != nil {
			return errorResult("failed to convert updated ticket: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal updated ticket: %v", err)
		}

		return textResult("%s", string(data))
	}
}
