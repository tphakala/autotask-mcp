package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterLazyTools_DoesNotPanic(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)
	// Should not panic.
	RegisterLazyTools(s)
}

func TestListCategories_ReturnsExpectedCategories(t *testing.T) {
	handler := listCategoriesHandler()

	result, _, err := handler(context.Background(), nil, ListCategoriesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	// Parse the JSON response.
	var categories map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &categories); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Check all expected categories are present.
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

	// Verify tickets category has expected tools.
	ticketsCat, ok := ToolCategories["tickets"]
	if !ok {
		t.Fatal("expected tickets category")
	}
	if len(ticketsCat.Tools) == 0 {
		t.Error("tickets category should have tools")
	}
	// Check that autotask_search_tickets is in there.
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
	handler := listCategoryToolsHandler()

	result, _, err := handler(context.Background(), nil, ListCategoryToolsInput{Category: "tickets"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatal("expected successful result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["category"] != "tickets" {
		t.Errorf("expected category=tickets, got %v", resp["category"])
	}
	tools, ok := resp["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Error("expected non-empty tools array")
	}
}

func TestListCategoryTools_UnknownCategory(t *testing.T) {
	handler := listCategoryToolsHandler()

	result, _, err := handler(context.Background(), nil, ListCategoryToolsInput{Category: "nonexistent_category_xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown category")
	}
}

func TestExecuteTool_ProxyResponse(t *testing.T) {
	handler := executeToolHandler()

	result, _, err := handler(context.Background(), nil, ExecuteToolInput{
		ToolName:  "autotask_search_tickets",
		Arguments: map[string]any{"status": 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatal("expected successful result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if !strings.Contains(textContent.Text, "autotask_search_tickets") {
		t.Errorf("expected response to mention tool name, got: %s", textContent.Text)
	}
}

func TestExecuteTool_EmptyToolName(t *testing.T) {
	handler := executeToolHandler()

	result, _, err := handler(context.Background(), nil, ExecuteToolInput{ToolName: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty toolName")
	}
}

func TestRouter_MatchesKeywords(t *testing.T) {
	handler := routerHandler()

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
			result, _, err := handler(context.Background(), nil, RouterInput{Intent: tt.intent})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil || result.IsError {
				t.Fatal("expected successful result")
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatal("expected TextContent")
			}

			var resp map[string]any
			if err := json.Unmarshal([]byte(textContent.Text), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp["suggestedTool"] != tt.expectsTool {
				t.Errorf("intent %q: expected suggestedTool=%q, got %q", tt.intent, tt.expectsTool, resp["suggestedTool"])
			}
		})
	}
}

func TestRouter_EmptyIntent(t *testing.T) {
	handler := routerHandler()

	result, _, err := handler(context.Background(), nil, RouterInput{Intent: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty intent")
	}
}

// TestToolCategories_AllToolsHaveDescriptions ensures every tool listed in
// ToolCategories has a corresponding entry in toolDescriptions, and vice versa.
// This prevents drift when tools are added to RegisterAll but not to lazy.go.
func TestToolCategories_AllToolsHaveDescriptions(t *testing.T) {
	// Collect all tool names from ToolCategories.
	categoryTools := map[string]bool{}
	for _, cat := range ToolCategories {
		for _, tool := range cat.Tools {
			categoryTools[tool] = true
		}
	}

	// Every tool in ToolCategories must have a description.
	for tool := range categoryTools {
		if _, ok := toolDescriptions[tool]; !ok {
			t.Errorf("tool %q is in ToolCategories but missing from toolDescriptions", tool)
		}
	}

	// Every tool in toolDescriptions must be in some category.
	for tool := range toolDescriptions {
		if !categoryTools[tool] {
			t.Errorf("tool %q is in toolDescriptions but not in any ToolCategories category", tool)
		}
	}
}

// TestToolCategories_NoDuplicates ensures no tool appears in multiple categories.
func TestToolCategories_NoDuplicates(t *testing.T) {
	seen := map[string]string{} // tool → category
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
	handler := routerHandler()

	result, _, err := handler(context.Background(), nil, RouterInput{Intent: "xyzzy something completely unknown"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatal("expected successful result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["suggestedTool"] != "autotask_list_categories" {
		t.Errorf("expected fallback to autotask_list_categories, got %v", resp["suggestedTool"])
	}
}
