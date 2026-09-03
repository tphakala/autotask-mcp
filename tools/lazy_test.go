package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
)

func TestRegisterLazyTools_DoesNotPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	picklist := services.NewPicklistCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)

	RegisterLazyTools(s, client, mapper, picklist)
}

func TestListCategories_ReturnsExpectedCategories(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_list_categories",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected IsError=false, got: %v", result.Content)
	}

	categories := parseStructuredContent[map[string]CategorySummary](t, result)
	expectedCategories := []string{
		"utility", "companies", "contacts", "tickets", "projects",
		"time_and_billing", "financial", "products_and_services",
		"resources", "configuration_items", "company_notes",
	}
	for _, cat := range expectedCategories {
		if _, ok := categories[cat]; !ok {
			t.Errorf("expected category %q to be present", cat)
		}
	}
}

func TestListCategories_ToolCategoriesMap(t *testing.T) {
	if len(ToolCategories) == 0 {
		t.Fatal("ToolCategories should not be empty")
	}

	ticketsCat, ok := ToolCategories["tickets"]
	if !ok {
		t.Fatal("expected tickets category")
	}
	if len(ticketsCat.Tools) == 0 {
		t.Error("tickets category should have tools")
	}
	found := false
	for _, tool := range ticketsCat.Tools {
		if tool == "autotask_search_tickets" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected autotask_search_tickets in tickets category")
	}
}

func TestListCategoryTools_KnownCategory(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_list_category_tools",
		Arguments: map[string]any{
			"category": "tickets",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful result, got: %v", result)
	}

	resp := parseStructuredContent[CategoryToolsOut](t, result)
	if resp.Category != "tickets" {
		t.Errorf("expected category=tickets, got %v", resp.Category)
	}
	if len(resp.Tools) == 0 {
		t.Error("expected non-empty tools array")
	}
}

func TestListCategoryTools_UnknownCategory(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_list_category_tools",
		Arguments: map[string]any{
			"category": "nonexistent_category_xyz",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown category")
	}
}

func TestExecuteTool_DispatchesRealTool(t *testing.T) {
	ticket := autotasktest.TicketFixture()
	cs, _ := setupLazyWireTest(t, autotasktest.WithEntity(ticket))
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_execute_tool",
		Arguments: map[string]any{
			"toolName":  "autotask_search_tickets",
			"arguments": map[string]any{"maxResults": 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful execution result, got: %v", result)
	}

	resp := parseStructuredContent[services.CompactResponse](t, result)
	if resp.Summary.Returned < 1 {
		t.Errorf("expected at least 1 returned ticket, got %d", resp.Summary.Returned)
	}
}

func TestExecuteTool_UnknownTool(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_execute_tool",
		Arguments: map[string]any{
			"toolName": "nonexistent_tool_name",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unknown tool")
	}
}

func TestExecuteTool_EmptyToolName(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_execute_tool",
		Arguments: map[string]any{
			"toolName": "",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty toolName")
	}
}

// TestExecuteTool_RejectsMissingRequiredArgs verifies the lazy proxy enforces the
// same required-field validation a direct tool call gets. search_ticket_notes
// requires ticketId; with it absent the handler would otherwise run against ticket
// 0 and return an empty success, so an error here proves schema validation fired
// before dispatch (see makeRunner). This is the #39 fix: without it, a call like
// {"toolName":"autotask_delete_quote_item","arguments":{}} would reach the handler
// with zero IDs.
func TestExecuteTool_RejectsMissingRequiredArgs(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_execute_tool",
		Arguments: map[string]any{
			"toolName":  "autotask_search_ticket_notes",
			"arguments": map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true: a required field is missing and validation must reject before dispatch")
	}
}

// TestExecuteTool_RejectsUnknownArg verifies the lazy proxy rejects unknown
// arguments (additionalProperties:false) exactly as a direct call would. A handler
// silently ignores unknown JSON fields, so only schema validation catches this.
func TestExecuteTool_RejectsUnknownArg(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_execute_tool",
		Arguments: map[string]any{
			"toolName":  "autotask_search_tickets",
			"arguments": map[string]any{"bogusField": "x"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true: an unknown argument must be rejected by additionalProperties:false")
	}
}

func TestDispatcher_AllCategoryToolsCovered(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	mapper := services.NewMappingCache(client)
	picklist := services.NewPicklistCache(client)
	dispatcher := buildToolDispatcher(client, mapper, picklist)

	toolCount := 0
	for catName, cat := range ToolCategories {
		toolCount += len(cat.Tools)
		for _, toolName := range cat.Tools {
			if _, ok := dispatcher[toolName]; !ok {
				t.Errorf("tool %q in category %q is not covered by dispatcher", toolName, catName)
			}
		}
	}

	if len(dispatcher) != toolCount {
		t.Errorf("dispatcher has %d tools, but ToolCategories defines %d", len(dispatcher), toolCount)
	}
}

func TestRouter_MatchesKeywords(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	tests := []struct {
		intent      string
		expectsTool string
	}{
		{"I want to find a ticket", "autotask_search_tickets"},
		{"search for companies", "autotask_search_companies"},
		{"log some hours", "autotask_search_time_entries"},
		{"show me projects", "autotask_search_projects"},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "autotask_router",
				Arguments: map[string]any{
					"intent": tt.intent,
				},
			})
			if err != nil {
				t.Fatalf("unexpected wire error: %v", err)
			}
			if result.IsError {
				t.Fatalf("expected successful result, got: %v", result)
			}

			resp := parseStructuredContent[RouterOut](t, result)
			if resp.SuggestedTool != tt.expectsTool {
				t.Errorf("intent %q: expected suggestedTool=%q, got %q", tt.intent, tt.expectsTool, resp.SuggestedTool)
			}
		})
	}
}

func TestRouter_EmptyIntent(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_router",
		Arguments: map[string]any{
			"intent": "",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty intent")
	}
}

// TestToolCategories_AllToolsHaveDescriptions ensures every tool listed in
// ToolCategories has a corresponding entry in toolDescriptions, and vice versa.
func TestToolCategories_AllToolsHaveDescriptions(t *testing.T) {
	categoryTools := map[string]bool{}
	for _, cat := range ToolCategories {
		for _, tool := range cat.Tools {
			categoryTools[tool] = true
		}
	}

	for tool := range categoryTools {
		if _, ok := toolDescriptions[tool]; !ok {
			t.Errorf("tool %q is in ToolCategories but missing from toolDescriptions", tool)
		}
	}

	for tool := range toolDescriptions {
		if !categoryTools[tool] {
			t.Errorf("tool %q is in toolDescriptions but not in any ToolCategories category", tool)
		}
	}
}

// TestToolCategories_NoDuplicates ensures no tool appears in multiple categories.
func TestToolCategories_NoDuplicates(t *testing.T) {
	seen := map[string]string{}
	for catName, cat := range ToolCategories {
		for _, tool := range cat.Tools {
			if prevCat, ok := seen[tool]; ok {
				t.Errorf("tool %q appears in both %q and %q categories", tool, prevCat, catName)
			}
			seen[tool] = catName
		}
	}
}

func TestRouter_FallbackToListCategories(t *testing.T) {
	cs, _ := setupLazyWireTest(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_router",
		Arguments: map[string]any{
			"intent": "xyzzy something completely unknown",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful result, got: %v", result)
	}

	resp := parseStructuredContent[RouterOut](t, result)
	if resp.SuggestedTool != "autotask_list_categories" {
		t.Errorf("expected fallback to autotask_list_categories, got %v", resp.SuggestedTool)
	}
}
