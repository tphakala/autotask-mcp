package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// GetTicketAttachmentInput defines the input parameters for getting a ticket attachment.
type GetTicketAttachmentInput struct {
	AttachmentID int64 `json:"attachmentId" jsonschema:"Attachment ID to retrieve"`
	IncludeData  bool  `json:"includeData,omitempty" jsonschema:"Set to true to include raw base64 file data (defaults to false to conserve context)"`
}

// SearchTicketAttachmentsInput defines the input parameters for searching ticket attachments.
type SearchTicketAttachmentsInput struct {
	TicketID   int64 `json:"ticketId" jsonschema:"Ticket ID to list attachments for"`
	MaxResults int   `json:"maxResults,omitempty" jsonschema:"Max attachments to return (default 25, max 100)"`
}

// RegisterAttachmentTools registers all attachment-related MCP tools with the server.
func RegisterAttachmentTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_get_ticket_attachment",
		Description: "Retrieve one ticket attachment by its numeric attachment ID. Returns attachment metadata by default (title, file size, content type); set includeData=true to retrieve raw base64 file data. Read-only.",
		Annotations: readOnlyTool("Get ticket attachment"),
	}, getTicketAttachmentHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_ticket_attachments",
		Description: "List attachments belonging to one ticket by ticketId, returning up to maxResults metadata records (default 25, max 100) with base64 data and internal file paths omitted. Read-only.",
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

		delete(m, "fullPath")
		if !in.IncludeData {
			delete(m, "data")
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket attachment: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// searchTicketAttachmentsHandler returns a handler that lists attachments for a ticket without heavy base64 data.
func searchTicketAttachmentsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, any, error) {
		attachments, err := autotask.ListChildRaw(ctx, client, "Tickets", in.TicketID, "TicketAttachments")
		if err != nil {
			return errorResult("failed to list ticket attachments for ticket %d: %v", in.TicketID, err)
		}

		if len(attachments) == 0 {
			return textResult("No ticket attachments found")
		}

		maxResults := defaultMaxResults(in.MaxResults, 25, 100)
		if len(attachments) > maxResults {
			attachments = attachments[:maxResults]
		}

		for _, attachment := range attachments {
			delete(attachment, "data")
			delete(attachment, "fullPath")
			services.FrameUntrustedMapFields(attachment)
		}

		data, err := json.MarshalIndent(attachments, "", "  ")
		if err != nil {
			return errorResult("failed to marshal ticket attachments: %v", err)
		}

		return textResult("%s", string(data))
	}
}
