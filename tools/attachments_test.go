package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
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

// TestGetTicketAttachmentHandler_NotFound tests that a missing attachment returns an error result over wire.
func TestGetTicketAttachmentHandler_NotFound(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_attachment",
		Arguments: map[string]any{
			"attachmentId": 99999,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing attachment")
	}
}

// TestSearchTicketAttachmentsHandler_NoResults verifies empty search response over wire.
func TestSearchTicketAttachmentsHandler_NoResults(t *testing.T) {
	cs, _ := setupWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_attachments",
		Arguments: map[string]any{
			"ticketId": 3001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 0 {
		t.Errorf("expected 0 returned attachments, got %d", resp.Summary.Returned)
	}
}

// TestSearchTicketAttachmentsHandler_WithResults verifies seeded attachments are returned with data stripped.
func TestSearchTicketAttachmentsHandler_WithResults(t *testing.T) {
	att := &entities.TicketAttachment{
		ID:       autotask.Set(int64(7001)),
		TicketID: autotask.Set(int64(3001)),
		Title:    autotask.Set("test attachment"),
		Data:     autotask.Set("SGVsbG8gV29ybGQ="),
		FullPath: autotask.Set("/path/to/test.txt"),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(att))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_attachments",
		Arguments: map[string]any{
			"ticketId": 3001,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Fatalf("expected at least 1 attachment, got %d", resp.Summary.Returned)
	}
	if _, exists := resp.Items[0]["data"]; exists {
		t.Errorf("expected 'data' field to be stripped from search results")
	}
	if _, exists := resp.Items[0]["fullPath"]; exists {
		t.Errorf("expected 'fullPath' field to be stripped from search results")
	}
}

// TestSearchTicketAttachmentsHandler_BoundedByMaxResults seeds more attachments
// than the requested limit and asserts the handler returns exactly maxResults with
// hasMore reported, and that the heavy base64 data and internal path are stripped
// from the bounded set. Bounding matters most here: each row carries a base64 blob.
func TestSearchTicketAttachmentsHandler_BoundedByMaxResults(t *testing.T) {
	atts := make([]*entities.TicketAttachment, 0, 5)
	for i := 0; i < 5; i++ {
		atts = append(atts, &entities.TicketAttachment{
			ID:       autotask.Set(int64(7000 + i)),
			TicketID: autotask.Set(int64(3001)),
			Title:    autotask.Set("attachment"),
			Data:     autotask.Set("SGVsbG8gV29ybGQ="),
			FullPath: autotask.Set("/path/to/test.txt"),
		})
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(atts...))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_search_ticket_attachments",
		Arguments: map[string]any{
			"ticketId":   3001,
			"maxResults": 2,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned != 2 {
		t.Errorf("expected 2 returned attachments (bounded to maxResults), got %d", resp.Summary.Returned)
	}
	if !resp.Summary.HasMore {
		t.Error("expected hasMore=true when more attachments exist beyond maxResults")
	}
	for _, item := range resp.Items {
		if _, exists := item["data"]; exists {
			t.Errorf("expected 'data' stripped from search results")
		}
		if _, exists := item["fullPath"]; exists {
			t.Errorf("expected 'fullPath' stripped from search results")
		}
	}
}

// TestGetTicketAttachmentHandler_StripDataByDefault tests that attachment data is stripped by default over wire.
func TestGetTicketAttachmentHandler_StripDataByDefault(t *testing.T) {
	att := &entities.TicketAttachment{
		ID:       autotask.Set(int64(7001)),
		TicketID: autotask.Set(int64(3001)),
		Title:    autotask.Set("test attachment"),
		Data:     autotask.Set("SGVsbG8gV29ybGQ="),
		FullPath: autotask.Set("/path/to/test.txt"),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(att))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_attachment",
		Arguments: map[string]any{
			"attachmentId": 7001,
			"includeData":  false,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	m := parseStructuredContent[map[string]any](t, result)
	if _, exists := m["data"]; exists {
		t.Errorf("expected 'data' field to be stripped, but found: %v", m["data"])
	}
	if _, exists := m["fullPath"]; exists {
		t.Errorf("expected 'fullPath' field to be stripped, but found: %v", m["fullPath"])
	}
}

// TestGetTicketAttachmentHandler_IncludeData tests that attachment data is included when requested over wire.
func TestGetTicketAttachmentHandler_IncludeData(t *testing.T) {
	att := &entities.TicketAttachment{
		ID:       autotask.Set(int64(7001)),
		TicketID: autotask.Set(int64(3001)),
		Title:    autotask.Set("test attachment"),
		Data:     autotask.Set("SGVsbG8gV29ybGQ="),
		FullPath: autotask.Set("/path/to/test.txt"),
	}
	cs, _ := setupWireTest(t, autotasktest.WithEntity(att))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_attachment",
		Arguments: map[string]any{
			"attachmentId": 7001,
			"includeData":  true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result)
	}

	m := parseStructuredContent[map[string]any](t, result)
	if _, exists := m["data"]; !exists {
		t.Errorf("expected 'data' field to be present when includeData=true")
	}
}
