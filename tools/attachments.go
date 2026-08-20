package tools

import (
	"context"
	"errors"

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
func getTicketAttachmentHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetTicketAttachmentInput) (*mcp.CallToolResult, map[string]any, error) {
		attachment, err := autotask.Get[entities.TicketAttachment](ctx, client, in.AttachmentID)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(attachment)
		if err != nil {
			return nil, nil, err
		}

		delete(m, "fullPath")
		if !in.IncludeData {
			delete(m, "data")
		}

		return nil, m, nil
	}
}

// searchTicketAttachmentsHandler returns a handler that lists attachments for a ticket without heavy base64 data.
func searchTicketAttachmentsHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTicketAttachmentsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		attachments, err := autotask.ListChildRaw(ctx, client, "Tickets", in.TicketID, "TicketAttachments")
		if err != nil {
			var notFound *autotask.NotFoundError
			if errors.As(err, &notFound) {
				return emptySearchResult()
			}
			return nil, services.CompactResponse{}, err
		}

		if len(attachments) == 0 {
			return emptySearchResult()
		}

		maxResults := defaultMaxResults(in.MaxResults, 25, 100)
		hasMore := len(attachments) >= maxResults && maxResults > 0
		if len(attachments) > maxResults {
			attachments = attachments[:maxResults]
		}

		for _, attachment := range attachments {
			delete(attachment, "data")
			delete(attachment, "fullPath")
			services.FrameUntrustedMapFields(attachment)
		}

		hint := ""
		if hasMore {
			hint = "Maximum result limit reached. Use narrower search filters to find specific records."
		}

		return nil, services.CompactResponse{
			Summary: services.CompactSummary{
				Returned:   len(attachments),
				HasMore:    hasMore,
				MaxResults: maxResults,
				Hint:       hint,
			},
			Items: attachments,
		}, nil
	}
}
