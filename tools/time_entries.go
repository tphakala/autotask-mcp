package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// CreateTimeEntryInput defines the input parameters for creating a new time entry.
type CreateTimeEntryInput struct {
	ResourceID    int64   `json:"resourceID" jsonschema:"Resource ID logging the time"`
	BillingCodeID int64   `json:"billingCodeID,omitempty" jsonschema:"Billing code ID for the time entry"`
	DateWorked    string  `json:"dateWorked" jsonschema:"Date worked (YYYY-MM-DD or ISO format)"`
	HoursWorked   float64 `json:"hoursWorked" jsonschema:"Number of hours worked"`
	SummaryNotes  string  `json:"summaryNotes" jsonschema:"Summary notes"`
	TicketID      int64   `json:"ticketID,omitempty" jsonschema:"Ticket ID for the time entry"`
	StartDateTime string  `json:"startDateTime,omitempty" jsonschema:"Start date/time (ISO format)"`
	EndDateTime   string  `json:"endDateTime,omitempty" jsonschema:"End date/time (ISO format)"`
	InternalNotes string  `json:"internalNotes,omitempty" jsonschema:"Internal notes"`
}

// SearchTimeEntriesInput defines the input parameters for searching time entries.
type SearchTimeEntriesInput struct {
	ResourceID       int64  `json:"resourceID,omitempty" jsonschema:"Filter by resource ID"`
	TicketID         int64  `json:"ticketID,omitempty" jsonschema:"Filter by ticket ID"`
	DateWorkedAfter  string `json:"dateWorkedAfter,omitempty" jsonschema:"Filter entries worked on or after this date (ISO format)"`
	DateWorkedBefore string `json:"dateWorkedBefore,omitempty" jsonschema:"Filter entries worked on or before this date (ISO format)"`
	Page             int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize         int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterTimeEntryTools registers all time entry-related MCP tools with the server.
func RegisterTimeEntryTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_time_entry",
		Description: "Create a new time entry in Autotask.",
	}, createTimeEntryHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_time_entries",
		Description: "Search for time entries in Autotask. Returns 25 results per page by default.",
	}, searchTimeEntriesHandler(client, mapper))
}

// createTimeEntryHandler returns a handler that creates a new time entry.
func createTimeEntryHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTimeEntryInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTimeEntryInput) (*mcp.CallToolResult, any, error) {
		dateWorked, err := parseDate(in.DateWorked)
		if err != nil {
			return errorResult("invalid dateWorked format (expected YYYY-MM-DD or ISO format): %v", err)
		}

		entry := &entities.TimeEntry{
			ResourceID:   autotask.Set(in.ResourceID),
			DateWorked:   autotask.Set(dateWorked),
			HoursWorked:  autotask.Set(in.HoursWorked),
			SummaryNotes: autotask.Set(in.SummaryNotes),
		}

		if in.TicketID != 0 {
			entry.TicketID = autotask.Set(in.TicketID)
		}
		if in.BillingCodeID != 0 {
			entry.BillingCodeID = autotask.Set(in.BillingCodeID)
		}
		if in.InternalNotes != "" {
			entry.InternalNotes = autotask.Set(in.InternalNotes)
		}
		if in.StartDateTime != "" {
			t, err := parseDate(in.StartDateTime)
			if err != nil {
				return errorResult("invalid startDateTime format (expected ISO 8601 / RFC3339): %v", err)
			}
			entry.StartDateTime = autotask.Set(t)
		}
		if in.EndDateTime != "" {
			t, err := parseDate(in.EndDateTime)
			if err != nil {
				return errorResult("invalid endDateTime format (expected ISO 8601 / RFC3339): %v", err)
			}
			entry.EndDateTime = autotask.Set(t)
		}

		created, err := autotask.Create[entities.TimeEntry](ctx, client, entry)
		if err != nil {
			return errorResult("failed to create time entry: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created time entry: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created time entry: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchTimeEntriesHandler returns a handler that searches time entries using the provided filters.
func searchTimeEntriesHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTimeEntriesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTimeEntriesInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 500)

		q := autotask.NewQuery().Limit(pageSize)

		if in.ResourceID != 0 {
			q.Where("resourceID", autotask.OpEq, in.ResourceID)
		}
		if in.TicketID != 0 {
			q.Where("ticketID", autotask.OpEq, in.TicketID)
		}
		if in.DateWorkedAfter != "" {
			q.Where("dateWorked", autotask.OpGte, in.DateWorkedAfter)
		}
		if in.DateWorkedBefore != "" {
			q.Where("dateWorked", autotask.OpLte, in.DateWorkedBefore)
		}

		entries, err := autotask.List[entities.TimeEntry](ctx, client, q)
		if err != nil {
			return errorResult("failed to search time entries: %v", err)
		}

		if len(entries) == 0 {
			return textResult("No time entries found")
		}

		maps, err := entitiesToMaps(entries)
		if err != nil {
			return errorResult("failed to convert time entries: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_time_entries", page, pageSize)
	}
}
