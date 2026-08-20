package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/entities"
)

// TestRegisterNoteTools_NoPanic verifies that RegisterNoteTools registers all tools without panicking.
func TestRegisterNoteTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterNoteTools(s, client)
}

// TestSearchTicketNotesHandler_NoNotes tests searching ticket notes on an empty server.
func TestSearchTicketNotesHandler_NoNotes(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_notes",
		Arguments: map[string]any{
			"ticketId": 3001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if result.IsError {
		if len(result.Content) > 0 {
			if tc, ok := result.Content[0].(*mcp.TextContent); ok {
				t.Fatalf("expected non-error result, got IsError=true; content: %s", tc.Text)
			}
		}
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned notes, got %d", resp.Summary.Returned)
	}
}

// TestSearchTicketNotesHandler_WithNotes tests that seeded notes are returned over wire.
func TestSearchTicketNotesHandler_WithNotes(t *testing.T) {
	note := autotasktest.TicketNoteFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(note))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_notes",
		Arguments: map[string]any{
			"ticketId": 3001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 note, got %d", resp.Summary.Returned)
	}
}

// TestSearchTicketNotesHandler_TruncationAndFraming tests that long note bodies are truncated and untrusted content is framed over wire.
func TestSearchTicketNotesHandler_TruncationAndFraming(t *testing.T) {
	longDescription := strings.Repeat("A", 600)
	note := autotasktest.TicketNoteFixture(func(n *entities.TicketNote) {
		n.Description = autotask.Set(longDescription)
		n.Title = autotask.Set("Untrusted Note Title")
	})
	cs, _ := setupWireTest(t, autotasktest.WithEntity(note))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_notes",
		Arguments: map[string]any{
			"ticketId":   3001,
			"maxResults": 10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if len(resp.Items) == 0 {
		t.Fatal("expected at least 1 note")
	}

	n := resp.Items[0]
	desc, _ := n["description"].(string)
	if !strings.Contains(desc, "truncated for search") {
		t.Errorf("expected description to be truncated, got: %v", desc)
	}
	if !strings.Contains(desc, "<untrusted_content>") {
		t.Errorf("expected description to be framed with untrusted_content tags, got: %v", desc)
	}

	title, _ := n["title"].(string)
	if !strings.Contains(title, "<untrusted_content>") {
		t.Errorf("expected title to be framed with untrusted_content tags, got: %v", title)
	}
}

// TestGetTicketNoteHandler_NotFound tests that a missing note returns an error result over wire.
func TestGetTicketNoteHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_note",
		Arguments: map[string]any{
			"noteId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing note")
	}
}

// TestGetTicketNoteHandler_Success tests that a seeded note is retrieved over wire.
func TestGetTicketNoteHandler_Success(t *testing.T) {
	note := autotasktest.TicketNoteFixture()
	noteID, ok := note.ID.Get()
	if !ok {
		t.Fatal("fixture note has no ID")
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(note))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_note",
		Arguments: map[string]any{
			"noteId": noteID,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	idVal, ok := m["id"]
	if !ok {
		t.Fatalf("expected 'id' field in note response")
	}
	fVal, isFloat := idVal.(float64)
	if !isFloat || int64(fVal) != noteID {
		t.Errorf("expected id=%d, got %v", noteID, idVal)
	}
}

// TestCreateTicketNoteHandler_Success tests creating a ticket note over wire.
func TestCreateTicketNoteHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.TicketNoteFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_ticket_note",
		Arguments: map[string]any{
			"ticketId":    3001,
			"description": "Test note description",
			"title":       "Test note",
			"noteType":    1,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	titleStr, _ := m["title"].(string)
	if !strings.Contains(titleStr, "Test note") {
		t.Errorf("expected title to contain 'Test note', got %v", m["title"])
	}
}

// TestGetProjectNoteHandler_NotFound tests that a missing project note returns an error result over wire.
func TestGetProjectNoteHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_project_note",
		Arguments: map[string]any{
			"noteId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing project note")
	}
}

// TestSearchProjectNotesHandler_NoNotes tests searching project notes on an empty server.
func TestSearchProjectNotesHandler_NoNotes(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_project_notes",
		Arguments: map[string]any{
			"projectId": 4001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned project notes, got %d", resp.Summary.Returned)
	}
}

// TestSearchCompanyNotesHandler_NoNotes tests searching company notes on an empty server.
func TestSearchCompanyNotesHandler_NoNotes(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_company_notes",
		Arguments: map[string]any{
			"companyId": 1001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned company notes, got %d", resp.Summary.Returned)
	}
}

// TestCreateProjectNoteHandler_Success tests creating a project note over wire.
func TestCreateProjectNoteHandler_Success(t *testing.T) {
	proj := autotasktest.ProjectFixture()
	projID, ok := proj.ID.Get()
	if !ok {
		t.Fatal("fixture project has no ID")
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(proj))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_project_note",
		Arguments: map[string]any{
			"projectId":   projID,
			"description": "Project note body",
			"title":       "Project note title",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	titleStr, _ := m["title"].(string)
	if !strings.Contains(titleStr, "Project note title") {
		t.Errorf("expected title to contain 'Project note title', got %v", m["title"])
	}
}

// TestCreateCompanyNoteHandler_Success tests creating a company note over wire.
func TestCreateCompanyNoteHandler_Success(t *testing.T) {
	comp := autotasktest.CompanyFixture()
	compID, ok := comp.ID.Get()
	if !ok {
		t.Fatal("fixture company has no ID")
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(comp))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_company_note",
		Arguments: map[string]any{
			"companyId":   compID,
			"description": "Company note body",
			"title":       "Company note title",
			"actionType":  1,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	nameStr, _ := m["name"].(string)
	if !strings.Contains(nameStr, "Company note title") {
		t.Errorf("expected name to contain 'Company note title', got %v", m["name"])
	}
}


