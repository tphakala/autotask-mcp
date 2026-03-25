package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetTicketNoteInput defines the input parameters for getting a single ticket note.
type GetTicketNoteInput struct {
	TicketID int64 `json:"ticketId" jsonschema:"Ticket ID that owns the note"`
	NoteID   int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchTicketNotesInput defines the input parameters for searching ticket notes.
type SearchTicketNotesInput struct {
	TicketID int64 `json:"ticketId" jsonschema:"Ticket ID to list notes for"`
}

// CreateTicketNoteInput defines the input parameters for creating a ticket note.
type CreateTicketNoteInput struct {
	TicketID    int64  `json:"ticketId" jsonschema:"Ticket ID to add the note to"`
	Description string `json:"description" jsonschema:"Note body text"`
	Title       string `json:"title,omitempty" jsonschema:"Note title"`
	NoteType    int    `json:"noteType,omitempty" jsonschema:"Note type ID"`
	Publish     int    `json:"publish,omitempty" jsonschema:"Publish target ID"`
}

// GetProjectNoteInput defines the input parameters for getting a single project note.
type GetProjectNoteInput struct {
	ProjectID int64 `json:"projectId" jsonschema:"Project ID that owns the note"`
	NoteID    int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchProjectNotesInput defines the input parameters for searching project notes.
type SearchProjectNotesInput struct {
	ProjectID int64 `json:"projectId" jsonschema:"Project ID to list notes for"`
}

// CreateProjectNoteInput defines the input parameters for creating a project note.
type CreateProjectNoteInput struct {
	ProjectID   int64  `json:"projectId" jsonschema:"Project ID to add the note to"`
	Description string `json:"description" jsonschema:"Note body text"`
	Title       string `json:"title,omitempty" jsonschema:"Note title"`
	NoteType    int    `json:"noteType,omitempty" jsonschema:"Note type ID"`
}

// GetCompanyNoteInput defines the input parameters for getting a single company note.
type GetCompanyNoteInput struct {
	CompanyID int64 `json:"companyId" jsonschema:"Company ID that owns the note"`
	NoteID    int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchCompanyNotesInput defines the input parameters for searching company notes.
type SearchCompanyNotesInput struct {
	CompanyID int64 `json:"companyId" jsonschema:"Company ID to list notes for"`
}

// CreateCompanyNoteInput defines the input parameters for creating a company note.
type CreateCompanyNoteInput struct {
	CompanyID   int64  `json:"companyId" jsonschema:"Company ID to add the note to"`
	Description string `json:"description" jsonschema:"Note body text"`
	Title       string `json:"title,omitempty" jsonschema:"Note title"`
	ActionType  int    `json:"actionType,omitempty" jsonschema:"Action type ID"`
}

// RegisterNoteTools registers all note-related MCP tools with the server.
func RegisterNoteTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_note",
		Description: "Get a specific note for a ticket by note ID.",
	}, getTicketNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_ticket_notes",
		Description: "List all notes for a ticket.",
	}, searchTicketNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_ticket_note",
		Description: "Create a new note on a ticket.",
	}, createTicketNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_project_note",
		Description: "Get a specific note for a project by note ID.",
	}, getProjectNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_project_notes",
		Description: "List all notes for a project.",
	}, searchProjectNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_project_note",
		Description: "Create a new note on a project.",
	}, createProjectNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_company_note",
		Description: "Get a specific note for a company by note ID.",
	}, getCompanyNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_company_notes",
		Description: "List all notes for a company.",
	}, searchCompanyNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_company_note",
		Description: "Create a new note on a company.",
	}, createCompanyNoteHandler(client))
}

// getTicketNoteHandler returns a handler that retrieves a single ticket note by ID.
func getTicketNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketNoteInput) (*mcp.CallToolResult, any, error) {
		note, err := autotask.GetRaw(ctx, client, "TicketNotes", in.NoteID)
		if err != nil {
			return errorResult("failed to get ticket note %d: %v", in.NoteID, err)
		}

		data, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket note: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchTicketNotesHandler returns a handler that lists all notes for a ticket.
func searchTicketNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketNotesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketNotesInput) (*mcp.CallToolResult, any, error) {
		notes, err := autotask.ListChild[entities.Ticket, entities.TicketNote](ctx, client, in.TicketID)
		if err != nil {
			return errorResult("failed to list ticket notes for ticket %d: %v", in.TicketID, err)
		}

		if len(notes) == 0 {
			return textResult("No ticket notes found")
		}

		maps, err := entitiesToMaps(notes)
		if err != nil {
			return errorResult("failed to convert ticket notes: %v", err)
		}

		data, err := json.MarshalIndent(maps, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket notes: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createTicketNoteHandler returns a handler that creates a new note on a ticket.
func createTicketNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketNoteInput) (*mcp.CallToolResult, any, error) {
		note := &entities.TicketNote{
			Description: autotask.Set(in.Description),
		}

		if in.Title != "" {
			note.Title = autotask.Set(in.Title)
		}
		if in.NoteType != 0 {
			note.NoteType = autotask.Set(in.NoteType)
		}
		if in.Publish != 0 {
			note.Publish = autotask.Set(in.Publish)
		}

		created, err := autotask.CreateChild[entities.Ticket, entities.TicketNote](ctx, client, in.TicketID, note)
		if err != nil {
			return errorResult("failed to create ticket note: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created ticket note: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created ticket note: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// getProjectNoteHandler returns a handler that retrieves a single project note by ID.
func getProjectNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetProjectNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetProjectNoteInput) (*mcp.CallToolResult, any, error) {
		note, err := autotask.GetRaw(ctx, client, "ProjectNotes", in.NoteID)
		if err != nil {
			return errorResult("failed to get project note %d: %v", in.NoteID, err)
		}

		data, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return errorResult("failed to marshal project note: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchProjectNotesHandler returns a handler that lists all notes for a project.
func searchProjectNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectNotesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectNotesInput) (*mcp.CallToolResult, any, error) {
		notes, err := autotask.ListChildRaw(ctx, client, "Projects", in.ProjectID, "ProjectNotes")
		if err != nil {
			return errorResult("failed to list project notes for project %d: %v", in.ProjectID, err)
		}

		if len(notes) == 0 {
			return textResult("No project notes found")
		}

		data, err := json.MarshalIndent(notes, "", "  ")
		if err != nil {
			return errorResult("failed to marshal project notes: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createProjectNoteHandler returns a handler that creates a new note on a project.
func createProjectNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectNoteInput) (*mcp.CallToolResult, any, error) {
		data := map[string]any{
			"description": in.Description,
		}
		if in.Title != "" {
			data["title"] = in.Title
		}
		if in.NoteType != 0 {
			data["noteType"] = in.NoteType
		}

		created, err := autotask.CreateChildRaw(ctx, client, "Projects", in.ProjectID, "ProjectNotes", data)
		if err != nil {
			return errorResult("failed to create project note: %v", err)
		}

		out, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created project note: %v", err)
		}

		return textResult("%s", string(out))
	}
}

// getCompanyNoteHandler returns a handler that retrieves a single company note by ID.
func getCompanyNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetCompanyNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetCompanyNoteInput) (*mcp.CallToolResult, any, error) {
		note, err := autotask.GetRaw(ctx, client, "CompanyNotes", in.NoteID)
		if err != nil {
			return errorResult("failed to get company note %d: %v", in.NoteID, err)
		}

		data, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return errorResult("failed to marshal company note: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchCompanyNotesHandler returns a handler that lists all notes for a company.
func searchCompanyNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompanyNotesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompanyNotesInput) (*mcp.CallToolResult, any, error) {
		notes, err := autotask.ListChildRaw(ctx, client, "Companies", in.CompanyID, "CompanyNotes")
		if err != nil {
			return errorResult("failed to list company notes for company %d: %v", in.CompanyID, err)
		}

		if len(notes) == 0 {
			return textResult("No company notes found")
		}

		data, err := json.MarshalIndent(notes, "", "  ")
		if err != nil {
			return errorResult("failed to marshal company notes: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// createCompanyNoteHandler returns a handler that creates a new note on a company.
func createCompanyNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyNoteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyNoteInput) (*mcp.CallToolResult, any, error) {
		data := map[string]any{
			"description": in.Description,
		}
		if in.Title != "" {
			data["title"] = in.Title
		}
		if in.ActionType != 0 {
			data["actionType"] = in.ActionType
		}

		created, err := autotask.CreateChildRaw(ctx, client, "Companies", in.CompanyID, "CompanyNotes", data)
		if err != nil {
			return errorResult("failed to create company note: %v", err)
		}

		out, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created company note: %v", err)
		}

		return textResult("%s", string(out))
	}
}
