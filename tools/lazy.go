package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CategoryInfo describes a category of tools.
type CategoryInfo struct {
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
}

// ToolCategories is the authoritative map of tool category names to their metadata.
var ToolCategories = map[string]CategoryInfo{
	"utility": {
		Description: "Connection testing and field/picklist discovery",
		Tools:       []string{"autotask_test_connection", "autotask_list_queues", "autotask_list_ticket_statuses", "autotask_list_ticket_priorities", "autotask_get_field_info"},
	},
	"companies": {
		Description: "Search, create, and update companies",
		Tools:       []string{"autotask_search_companies", "autotask_create_company", "autotask_update_company"},
	},
	"contacts": {
		Description: "Search and create contacts",
		Tools:       []string{"autotask_search_contacts", "autotask_create_contact"},
	},
	"tickets": {
		Description: "Search, create, update tickets and manage notes/attachments",
		Tools:       []string{"autotask_search_tickets", "autotask_get_ticket_details", "autotask_create_ticket", "autotask_update_ticket", "autotask_get_ticket_note", "autotask_search_ticket_notes", "autotask_create_ticket_note", "autotask_get_ticket_attachment", "autotask_search_ticket_attachments"},
	},
	"projects": {
		Description: "Search and create projects, tasks, and project notes",
		Tools:       []string{"autotask_search_projects", "autotask_create_project", "autotask_search_tasks", "autotask_create_task", "autotask_get_project_note", "autotask_search_project_notes", "autotask_create_project_note"},
	},
	"time_and_billing": {
		Description: "Time entries, billing items, and expense management",
		Tools:       []string{"autotask_create_time_entry", "autotask_search_time_entries", "autotask_search_billing_items", "autotask_get_billing_item", "autotask_search_billing_item_approval_levels", "autotask_get_expense_report", "autotask_search_expense_reports", "autotask_create_expense_report", "autotask_create_expense_item"},
	},
	"financial": {
		Description: "Quotes, quote items, opportunities, invoices, and contracts",
		Tools:       []string{"autotask_get_quote", "autotask_search_quotes", "autotask_create_quote", "autotask_get_quote_item", "autotask_search_quote_items", "autotask_create_quote_item", "autotask_update_quote_item", "autotask_delete_quote_item", "autotask_get_opportunity", "autotask_search_opportunities", "autotask_create_opportunity", "autotask_search_invoices", "autotask_search_contracts"},
	},
	"products_and_services": {
		Description: "Products, services, and service bundles catalog",
		Tools:       []string{"autotask_get_product", "autotask_search_products", "autotask_get_service", "autotask_search_services", "autotask_get_service_bundle", "autotask_search_service_bundles"},
	},
	"resources": {
		Description: "Search for Autotask resources",
		Tools:       []string{"autotask_search_resources"},
	},
	"configuration_items": {
		Description: "Search configuration items",
		Tools:       []string{"autotask_search_configuration_items"},
	},
	"company_notes": {
		Description: "Get, search, and create company notes",
		Tools:       []string{"autotask_get_company_note", "autotask_search_company_notes", "autotask_create_company_note"},
	},
}

// toolDescriptions maps tool names to their human-readable descriptions.
// Used by autotask_list_category_tools.
var toolDescriptions = map[string]string{
	"autotask_test_connection":                "Test connectivity to the Autotask API",
	"autotask_list_queues":                    "List available ticket queues",
	"autotask_list_ticket_statuses":           "List available ticket status values",
	"autotask_list_ticket_priorities":         "List available ticket priority values",
	"autotask_get_field_info":                 "Get field metadata for an entity type",
	"autotask_search_companies":               "Search for companies",
	"autotask_create_company":                 "Create a new company",
	"autotask_update_company":                 "Update an existing company",
	"autotask_search_contacts":                "Search for contacts",
	"autotask_create_contact":                 "Create a new contact",
	"autotask_search_tickets":                 "Search for tickets",
	"autotask_get_ticket_details":             "Get detailed information for a ticket",
	"autotask_create_ticket":                  "Create a new ticket",
	"autotask_update_ticket":                  "Update an existing ticket",
	"autotask_get_ticket_note":                "Get a specific ticket note by ID",
	"autotask_search_ticket_notes":            "Search ticket notes",
	"autotask_create_ticket_note":             "Create a new note on a ticket",
	"autotask_get_ticket_attachment":          "Get a ticket attachment by ID",
	"autotask_search_ticket_attachments":      "Search ticket attachments",
	"autotask_search_projects":               "Search for projects",
	"autotask_create_project":                "Create a new project",
	"autotask_search_tasks":                  "Search for project tasks",
	"autotask_create_task":                   "Create a new project task",
	"autotask_get_project_note":              "Get a project note by ID",
	"autotask_search_project_notes":          "Search project notes",
	"autotask_create_project_note":           "Create a new project note",
	"autotask_create_time_entry":             "Create a new time entry",
	"autotask_search_time_entries":           "Search time entries",
	"autotask_search_billing_items":          "Search billing items",
	"autotask_get_billing_item":              "Get a billing item by ID",
	"autotask_search_billing_item_approval_levels": "List billing item approval levels",
	"autotask_get_expense_report":            "Get an expense report by ID",
	"autotask_search_expense_reports":        "Search expense reports",
	"autotask_create_expense_report":         "Create a new expense report",
	"autotask_create_expense_item":           "Create a new expense item",
	"autotask_get_quote":                     "Get a quote by ID",
	"autotask_search_quotes":                 "Search quotes",
	"autotask_create_quote":                  "Create a new quote",
	"autotask_get_quote_item":                "Get a quote item by ID",
	"autotask_search_quote_items":            "Search quote items",
	"autotask_create_quote_item":             "Create a new quote item",
	"autotask_update_quote_item":             "Update an existing quote item",
	"autotask_delete_quote_item":             "Delete a quote item",
	"autotask_get_opportunity":               "Get an opportunity by ID",
	"autotask_search_opportunities":          "Search opportunities",
	"autotask_create_opportunity":            "Create a new opportunity",
	"autotask_search_invoices":               "Search invoices",
	"autotask_search_contracts":              "Search contracts",
	"autotask_get_product":                   "Get a product by ID",
	"autotask_search_products":               "Search products",
	"autotask_get_service":                   "Get a service by ID",
	"autotask_search_services":               "Search services",
	"autotask_get_service_bundle":            "Get a service bundle by ID",
	"autotask_search_service_bundles":        "Search service bundles",
	"autotask_search_resources":              "Search for Autotask resources (employees/contacts)",
	"autotask_search_configuration_items":    "Search configuration items",
	"autotask_get_company_note":              "Get a company note by ID",
	"autotask_search_company_notes":          "Search company notes",
	"autotask_create_company_note":           "Create a new company note",
}

// routingRules maps keywords to suggested tools for autotask_router.
var routingRules = []struct {
	keywords    []string
	tool        string
	description string
}{
	{[]string{"ticket", "issue", "problem", "request"}, "autotask_search_tickets", "Search for tickets"},
	{[]string{"create ticket", "new ticket", "open ticket"}, "autotask_create_ticket", "Create a new ticket"},
	{[]string{"company", "companies", "account", "client", "customer"}, "autotask_search_companies", "Search for companies"},
	{[]string{"contact", "contacts", "person", "user"}, "autotask_search_contacts", "Search for contacts"},
	{[]string{"time", "hours", "timesheet"}, "autotask_search_time_entries", "Search time entries"},
	{[]string{"project"}, "autotask_search_projects", "Search for projects"},
	{[]string{"task"}, "autotask_search_tasks", "Search for tasks"},
	{[]string{"billing", "invoice", "charge"}, "autotask_search_billing_items", "Search billing items"},
	{[]string{"expense", "cost"}, "autotask_search_expense_reports", "Search expense reports"},
	{[]string{"quote", "proposal"}, "autotask_search_quotes", "Search quotes"},
	{[]string{"opportunity", "deal", "sale"}, "autotask_search_opportunities", "Search opportunities"},
	{[]string{"contract"}, "autotask_search_contracts", "Search contracts"},
	{[]string{"product", "item", "sku"}, "autotask_search_products", "Search products"},
	{[]string{"service"}, "autotask_search_services", "Search services"},
	{[]string{"resource", "employee", "tech"}, "autotask_search_resources", "Search resources"},
	{[]string{"config", "configuration", "device", "asset"}, "autotask_search_configuration_items", "Search configuration items"},
}

// ListCategoriesInput has no required fields.
type ListCategoriesInput struct{}

// ListCategoryToolsInput defines input for the list_category_tools meta-tool.
type ListCategoryToolsInput struct {
	Category string `json:"category" jsonschema:"Category name (e.g. tickets, companies, projects)"`
}

// ExecuteToolInput defines input for the execute_tool meta-tool.
type ExecuteToolInput struct {
	ToolName  string         `json:"toolName" jsonschema:"The name of the tool to execute"`
	Arguments map[string]any `json:"arguments,omitempty" jsonschema:"Arguments to pass to the tool"`
}

// RouterInput defines input for the router meta-tool.
type RouterInput struct {
	Intent string `json:"intent" jsonschema:"Natural language description of what you want to do"`
}

// RegisterLazyTools registers the 4 lazy-loading meta-tools with the server.
func RegisterLazyTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_categories",
		Description: "List all available tool categories. Use this as a starting point to discover which tools are available for a given domain (tickets, companies, projects, etc.).",
	}, listCategoriesHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_category_tools",
		Description: "List the tools available in a specific category, with their descriptions. Call autotask_list_categories first to see available category names.",
	}, listCategoryToolsHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_execute_tool",
		Description: "Proxy tool for lazy-loading mode. Returns instructions to call the requested tool directly. In lazy-loading mode, individual tools are not registered; use this to get guidance on how to call them.",
	}, executeToolHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_router",
		Description: "Route a natural-language intent to the most appropriate Autotask tool. Provide a plain-English description of what you want to do and get a suggested tool name and description.",
	}, routerHandler())
}

// listCategoriesHandler returns a handler that returns the full category map.
func listCategoriesHandler() func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoriesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoriesInput) (*mcp.CallToolResult, any, error) {
		// Build a simplified summary with category name, description, and tool count.
		type categorySummary struct {
			Description string   `json:"description"`
			ToolCount   int      `json:"toolCount"`
			Tools       []string `json:"tools"`
		}
		summary := make(map[string]categorySummary, len(ToolCategories))
		for name, info := range ToolCategories {
			summary[name] = categorySummary{
				Description: info.Description,
				ToolCount:   len(info.Tools),
				Tools:       info.Tools,
			}
		}

		data, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return errorResult("failed to marshal categories: %v", err)
		}
		return textResult("%s", string(data))
	}
}

// listCategoryToolsHandler returns a handler that lists tools for a given category.
func listCategoryToolsHandler() func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoryToolsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoryToolsInput) (*mcp.CallToolResult, any, error) {
		cat, ok := ToolCategories[in.Category]
		if !ok {
			// Try case-insensitive match.
			lower := strings.ToLower(in.Category)
			for k, v := range ToolCategories {
				if strings.ToLower(k) == lower {
					cat = v
					ok = true
					break
				}
			}
		}
		if !ok {
			return errorResult("unknown category %q; call autotask_list_categories to see available categories", in.Category)
		}

		type toolEntry struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		tools := make([]toolEntry, 0, len(cat.Tools))
		for _, name := range cat.Tools {
			desc := toolDescriptions[name]
			if desc == "" {
				desc = name
			}
			tools = append(tools, toolEntry{Name: name, Description: desc})
		}

		result := map[string]any{
			"category":    in.Category,
			"description": cat.Description,
			"tools":       tools,
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return errorResult("failed to marshal tools: %v", err)
		}
		return textResult("%s", string(data))
	}
}

// executeToolHandler returns a handler that provides proxy/guidance for a named tool.
func executeToolHandler() func(ctx context.Context, req *mcp.CallToolRequest, in ExecuteToolInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ExecuteToolInput) (*mcp.CallToolResult, any, error) {
		if in.ToolName == "" {
			return errorResult("toolName is required")
		}

		desc := toolDescriptions[in.ToolName]
		if desc == "" {
			desc = "unknown tool"
		}

		var argHint string
		if len(in.Arguments) > 0 {
			data, _ := json.Marshal(in.Arguments)
			argHint = string(data)
		} else {
			argHint = "{}"
		}

		return textResult(
			"Please call the tool directly: %s\nDescription: %s\nSuggested arguments: %s\n\nIn lazy-loading mode, individual tools are not registered. Switch to full mode (LAZY_LOADING=false) to call tools directly, or use the MCP client to invoke the named tool.",
			in.ToolName, desc, argHint,
		)
	}
}

// routerHandler returns a handler that routes a natural language intent to a tool.
func routerHandler() func(ctx context.Context, req *mcp.CallToolRequest, in RouterInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in RouterInput) (*mcp.CallToolResult, any, error) {
		if in.Intent == "" {
			return errorResult("intent is required")
		}

		intentLower := strings.ToLower(in.Intent)
		suggestedTool := ""
		description := ""

		for _, rule := range routingRules {
			for _, kw := range rule.keywords {
				if strings.Contains(intentLower, kw) {
					suggestedTool = rule.tool
					description = rule.description
					break
				}
			}
			if suggestedTool != "" {
				break
			}
		}

		if suggestedTool == "" {
			suggestedTool = "autotask_list_categories"
			description = "No specific match found. Use autotask_list_categories to browse available tools."
		}

		result := map[string]any{
			"intent":        in.Intent,
			"suggestedTool": suggestedTool,
			"description":   description,
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return errorResult("failed to marshal result: %v", err)
		}
		return textResult("%s", string(data))
	}
}
