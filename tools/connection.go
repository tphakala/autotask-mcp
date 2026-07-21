package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/metadata"
)

// RegisterConnectionTools registers the connection test MCP tool with the server.
func RegisterConnectionTools(s *mcp.Server, client *autotask.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_test_connection",
		Description: "Verify that the configured Autotask API credentials authenticate by fetching entity metadata for the Tickets entity, and report its canCreate, canUpdate, and canQuery permission flags. Takes no arguments; use this first to confirm connectivity and permissions before calling data tools such as autotask_search_tickets. Read-only.",
		Annotations: readOnlyTool("Test connection"),
	}, testConnectionHandler(client))
}

// testConnectionHandler returns a handler that tests the Autotask API connection.
func testConnectionHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		info, err := metadata.GetEntityInfo(ctx, client, "Tickets")
		if err != nil {
			return errorResult("connection test failed: %v", err)
		}

		return textResult("Connection successful. Tickets entity: canCreate=%v canUpdate=%v canQuery=%v",
			info.CanCreate, info.CanUpdate, info.CanQuery)
	}
}
