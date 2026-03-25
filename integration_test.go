package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/entities"
)

// connectMCP builds the MCP server from an autotask client, connects an
// in-memory MCP client to it, and returns the client session.
// The server and client sessions are cleaned up via t.Cleanup.
func connectMCP(t *testing.T, client *autotask.Client) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	s := buildServer(client, false)
	ct, st := mcp.NewInMemoryTransports()

	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return cs
}

// TestIntegration_SearchTickets verifies that the search_tickets tool
// returns a non-error response when the mock server has ticket fixtures.
func TestIntegration_SearchTickets(t *testing.T) {
	ctx := context.Background()

	ticket := autotasktest.TicketFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))
	cs := connectMCP(t, client)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_tickets",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}
}

// TestIntegration_SearchTickets_NoResults verifies that the tool handles
// an empty ticket list gracefully.
func TestIntegration_SearchTickets_NoResults(t *testing.T) {
	ctx := context.Background()

	_, client := autotasktest.NewServer(t)
	cs := connectMCP(t, client)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_tickets",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result even with no tickets")
	}
}

// TestIntegration_CreateAndGetTicket exercises the create_ticket and
// get_ticket_details tools end-to-end.
func TestIntegration_CreateAndGetTicket(t *testing.T) {
	ctx := context.Background()

	company := autotasktest.CompanyFixture()
	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(company))
	cs := connectMCP(t, client)

	companyID, ok := company.ID.Get()
	if !ok {
		t.Fatal("fixture company has no ID")
	}

	// Create a ticket.
	createResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_create_ticket",
		Arguments: map[string]any{
			"companyID":   companyID,
			"title":       "Integration test ticket",
			"description": "Created by integration test",
		},
	})
	if err != nil {
		t.Fatalf("create_ticket CallTool failed: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("create_ticket returned error: %v", createResult.Content)
	}

	// Parse and verify the create_ticket response contains expected fields.
	textContent, ok := createResult.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent in create_ticket response")
	}

	var ticketResp map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &ticketResp); err != nil {
		t.Fatalf("failed to parse create_ticket response: %v", err)
	}

	// The response should include the title that was provided.
	if ticketResp["title"] != "Integration test ticket" {
		t.Errorf("expected title in response, got: %v", ticketResp["title"])
	}

	// Now seed a ticket directly and retrieve it by ID.
	ticket := autotasktest.TicketFixture()
	ticketID, ticketOK := ticket.ID.Get()
	if !ticketOK {
		t.Fatal("fixture ticket has no ID")
	}
	_, client2 := autotasktest.NewServer(t, autotasktest.WithEntity(ticket))
	cs2 := connectMCP(t, client2)

	getResult, err := cs2.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_ticket_details",
		Arguments: map[string]any{
			"ticketID": ticketID,
		},
	})
	if err != nil {
		t.Fatalf("get_ticket_details CallTool failed: %v", err)
	}
	if getResult.IsError {
		t.Fatalf("get_ticket_details returned error: %v", getResult.Content)
	}
}

// TestIntegration_SearchCompanies verifies the search_companies tool returns
// seeded company data.
func TestIntegration_SearchCompanies(t *testing.T) {
	ctx := context.Background()

	alpha := autotasktest.CompanyFixture(func(c *entities.Company) {
		c.CompanyName = autotask.Set("Alpha Corp")
	})
	beta := autotasktest.CompanyFixture(func(c *entities.Company) {
		c.CompanyName = autotask.Set("Beta LLC")
	})

	_, client := autotasktest.NewServer(t, autotasktest.WithEntity(alpha, beta))
	cs := connectMCP(t, client)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_search_companies",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected non-error result, got: %v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}
}

// TestIntegration_BuildServer_ToolsRegistered verifies that buildServer
// registers the expected set of tools.
func TestIntegration_BuildServer_ToolsRegistered(t *testing.T) {
	ctx := context.Background()

	_, client := autotasktest.NewServer(t)
	cs := connectMCP(t, client)

	// Collect all tools.
	toolNames := map[string]bool{}
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		toolNames[tool.Name] = true
	}

	// Verify a representative set of expected tools.
	expectedTools := []string{
		"autotask_search_tickets",
		"autotask_create_ticket",
		"autotask_search_companies",
		"autotask_search_contacts",
		"autotask_search_time_entries",
		"autotask_test_connection",
	}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}
