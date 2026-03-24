package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterNoteTools_NoPanic verifies that RegisterNoteTools registers all tools without panicking.
func TestRegisterNoteTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterNoteTools(s, client)
}

// TestSearchTicketNotesHandler_NoNotes tests that the handler returns a result (possibly empty
// or error) without panicking when called on an empty server.
func TestSearchTicketNotesHandler_NoNotes(t *testing.T) {
	// The mock server returns 404 when the TicketNotes entity store does not exist,
	// which propagates as an error result (not a protocol error). Verify no panic.
	_, client := autotasktest.NewServer(t)
	handler := searchTicketNotesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTicketNotesInput{TicketID: 3001})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The result may be an error (404 from mock) or success — both are acceptable.
	// We only verify there is no panic and a result is returned.
}

// TestSearchTicketNotesHandler_WithNotes tests that seeded notes are returned.
func TestSearchTicketNotesHandler_WithNotes(t *testing.T) {
	note := autotasktest.TicketNoteFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(note))

	handler := searchTicketNotesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTicketNotesInput{TicketID: 3001})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestGetTicketNoteHandler_NotFound tests that a missing note returns an error result.
func TestGetTicketNoteHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getTicketNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketNoteInput{TicketID: 3001, NoteID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing note")
	}
}

// TestGetTicketNoteHandler_Success tests that a seeded note is retrieved.
func TestGetTicketNoteHandler_Success(t *testing.T) {
	note := autotasktest.TicketNoteFixture()
	noteID, ok := note.ID.Get()
	if !ok {
		t.Fatal("fixture note has no ID")
	}
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(note))

	handler := getTicketNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketNoteInput{TicketID: 3001, NoteID: noteID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestCreateTicketNoteHandler_Success tests that a ticket note can be created.
func TestCreateTicketNoteHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.TicketNoteFixture()),
	)

	handler := createTicketNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateTicketNoteInput{
		TicketID:    3001,
		Description: "Test note description",
		Title:       "Test note",
		NoteType:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestGetProjectNoteHandler_NotFound tests that a missing project note returns an error result.
func TestGetProjectNoteHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getProjectNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetProjectNoteInput{ProjectID: 4001, NoteID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing project note")
	}
}

// TestSearchProjectNotesHandler_NoNotes tests that the handler does not panic on an empty server.
func TestSearchProjectNotesHandler_NoNotes(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchProjectNotesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchProjectNotesInput{ProjectID: 4001})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestSearchCompanyNotesHandler_NoNotes tests that the handler does not panic on an empty server.
func TestSearchCompanyNotesHandler_NoNotes(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchCompanyNotesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchCompanyNotesInput{CompanyID: 1001})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestCreateProjectNoteHandler_Success tests that a project note can be created.
func TestCreateProjectNoteHandler_Success(t *testing.T) {
	proj := autotasktest.ProjectFixture()
	projID, ok := proj.ID.Get()
	if !ok {
		t.Fatal("fixture project has no ID")
	}
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(proj))

	handler := createProjectNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateProjectNoteInput{
		ProjectID:   projID,
		Description: "Project note body",
		Title:       "Project note title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestCreateCompanyNoteHandler_Success tests that a company note can be created.
func TestCreateCompanyNoteHandler_Success(t *testing.T) {
	comp := autotasktest.CompanyFixture()
	compID, ok := comp.ID.Get()
	if !ok {
		t.Fatal("fixture company has no ID")
	}
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(comp))

	handler := createCompanyNoteHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, CreateCompanyNoteInput{
		CompanyID:   compID,
		Description: "Company note body",
		Title:       "Company note title",
		ActionType:  1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

