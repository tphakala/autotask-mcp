package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetTicketAttachmentInput defines the input parameters for getting a ticket attachment.
type GetTicketAttachmentInput struct {
	AttachmentID int64 `json:"attachmentId" jsonschema:"Attachment ID to retrieve"`
}

// SearchTicketAttachmentsInput defines the input parameters for searching ticket attachments.
type SearchTicketAttachmentsInput struct {
	TicketID int64 `json:"ticketId" jsonschema:"Ticket ID to list attachments for"`
}

// RegisterAttachmentTools registers all attachment-related MCP tools with the server.
func RegisterAttachmentTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_attachment",
		Description: "Retrieve one ticket attachment by its numeric attachment ID, returning its full record. Use when you already have a specific attachment ID; to list every attachment belonging to a ticket use autotask_search_ticket_attachments instead. Read-only.",
		Annotations: readOnlyTool("Get ticket attachment"),
	}, getTicketAttachmentHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_ticket_attachments",
		Description: "List every attachment belonging to one ticket, identified by its ticketId, returning all attachment records for that ticket without pagination. Use this to discover a ticket's attachments, then autotask_get_ticket_attachment to retrieve one by its attachment ID. Read-only.",
		Annotations: readOnlyTool("Search ticket attachments"),
	}, searchTicketAttachmentsHandler(client))
}

// getTicketAttachmentHandler returns a handler that retrieves a single ticket attachment.
func getTicketAttachmentHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, any, error) {
		attachment, err := autotask.Get[entities.TicketAttachment](ctx, client, in.AttachmentID)
		if err != nil {
			return errorResult("failed to get ticket attachment %d: %v", in.AttachmentID, err)
		}

		m, err := entityToMap(attachment)
		if err != nil {
			return errorResult("failed to convert ticket attachment: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
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
