package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetTicketNoteInput defines the input parameters for getting a single ticket note.
type GetTicketNoteInput struct {
	NoteID int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchTicketNotesInput defines the input parameters for searching ticket notes.
type SearchTicketNotesInput struct {
	TicketID   int64 `json:"ticketId" jsonschema:"Ticket ID to list notes for"`
	MaxResults int   `json:"maxResults,omitempty" jsonschema:"Max notes to return (default 25, max 100)"`
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
	NoteID int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchProjectNotesInput defines the input parameters for searching project notes.
type SearchProjectNotesInput struct {
	ProjectID  int64 `json:"projectId" jsonschema:"Project ID to list notes for"`
	MaxResults int   `json:"maxResults,omitempty" jsonschema:"Max notes to return (default 25, max 100)"`
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
	NoteID int64 `json:"noteId" jsonschema:"Note ID to retrieve"`
}

// SearchCompanyNotesInput defines the input parameters for searching company notes.
type SearchCompanyNotesInput struct {
	CompanyID  int64 `json:"companyId" jsonschema:"Company ID to list notes for"`
	MaxResults int   `json:"maxResults,omitempty" jsonschema:"Max notes to return (default 25, max 100)"`
}

// CreateCompanyNoteInput defines the input parameters for creating a company note.
type CreateCompanyNoteInput struct {
	CompanyID   int64  `json:"companyId" jsonschema:"Company ID to add the note to"`
	Description string `json:"description" jsonschema:"Note body text"`
	Title       string `json:"title,omitempty" jsonschema:"Note title"`
	ActionType  int    `json:"actionType,omitempty" jsonschema:"Action type ID"`
}

const maxNoteSummaryLength = 500

func truncateNoteBody(m map[string]any) {
	for _, field := range []string{"description", "note"} {
		val, ok := m[field].(string)
		if !ok || len(val) <= maxNoteSummaryLength {
			continue
		}
		count := 0
		byteIdx := len(val)
		for i := range val {
			if count == maxNoteSummaryLength {
				byteIdx = i
				break
			}
			count++
		}
		if byteIdx < len(val) {
			m[field] = val[:byteIdx] + "\n... [truncated for search, use get note tool for full text]"
		}
	}
}

// RegisterNoteTools registers all note-related MCP tools with the server.
func RegisterNoteTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_note",
		Description: "Retrieve one ticket note by its numeric note ID, returning the complete note title, body, type, and publish scope recorded against a service ticket. Read-only.",
		Annotations: readOnlyTool("Get ticket note"),
	}, getTicketNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_ticket_notes",
		Description: "List note headers and truncated bodies (first 500 characters) for a ticket, returning a compact summary capped at maxResults (default 25, max 100). Use this to scan a ticket history; use autotask_get_ticket_note for full text. Requires ticketId. Read-only.",
		Annotations: readOnlyTool("Search ticket notes"),
	}, searchTicketNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_ticket_note",
		Description: "Add a note to an existing ticket from a title and description, with optional noteType and publish scope. Requires ticketId and description. Writes to Autotask.",
		Annotations: createTool("Create ticket note"),
	}, createTicketNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_project_note",
		Description: "Retrieve one project note by its numeric note ID, returning the complete note title, description, and type recorded against a project. Read-only.",
		Annotations: readOnlyTool("Get project note"),
	}, getProjectNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_project_notes",
		Description: "List note headers and truncated descriptions (first 500 characters) for a project, returning a compact summary capped at maxResults (default 25, max 100). Use autotask_get_project_note for full text. Requires projectId. Read-only.",
		Annotations: readOnlyTool("Search project notes"),
	}, searchProjectNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_project_note",
		Description: "Add a note to an existing project from a description and optional title and noteType. Requires projectId and description. Writes to Autotask.",
		Annotations: createTool("Create project note"),
	}, createProjectNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_company_note",
		Description: "Retrieve one company note by its numeric note ID, returning the complete note name, body, and action type recorded against a company account. Read-only.",
		Annotations: readOnlyTool("Get company note"),
	}, getCompanyNoteHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_company_notes",
		Description: "List note headers and truncated bodies (first 500 characters) for a company, returning a compact summary capped at maxResults (default 25, max 100). Use autotask_get_company_note for full text. Requires companyId. Read-only.",
		Annotations: readOnlyTool("Search company notes"),
	}, searchCompanyNotesHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_company_note",
		Description: "Add a note to an existing company account from a description (body) and optional title (name) and actionType. Requires companyId and description. Writes to Autotask.",
		Annotations: createTool("Create company note"),
	}, createCompanyNoteHandler(client))
}

// getTicketNoteHandler returns a handler that retrieves a single ticket note by ID.
func getTicketNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		note, err := autotask.Get[entities.TicketNote](ctx, client, in.NoteID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(note)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchTicketNotesHandler returns a handler that lists notes for a ticket.
func searchTicketNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		notes, err := autotask.ListChild[entities.Ticket, entities.TicketNote](ctx, client, in.TicketID)
		if err != nil {
			var notFound *autotask.NotFoundError
			if errors.As(err, &notFound) {
				return emptySearchResult()
			}
			return nil, services.CompactResponse{}, err
		}

		if len(notes) == 0 {
			return emptySearchResult()
		}

		maxResults := defaultMaxResults(in.MaxResults, 25, 100)
		hasMore := len(notes) >= maxResults && maxResults > 0
		if len(notes) > maxResults {
			notes = notes[:maxResults]
		}

		maps, err := entitiesToMaps(notes)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		for _, m := range maps {
			truncateNoteBody(m)
			services.FrameUntrustedMapFields(m)
		}

		hint := ""
		if hasMore {
			hint = "Maximum result limit reached. Use narrower search filters to find specific records."
		}

		return nil, services.CompactResponse{
			Summary: services.CompactSummary{
				Returned:   len(maps),
				HasMore:    hasMore,
				MaxResults: maxResults,
				Hint:       hint,
			},
			Items: maps,
		}, nil
	}
}

// createTicketNoteHandler returns a handler that creates a new note on a ticket.
func createTicketNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTicketNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		if in.Description == "" {
			return nil, nil, fmt.Errorf("description is required")
		}
		note := &entities.TicketNote{
			Description: autotask.Set(in.Description),
		}

		if in.Title != "" {
			note.Title = autotask.Set(in.Title)
		}
		if in.NoteType != 0 {
			note.NoteType = autotask.Set(int64(in.NoteType))
		}
		if in.Publish != 0 {
			note.Publish = autotask.Set(int64(in.Publish))
		}

		created, err := autotask.CreateChild[entities.Ticket, entities.TicketNote](ctx, client, in.TicketID, note)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(created)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// getProjectNoteHandler returns a handler that retrieves a single project note by ID.
func getProjectNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetProjectNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetProjectNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		note, err := autotask.Get[entities.ProjectNote](ctx, client, in.NoteID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(note)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchProjectNotesHandler returns a handler that lists notes for a project.
func searchProjectNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		notes, err := autotask.ListChildRaw(ctx, client, "Projects", in.ProjectID, "ProjectNotes")
		if err != nil {
			var notFound *autotask.NotFoundError
			if errors.As(err, &notFound) {
				return emptySearchResult()
			}
			return nil, services.CompactResponse{}, err
		}

		if len(notes) == 0 {
			return emptySearchResult()
		}

		maxResults := defaultMaxResults(in.MaxResults, 25, 100)
		hasMore := len(notes) >= maxResults && maxResults > 0
		if len(notes) > maxResults {
			notes = notes[:maxResults]
		}

		for _, m := range notes {
			truncateNoteBody(m)
			services.FrameUntrustedMapFields(m)
		}

		hint := ""
		if hasMore {
			hint = "Maximum result limit reached. Use narrower search filters to find specific records."
		}

		return nil, services.CompactResponse{
			Summary: services.CompactSummary{
				Returned:   len(notes),
				HasMore:    hasMore,
				MaxResults: maxResults,
				Hint:       hint,
			},
			Items: notes,
		}, nil
	}
}

// createProjectNoteHandler returns a handler that creates a new note on a project.
func createProjectNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		if in.Description == "" {
			return nil, nil, fmt.Errorf("description is required")
		}
		note := &entities.ProjectNote{
			Description: autotask.Set(in.Description),
		}
		if in.Title != "" {
			note.Title = autotask.Set(in.Title)
		}
		if in.NoteType != 0 {
			note.NoteType = autotask.Set(int64(in.NoteType))
		}

		created, err := autotask.CreateChild[entities.Project, entities.ProjectNote](ctx, client, in.ProjectID, note)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(created)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// getCompanyNoteHandler returns a handler that retrieves a single company note by ID.
func getCompanyNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetCompanyNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetCompanyNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		note, err := autotask.Get[entities.CompanyNote](ctx, client, in.NoteID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(note)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}

// searchCompanyNotesHandler returns a handler that lists notes for a company.
func searchCompanyNotesHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompanyNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompanyNotesInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		notes, err := autotask.ListChildRaw(ctx, client, "Companies", in.CompanyID, "CompanyNotes")
		if err != nil {
			var notFound *autotask.NotFoundError
			if errors.As(err, &notFound) {
				return emptySearchResult()
			}
			return nil, services.CompactResponse{}, err
		}

		if len(notes) == 0 {
			return emptySearchResult()
		}

		maxResults := defaultMaxResults(in.MaxResults, 25, 100)
		hasMore := len(notes) >= maxResults && maxResults > 0
		if len(notes) > maxResults {
			notes = notes[:maxResults]
		}

		for _, m := range notes {
			truncateNoteBody(m)
			services.FrameUntrustedMapFields(m)
		}

		hint := ""
		if hasMore {
			hint = "Maximum result limit reached. Use narrower search filters to find specific records."
		}

		return nil, services.CompactResponse{
			Summary: services.CompactSummary{
				Returned:   len(notes),
				HasMore:    hasMore,
				MaxResults: maxResults,
				Hint:       hint,
			},
			Items: notes,
		}, nil
	}
}

// createCompanyNoteHandler returns a handler that creates a new note on a company.
func createCompanyNoteHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyNoteInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		if in.Description == "" {
			return nil, nil, fmt.Errorf("description is required")
		}
		// CompanyNote uses different field names than TicketNote/ProjectNote:
		// Input.Description -> entity.Note (body text)
		// Input.Title -> entity.Name (display name)
		note := &entities.CompanyNote{
			Note: autotask.Set(in.Description),
		}
		if in.Title != "" {
			note.Name = autotask.Set(in.Title)
		}
		if in.ActionType != 0 {
			note.ActionType = autotask.Set(int64(in.ActionType))
		}

		created, err := autotask.CreateChild[entities.Company, entities.CompanyNote](ctx, client, in.CompanyID, note)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(created)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}
