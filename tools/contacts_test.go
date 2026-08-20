package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterContactTools_NoPanic verifies that RegisterContactTools registers all
// two tools without panicking.
func TestRegisterContactTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	RegisterContactTools(s, client, mapper)
}

// TestSearchContactsHandler_ReturnsNoContactsFound tests the empty-result case over wire.
func TestSearchContactsHandler_ReturnsNoContactsFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_contacts",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned contacts, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

// TestSearchContactsHandler_ReturnsContacts tests that seeded contacts are returned over wire.
func TestSearchContactsHandler_ReturnsContacts(t *testing.T) {
	contact := autotasktest.ContactFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(contact))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_contacts",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 returned contact, got %d", resp.Summary.Returned)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Items[0]["id"] == nil {
		t.Error("expected contact to contain 'id' field")
	}
}

// TestCreateContactHandler_Success tests creating a contact over wire.
func TestCreateContactHandler_Success(t *testing.T) {
	cs, _ := setupWireTest(t, autotasktest.WithEntity(autotasktest.ContactFixture()))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_contact",
		Arguments: map[string]any{
			"companyID":    1001,
			"firstName":    "Alice",
			"lastName":     "Example",
			"emailAddress": "alice@example.com",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	m := parseStructuredContent[map[string]any](t, result)
	// firstName is externally supplied, so it must be framed as untrusted content:
	// assert both that it is framed AND that the resolved name survives.
	if fn, ok := m["firstName"].(string); !ok || !strings.Contains(fn, "Alice") || !strings.Contains(fn, "<untrusted_content>") {
		t.Errorf("expected framed firstName containing 'Alice', got %v", m["firstName"])
	}
}

// TestSearchContactsHandler_WithFilters verifies filters execution over wire.
func TestSearchContactsHandler_WithFilters(t *testing.T) {
	contact := autotasktest.ContactFixture()
	cs, _ := setupWireTest(t, autotasktest.WithEntity(contact))
	ctx := context.Background()

	active := 1
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_contacts",
		Arguments: map[string]any{
			"searchTerm": "Jane",
			"companyID":  1001,
			"isActive":   active,
			"maxResults": 10,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}
