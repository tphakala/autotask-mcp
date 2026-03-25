package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
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
