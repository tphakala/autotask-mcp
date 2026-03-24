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
		Description: "Test the connection to Autotask and verify credentials are valid.",
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
