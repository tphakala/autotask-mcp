package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterContactTools_NoPanic verifies that RegisterContactTools registers all
// two tools without panicking.
func TestRegisterContactTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)

	// Should not panic.
	RegisterContactTools(s, client, mapper)
}

// TestSearchContactsHandler_ReturnsNoContactsFound tests the empty-result case.
func TestSearchContactsHandler_ReturnsNoContactsFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := newTestMapper(client)

	handler := searchContactsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchContactsInput{})
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
	if text.Text != "No contacts found" {
		t.Errorf("expected 'No contacts found', got %q", text.Text)
	}
}

// TestSearchContactsHandler_ReturnsContacts tests that seeded contacts are returned.
func TestSearchContactsHandler_ReturnsContacts(t *testing.T) {
	contact := autotasktest.ContactFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(contact))
	mapper := newTestMapper(client)

	handler := searchContactsHandler(client, mapper)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchContactsInput{})
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

	var resp map[string]any
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		t.Fatalf("result is not valid JSON: %v\ncontent: %s", err, text.Text)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatalf("expected 'items' array in response, got: %v", resp)
	}
	if len(items) == 0 {
		t.Error("expected at least one contact in results")
	}
}

// TestCreateContactHandler_Success tests that a contact can be created.
func TestCreateContactHandler_Success(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(autotasktest.ContactFixture()),
	)

	handler := createContactHandler(client)
	ctx := context.Background()

	in := CreateContactInput{
		CompanyID:    1001,
		FirstName:    "Alice",
		LastName:     "Example",
		EmailAddress: "alice@example.com",
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

// TestSearchContactsHandler_WithFilters verifies that multiple filters can be applied.
func TestSearchContactsHandler_WithFilters(t *testing.T) {
	contact := autotasktest.ContactFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(contact))
	mapper := newTestMapper(client)

	handler := searchContactsHandler(client, mapper)
	ctx := context.Background()

	active := 1
	in := SearchContactsInput{
		SearchTerm: "Jane",
		CompanyID:  1001,
		IsActive:   &active,
		Page:       1,
		PageSize:   10,
	}

	result, _, err := handler(ctx, nil, in)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	_ = result
}
