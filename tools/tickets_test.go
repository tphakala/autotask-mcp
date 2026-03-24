package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/autotask-mcp/services"
)

// newTestMapper returns a MappingCache backed by the provided client.
func newTestMapper(client *autotask.Client) *services.MappingCache {
	return services.NewMappingCache(client)
}

// TestRegisterTicketTools_NoPanic verifies that RegisterTicketTools registers all
// four tools without panicking.
func TestRegisterTicketTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterTicketTools(s, client, mapper)
}

// TestSearchTicketsHandler_ReturnsNoTicketsFound tests the empty-result case.
func TestSearchTicketsHandler_ReturnsNoTicketsFound(t *testing.T) {
	// Server with no pre-seeded tickets — all queries return 0 items.
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchTicketsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTicketsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if text.Text != "No tickets found" {
		t.Errorf("expected 'No tickets found', got %q", text.Text)
	}
}

// TestSearchTicketsHandler_ReturnsTickets tests that seeded tickets are returned.
func TestSearchTicketsHandler_ReturnsTickets(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))
	mapper := newTestMapper(client)

	handler := searchTicketsHandler(client, mapper)
	ctx := context.Background()

	// Search with a specific status to avoid the default "exclude status 5" filter
	// causing ambiguity with the fixture's status value.
	result, _, err := handler(ctx, nil, SearchTicketsInput{Status: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	// The result should be a JSON compact response.
	var resp map[string]any
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatalf("expected 'items' array in response, got: %v", resp)
	}
	if len(items) == 0 {
		t.Error("expected at least one ticket in results")
	}
}

// TestSearchTicketsHandler_DefaultExcludesCompleted verifies that the default query
// path (no status filter provided) does not return a protocol error.
func TestSearchTicketsHandler_DefaultExcludesCompleted(t *testing.T) {
	// Seed an open ticket (status 1). The handler adds a "status != 5" filter by default.
	openTicket := autotasktest.TicketFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(openTicket))
	mapper := newTestMapper(client)

	handler := searchTicketsHandler(client, mapper)
	ctx := context.Background()

	// No Status provided → handler applies OpNotEq 5 filter.
	result, _, err := handler(ctx, nil, SearchTicketsInput{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The fixture ticket has status=1, so it should pass the != 5 filter.
	// However, the mock server does not implement filter evaluation server-side,
	// so the ticket will still be returned. We just verify no protocol error occurred.
}

// TestGetTicketDetailsHandler_Success tests that a ticket is retrieved and returned.
func TestGetTicketDetailsHandler_Success(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	ticketID, ok := ticket.ID.Get()
	if !ok {
		t.Fatal("fixture ticket has no ID")
	}

	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))
	mapper := newTestMapper(client)

	handler := getTicketDetailsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketDetailsInput{TicketID: ticketID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(text.Text), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}

	// Verify the ID matches.
	idVal, ok := m["id"]
	if !ok {
		t.Error("expected 'id' field in result")
	} else if int64(idVal.(float64)) != ticketID {
		t.Errorf("expected id=%d, got %v", ticketID, idVal)
	}
}

// TestGetTicketDetailsHandler_NotFound tests that a missing ticket returns an error result.
func TestGetTicketDetailsHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := getTicketDetailsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketDetailsInput{TicketID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing ticket")
	}
}

// TestCreateTicketHandler_Success tests that a ticket can be created.
func TestCreateTicketHandler_Success(t *testing.T) {
	// Seed the entity store so the server accepts Tickets POSTs.
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.TicketFixture()), // initialises the store
	)

	handler := createTicketHandler(client)
	ctx := context.Background()

	in := CreateTicketInput{
		CompanyID:   1001,
		Title:       "Test ticket",
		Description: "This is a test ticket created by the handler.",
		Priority:    2,
		Status:      1,
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(text.Text), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}
}

// TestUpdateTicketHandler_Success tests that an existing ticket can be updated.
func TestUpdateTicketHandler_Success(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	ticketID, ok := ticket.ID.Get()
	if !ok {
		t.Fatal("fixture ticket has no ID")
	}

	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))

	handler := updateTicketHandler(client)
	ctx := context.Background()

	in := UpdateTicketInput{
		TicketID: ticketID,
		Title:    "Updated title",
		Status:   2,
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(text.Text), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}
}

// TestUpdateTicketHandler_InvalidDueDateTime tests that an invalid date returns an error result.
func TestUpdateTicketHandler_InvalidDueDateTime(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	ticketID, ok := ticket.ID.Get()
	if !ok {
		t.Fatal("fixture ticket has no ID")
	}

	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))

	handler := updateTicketHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, UpdateTicketInput{
		TicketID:    ticketID,
		DueDateTime: "not-a-date",
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid date")
	}
}

// TestSearchTicketsHandler_WithFilters verifies that multiple filters can be applied
// without panicking and that the handler returns a result.
func TestSearchTicketsHandler_WithFilters(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))
	mapper := newTestMapper(client)

	handler := searchTicketsHandler(client, mapper)
	ctx := context.Background()

	// Apply several filters; the exact result count depends on the mock server's
	// filtering, but the handler must not return a protocol error.
	in := SearchTicketsInput{
		CompanyID:    1001,
		Status:       1,
		CreatedAfter: "2024-01-01T00:00:00Z",
		Page:         1,
		PageSize:     10,
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should not be a protocol-level error.
	_ = result
}

// TestSearchTicketsHandler_UnassignedFilter verifies that the Unassigned flag is accepted.
func TestSearchTicketsHandler_UnassignedFilter(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchTicketsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTicketsInput{Unassigned: true})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
