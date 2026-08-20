package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/entities"
)

// TestRegisterAttachmentTools_NoPanic verifies registration does not panic.
func TestRegisterAttachmentTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterAttachmentTools(s, client)
}

// TestGetTicketAttachmentHandler_NotFound tests that a missing attachment returns an error result.
func TestGetTicketAttachmentHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getTicketAttachmentHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketAttachmentInput{AttachmentID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing attachment")
	}
}

// TestSearchTicketAttachmentsHandler_NoPanic verifies the search handler does not panic.
func TestSearchTicketAttachmentsHandler_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchTicketAttachmentsHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchTicketAttachmentsInput{TicketID: 3001})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Note: IsError may be true when mock server has no ticket store for the parent ID.
	// The key assertion is that the handler doesn't panic and returns a valid result.
}

// TestGetTicketAttachmentHandler_StripDataByDefault tests that attachment data is stripped by default.
func TestGetTicketAttachmentHandler_StripDataByDefault(t *testing.T) {
	att := &entities.TicketAttachment{
		ID:       autotask.Set(int64(7001)),
		TicketID: autotask.Set(int64(3001)),
		Title:    autotask.Set("test attachment"),
		Data:     autotask.Set("SGVsbG8gV29ybGQ="),
		FullPath: autotask.Set("/path/to/test.txt"),
	}
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(att))

	handler := getTicketAttachmentHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketAttachmentInput{AttachmentID: 7001, IncludeData: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, exists := m["data"]; exists {
		t.Errorf("expected 'data' field to be stripped, but found: %v", m["data"])
	}
	if _, exists := m["fullPath"]; exists {
		t.Errorf("expected 'fullPath' field to be stripped, but found: %v", m["fullPath"])
	}
}

// TestGetTicketAttachmentHandler_IncludeData tests that attachment data is included when requested.
func TestGetTicketAttachmentHandler_IncludeData(t *testing.T) {
	att := &entities.TicketAttachment{
		ID:       autotask.Set(int64(7001)),
		TicketID: autotask.Set(int64(3001)),
		Title:    autotask.Set("test attachment"),
		Data:     autotask.Set("SGVsbG8gV29ybGQ="),
		FullPath: autotask.Set("/path/to/test.txt"),
	}
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(att))

	handler := getTicketAttachmentHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetTicketAttachmentInput{AttachmentID: 7001, IncludeData: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, exists := m["data"]; !exists {
		t.Errorf("expected 'data' field to be present when includeData=true")
	}
}
