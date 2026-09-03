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

	// Regression: ticket notes must be framed exactly ONCE, AFTER truncation.
	// The previous frame -> truncate -> frame sequence severed the closing marker,
	// left a stray escaped &lt;untrusted_content&gt; artifact, and truncated the body
	// ~20 chars early (counting the framing prefix as content).
	if !strings.HasPrefix(desc, "<untrusted_content>\n") || !strings.HasSuffix(desc, "\n</untrusted_content>") {
		t.Errorf("expected description framed exactly once with intact markers, got: %q", desc)
	}
	if strings.Contains(desc, "&lt;untrusted_content&gt;") || strings.Contains(desc, "&lt;/untrusted_content&gt;") {
		t.Errorf("description contains a stray escaped boundary artifact (double-framing): %q", desc)
	}
	if got := strings.Count(desc, "A"); got != maxNoteSummaryLength {
		t.Errorf("expected exactly %d body chars before truncation (raw truncation), got %d", maxNoteSummaryLength, got)
	}

	title, _ := n["title"].(string)
	if !strings.Contains(title, "<untrusted_content>") {
		t.Errorf("expected title to be framed with untrusted_content tags, got: %v", title)
	}
}

// TestSearchTicketNotesHandler_BoundedByMaxResults seeds more notes than the
// requested limit and asserts the handler returns exactly maxResults and reports
// hasMore. It pins the bounded-fetch contract: the caller sees only the capped set.
func TestSearchTicketNotesHandler_BoundedByMaxResults(t *testing.T) {
	notes := make([]entities.TicketNote, 0, 5)
	for i := 0; i < 5; i++ {
		notes = append(notes, autotasktest.TicketNoteFixture())
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(notes...))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_notes",
		Arguments: map[string]any{
			"ticketId":   3001,
			"maxResults": 3,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 3 {
		t.Errorf("expected 3 returned notes (bounded to maxResults), got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Items))
	}
	if !resp.Summary.HasMore {
		t.Error("expected hasMore=true when more notes exist beyond maxResults")
	}
	if resp.Summary.MaxResults != 3 {
		t.Errorf("expected maxResults=3 in summary, got %d", resp.Summary.MaxResults)
	}
}

// TestSearchTicketNotesHandler_ExactCountReportsNoMore pins the corrected hasMore
// semantics: with exactly maxResults notes and no more, hasMore is false. The prior
// implementation reported hasMore=true here (it tested len(all) >= maxResults on the
// fully buffered slice), a false positive the bounded peek-one-extra fetch removes.
func TestSearchTicketNotesHandler_ExactCountReportsNoMore(t *testing.T) {
	notes := make([]entities.TicketNote, 0, 3)
	for i := 0; i < 3; i++ {
		notes = append(notes, autotasktest.TicketNoteFixture())
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(notes...))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_notes",
		Arguments: map[string]any{
			"ticketId":   3001,
			"maxResults": 3,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 3 {
		t.Errorf("expected 3 returned notes, got %d", resp.Summary.Returned)
	}
	if resp.Summary.HasMore {
		t.Error("expected hasMore=false when exactly maxResults notes exist with none beyond")
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

// TestSearchProjectNotesHandler_BoundedAndExactCount pins the bounded-fetch and
// corrected hasMore semantics for the project-note handler. It shares the
// collectBoundedChildRaw refactor exercised by the ticket-note tests; each handler
// that uses it needs its own coverage so a per-handler regression cannot hide.
func TestSearchProjectNotesHandler_BoundedAndExactCount(t *testing.T) {
	seed := func(n int) []entities.ProjectNote {
		notes := make([]entities.ProjectNote, 0, n)
		for i := 0; i < n; i++ {
			notes = append(notes, entities.ProjectNote{ID: autotask.Set(int64(4000 + i))})
		}
		return notes
	}

	// More notes than the limit: bounded to maxResults, hasMore true.
	cs, _ := setupWireTest(t, autotasktest.WithEntity(seed(5)...))
	resp := callSearchNotes(t, cs, "autotask_search_project_notes", "projectId", 3)
	if resp.Summary.Returned != 3 || !resp.Summary.HasMore {
		t.Errorf("bounded: got Returned=%d HasMore=%v, want 3/true", resp.Summary.Returned, resp.Summary.HasMore)
	}

	// Exactly maxResults notes: hasMore must be false (no false positive).
	cs2, _ := setupWireTest(t, autotasktest.WithEntity(seed(3)...))
	resp2 := callSearchNotes(t, cs2, "autotask_search_project_notes", "projectId", 3)
	if resp2.Summary.Returned != 3 || resp2.Summary.HasMore {
		t.Errorf("exact count: got Returned=%d HasMore=%v, want 3/false", resp2.Summary.Returned, resp2.Summary.HasMore)
	}
}

// TestSearchCompanyNotesHandler_BoundedAndExactCount pins the same for the company
// note handler (the third site of the shared refactor).
func TestSearchCompanyNotesHandler_BoundedAndExactCount(t *testing.T) {
	seed := func(n int) []entities.CompanyNote {
		notes := make([]entities.CompanyNote, 0, n)
		for i := 0; i < n; i++ {
			notes = append(notes, entities.CompanyNote{ID: autotask.Set(int64(5000 + i))})
		}
		return notes
	}

	cs, _ := setupWireTest(t, autotasktest.WithEntity(seed(5)...))
	resp := callSearchNotes(t, cs, "autotask_search_company_notes", "companyId", 3)
	if resp.Summary.Returned != 3 || !resp.Summary.HasMore {
		t.Errorf("bounded: got Returned=%d HasMore=%v, want 3/true", resp.Summary.Returned, resp.Summary.HasMore)
	}

	cs2, _ := setupWireTest(t, autotasktest.WithEntity(seed(3)...))
	resp2 := callSearchNotes(t, cs2, "autotask_search_company_notes", "companyId", 3)
	if resp2.Summary.Returned != 3 || resp2.Summary.HasMore {
		t.Errorf("exact count: got Returned=%d HasMore=%v, want 3/false", resp2.Summary.Returned, resp2.Summary.HasMore)
	}
}

// callSearchNotes invokes a child-note search tool and returns its compact response.
func callSearchNotes(t *testing.T, cs *mcp.ClientSession, tool, idField string, maxResults int) services.CompactResponse {
	t.Helper()
	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: tool,
		Arguments: map[string]any{
			idField:      3001,
			"maxResults": maxResults,
		},
	})
	if err != nil {
		t.Fatalf("%s: unexpected wire error: %v", tool, err)
	}
	if result.IsError {
		t.Fatalf("%s: expected non-error result, got: %v", tool, result)
	}
	return parseStructuredContent[services.CompactResponse](t, result)
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
