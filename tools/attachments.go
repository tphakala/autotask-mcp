package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
)

// GetTicketAttachmentInput defines the input parameters for getting a ticket attachment.
type GetTicketAttachmentInput struct {
	TicketID     int64 `json:"ticketId" jsonschema:"Ticket ID that owns the attachment"`
	AttachmentID int64 `json:"attachmentId" jsonschema:"Attachment ID to retrieve"`
	IncludeData  bool  `json:"includeData,omitempty" jsonschema:"Whether to include the attachment binary data"`
}

// SearchTicketAttachmentsInput defines the input parameters for searching ticket attachments.
type SearchTicketAttachmentsInput struct {
	TicketID int64 `json:"ticketId" jsonschema:"Ticket ID to list attachments for"`
	PageSize int   `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 500)"`
}

// RegisterAttachmentTools registers all attachment-related MCP tools with the server.
func RegisterAttachmentTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_attachment",
		Description: "Get a specific attachment for a ticket by attachment ID.",
	}, getTicketAttachmentHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_ticket_attachments",
		Description: "List all attachments for a ticket.",
	}, searchTicketAttachmentsHandler(client))
}

// getTicketAttachmentHandler returns a handler that retrieves a single ticket attachment.
func getTicketAttachmentHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, any, error) {
		attachment, err := autotask.GetRaw(ctx, client, "TicketAttachments", in.AttachmentID)
		if err != nil {
			return errorResult("failed to get ticket attachment %d: %v", in.AttachmentID, err)
		}

		data, err := json.MarshalIndent(attachment, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket attachment: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchTicketAttachmentsHandler returns a handler that lists all attachments for a ticket.
func searchTicketAttachmentsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, any, error) {
		attachments, err := autotask.ListChildRaw(ctx, client, "Tickets", in.TicketID, "TicketAttachments")
		if err != nil {
			return errorResult("failed to list ticket attachments for ticket %d: %v", in.TicketID, err)
		}

		if len(attachments) == 0 {
			return textResult("No ticket attachments found")
		}

		data, err := json.MarshalIndent(attachments, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket attachments: %v", err)
		}

		return textResult("%s", string(data))
	}
}
