package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
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
	"autotask_test_connection":                     "Test connectivity to the Autotask API",
	"autotask_list_queues":                         "List available ticket queues",
	"autotask_list_ticket_statuses":                "List available ticket status values",
	"autotask_list_ticket_priorities":              "List available ticket priority values",
	"autotask_get_field_info":                      "Get field metadata for an entity type",
	"autotask_search_companies":                    "Search for companies",
	"autotask_create_company":                      "Create a new company",
	"autotask_update_company":                      "Update an existing company",
	"autotask_search_contacts":                     "Search for contacts",
	"autotask_create_contact":                      "Create a new contact",
	"autotask_search_tickets":                      "Search for tickets",
	"autotask_get_ticket_details":                  "Get detailed information for a ticket",
	"autotask_create_ticket":                       "Create a new ticket",
	"autotask_update_ticket":                       "Update an existing ticket",
	"autotask_get_ticket_note":                     "Get a specific ticket note by ID",
	"autotask_search_ticket_notes":                 "Search ticket notes",
	"autotask_create_ticket_note":                  "Create a new note on a ticket",
	"autotask_get_ticket_attachment":               "Get a ticket attachment by ID",
	"autotask_search_ticket_attachments":           "Search ticket attachments",
	"autotask_search_projects":                     "Search for projects",
	"autotask_create_project":                      "Create a new project",
	"autotask_search_tasks":                        "Search for project tasks",
	"autotask_create_task":                         "Create a new project task",
	"autotask_get_project_note":                    "Get a project note by ID",
	"autotask_search_project_notes":                "Search project notes",
	"autotask_create_project_note":                 "Create a new project note",
	"autotask_create_time_entry":                   "Create a new time entry",
	"autotask_search_time_entries":                 "Search time entries",
	"autotask_search_billing_items":                "Search billing items",
	"autotask_get_billing_item":                    "Get a billing item by ID",
	"autotask_search_billing_item_approval_levels": "List billing item approval levels",
	"autotask_get_expense_report":                  "Get an expense report by ID",
	"autotask_search_expense_reports":              "Search expense reports",
	"autotask_create_expense_report":               "Create a new expense report",
	"autotask_create_expense_item":                 "Create a new expense item",
	"autotask_get_quote":                           "Get a quote by ID",
	"autotask_search_quotes":                       "Search quotes",
	"autotask_create_quote":                        "Create a new quote",
	"autotask_get_quote_item":                      "Get a quote item by ID",
	"autotask_search_quote_items":                  "Search quote items",
	"autotask_create_quote_item":                   "Create a new quote item",
	"autotask_update_quote_item":                   "Update an existing quote item",
	"autotask_delete_quote_item":                   "Delete a quote item",
	"autotask_get_opportunity":                     "Get an opportunity by ID",
	"autotask_search_opportunities":                "Search opportunities",
	"autotask_create_opportunity":                  "Create a new opportunity",
	"autotask_search_invoices":                     "Search invoices",
	"autotask_search_contracts":                    "Search contracts",
	"autotask_get_product":                         "Get a product by ID",
	"autotask_search_products":                     "Search products",
	"autotask_get_service":                         "Get a service by ID",
	"autotask_search_services":                     "Search services",
	"autotask_get_service_bundle":                  "Get a service bundle by ID",
	"autotask_search_service_bundles":              "Search service bundles",
	"autotask_search_resources":                    "Search for Autotask resources (employees/contacts)",
	"autotask_search_configuration_items":          "Search configuration items",
	"autotask_get_company_note":                    "Get a company note by ID",
	"autotask_search_company_notes":                "Search company notes",
	"autotask_create_company_note":                 "Create a new company note",
}

// routingRules maps keywords to suggested tools for autotask_router.
var routingRules = []struct {
	keywords    []string
	tool        string
	description string
}{
	{[]string{"create ticket", "new ticket", "open ticket"}, "autotask_create_ticket", "Create a new ticket"},
	{[]string{"ticket", "issue", "problem", "request"}, "autotask_search_tickets", "Search for tickets"},
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

// ToolRunner represents a dynamic dispatch function for executing a tool with generic JSON arguments.
type ToolRunner func(ctx context.Context, rawArgs map[string]any) (*mcp.CallToolResult, any, error)

// resolveInputSchema builds the resolved JSON schema for a tool's input type the
// same way the MCP SDK does when registering a tool (jsonschema.For + Resolve with
// ValidateDefaults), so lazy dispatch can enforce the identical contract. It returns
// nil if inference or resolution fails; callers then skip validation rather than
// reject every call for a schema that never validated anything to begin with.
func resolveInputSchema[In any]() *jsonschema.Resolved {
	schema, err := jsonschema.For[In](nil)
	if err != nil {
		return nil
	}
	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	if err != nil {
		return nil
	}
	return resolved
}

// makeRunner converts a typed MCP tool handler into a dynamic ToolRunner.
//
// A direct MCP tool call is validated by the SDK against the tool's inferred input
// schema (required fields present, no unknown properties) before the handler runs.
// The lazy autotask_execute_tool proxy dispatches here instead, so makeRunner must
// reproduce that validation itself: without it a call like
// {"toolName":"autotask_delete_quote_item","arguments":{}} would reach the delete
// handler with zero IDs. The schema is resolved once per tool at dispatcher build
// time and reused for every call.
func makeRunner[In, Out any](handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) ToolRunner {
	resolved := resolveInputSchema[In]()
	return func(ctx context.Context, rawArgs map[string]any) (*mcp.CallToolResult, any, error) {
		if handler == nil {
			return nil, nil, fmt.Errorf("tool handler is not configured")
		}

		args := rawArgs
		if args == nil {
			args = map[string]any{}
		}

		// Validate the raw arguments against the tool's input schema, mirroring the
		// SDK's direct-call path: unmarshal to a map, apply defaults, then validate.
		if resolved != nil {
			var v any = args
			if err := resolved.ApplyDefaults(&v); err != nil {
				return nil, nil, fmt.Errorf("applying argument defaults: %w", err)
			}
			if err := resolved.Validate(&v); err != nil {
				return nil, nil, fmt.Errorf("invalid arguments: %w", err)
			}
			if m, ok := v.(map[string]any); ok {
				args = m
			}
		}

		var in In
		if len(args) > 0 {
			data, err := json.Marshal(args)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to encode tool arguments: %w", err)
			}
			if err := json.Unmarshal(data, &in); err != nil {
				return nil, nil, fmt.Errorf("failed to decode arguments into target parameter schema: %w", err)
			}
		}
		res, out, err := handler(ctx, &mcp.CallToolRequest{}, in)
		if err != nil {
			return nil, nil, err
		}
		return res, out, nil
	}
}

// buildToolDispatcher builds the internal dispatcher mapping tool names to their execution runners.
func buildToolDispatcher(client *autotask.Client, mapper *services.MappingCache, picklist *services.PicklistCache) map[string]ToolRunner {
	return map[string]ToolRunner{
		// Connection
		"autotask_test_connection": makeRunner(testConnectionHandler(client)),

		// Picklists and metadata
		"autotask_list_queues":            makeRunner(listQueuesHandler(picklist)),
		"autotask_list_ticket_statuses":   makeRunner(listTicketStatusesHandler(picklist)),
		"autotask_list_ticket_priorities": makeRunner(listTicketPrioritiesHandler(picklist)),
		"autotask_get_field_info":         makeRunner(getFieldInfoHandler(picklist)),

		// Tickets
		"autotask_search_tickets":     makeRunner(searchTicketsHandler(client, mapper)),
		"autotask_get_ticket_details": makeRunner(getTicketDetailsHandler(client, mapper)),
		"autotask_create_ticket":      makeRunner(createTicketHandler(client)),
		"autotask_update_ticket":      makeRunner(updateTicketHandler(client)),

		// Companies
		"autotask_search_companies": makeRunner(searchCompaniesHandler(client, mapper)),
		"autotask_create_company":   makeRunner(createCompanyHandler(client)),
		"autotask_update_company":   makeRunner(updateCompanyHandler(client)),

		// Contacts
		"autotask_search_contacts": makeRunner(searchContactsHandler(client, mapper)),
		"autotask_create_contact":  makeRunner(createContactHandler(client)),

		// Projects
		"autotask_search_projects": makeRunner(searchProjectsHandler(client, mapper)),
		"autotask_create_project":  makeRunner(createProjectHandler(client)),

		// Tasks
		"autotask_search_tasks": makeRunner(searchTasksHandler(client, mapper)),
		"autotask_create_task":  makeRunner(createTaskHandler(client)),

		// Time entries
		"autotask_search_time_entries": makeRunner(searchTimeEntriesHandler(client, mapper)),
		"autotask_create_time_entry":   makeRunner(createTimeEntryHandler(client)),

		// Resources
		"autotask_search_resources": makeRunner(searchResourcesHandler(client)),

		// Configuration items
		"autotask_search_configuration_items": makeRunner(searchConfigurationItemsHandler(client, mapper)),

		// Company, ticket, and project notes
		"autotask_get_ticket_note":      makeRunner(getTicketNoteHandler(client)),
		"autotask_search_ticket_notes":  makeRunner(searchTicketNotesHandler(client)),
		"autotask_create_ticket_note":   makeRunner(createTicketNoteHandler(client)),
		"autotask_get_project_note":     makeRunner(getProjectNoteHandler(client)),
		"autotask_search_project_notes": makeRunner(searchProjectNotesHandler(client)),
		"autotask_create_project_note":  makeRunner(createProjectNoteHandler(client)),
		"autotask_get_company_note":     makeRunner(getCompanyNoteHandler(client)),
		"autotask_search_company_notes": makeRunner(searchCompanyNotesHandler(client)),
		"autotask_create_company_note":  makeRunner(createCompanyNoteHandler(client)),

		// Ticket attachments
		"autotask_get_ticket_attachment":     makeRunner(getTicketAttachmentHandler(client)),
		"autotask_search_ticket_attachments": makeRunner(searchTicketAttachmentsHandler(client)),

		// Billing
		"autotask_get_billing_item":                    makeRunner(getBillingItemHandler(client)),
		"autotask_search_billing_items":                makeRunner(searchBillingItemsHandler(client, mapper)),
		"autotask_search_billing_item_approval_levels": makeRunner(searchBillingItemApprovalLevelsHandler(client)),

		// Expenses
		"autotask_get_expense_report":     makeRunner(getExpenseReportHandler(client)),
		"autotask_search_expense_reports": makeRunner(searchExpenseReportsHandler(client)),
		"autotask_create_expense_report":  makeRunner(createExpenseReportHandler(client)),
		"autotask_create_expense_item":    makeRunner(createExpenseItemHandler(client)),

		// Sales
		"autotask_get_product":            makeRunner(getProductHandler(client)),
		"autotask_search_products":        makeRunner(searchProductsHandler(client)),
		"autotask_get_service":            makeRunner(getServiceHandler(client)),
		"autotask_search_services":        makeRunner(searchServicesHandler(client)),
		"autotask_get_service_bundle":     makeRunner(getServiceBundleHandler(client)),
		"autotask_search_service_bundles": makeRunner(searchServiceBundlesHandler(client)),

		// Financial
		"autotask_get_quote":            makeRunner(getQuoteHandler(client)),
		"autotask_search_quotes":        makeRunner(searchQuotesHandler(client)),
		"autotask_create_quote":         makeRunner(createQuoteHandler(client)),
		"autotask_get_quote_item":       makeRunner(getQuoteItemHandler(client)),
		"autotask_search_quote_items":   makeRunner(searchQuoteItemsHandler(client)),
		"autotask_create_quote_item":    makeRunner(createQuoteItemHandler(client)),
		"autotask_update_quote_item":    makeRunner(updateQuoteItemHandler(client)),
		"autotask_delete_quote_item":    makeRunner(deleteQuoteItemHandler(client)),
		"autotask_get_opportunity":      makeRunner(getOpportunityHandler(client)),
		"autotask_search_opportunities": makeRunner(searchOpportunitiesHandler(client)),
		"autotask_create_opportunity":   makeRunner(createOpportunityHandler(client)),
		"autotask_search_contracts":     makeRunner(searchContractsHandler(client, mapper)),
		"autotask_search_invoices":      makeRunner(searchInvoicesHandler(client)),
	}
}

// precomputedCategorySummaries is computed once at package initialization.
var precomputedCategorySummaries map[string]CategorySummary

func init() {
	precomputedCategorySummaries = make(map[string]CategorySummary, len(ToolCategories))
	for name, info := range ToolCategories {
		precomputedCategorySummaries[name] = CategorySummary{
			Description: info.Description,
			ToolCount:   len(info.Tools),
			Tools:       info.Tools,
		}
	}
}

// RegisterLazyTools registers the 4 lazy-loading meta-tools with the server.
func RegisterLazyTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache, picklist *services.PicklistCache) {
	dispatcher := buildToolDispatcher(client, mapper, picklist)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_categories",
		Description: "Enumerate the Autotask tool categories (tickets, companies, projects, financial, and more), each with its description, member tool count, and tool names. Start here in lazy-loading mode to discover which domains exist, then call autotask_list_category_tools to see the tools within one category. Reads the static in-process registry with no Autotask API call. Read-only.",
		Annotations: localReadTool("List categories"),
	}, listCategoriesHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_list_category_tools",
		Description: "List the tool names and one-line descriptions belonging to one category, matched case-insensitively by category name. Call autotask_list_categories first to obtain valid category names; to invoke a listed tool use autotask_execute_tool or call it directly through the MCP client. Requires category and reads the static in-process registry with no Autotask API call. Read-only.",
		Annotations: localReadTool("List category tools"),
	}, listCategoryToolsHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_execute_tool",
		Description: "Execute a named Autotask tool with arguments in lazy-loading mode. Dispatches the tool call to the internal handler and returns the execution result. Use autotask_router or autotask_list_category_tools to discover tool names and argument schemas. Requires toolName. Open world.",
		// This proxy can dispatch create/update/delete tools, so it must advertise
		// itself as destructive: a host that gates confirmation on DestructiveHint
		// would otherwise grant blanket mutation access after one read-only call.
		Annotations: &mcp.ToolAnnotations{Title: "Execute tool", DestructiveHint: new(true), OpenWorldHint: new(true)},
	}, executeToolHandler(dispatcher))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_router",
		Description: "Map a natural-language intent to the single best-matching Autotask tool by keyword, returning the suggested tool name and its description. Use this when you know what you want to do but not the tool name; for a structured browse by domain use autotask_list_categories and autotask_list_category_tools instead. Requires intent and falls back to autotask_list_categories when nothing matches. Read-only.",
		Annotations: localReadTool("Route intent"),
	}, routerHandler())
}

// listCategoriesHandler returns a handler that returns the full category map.
func listCategoriesHandler() func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoriesInput) (*mcp.CallToolResult, map[string]CategorySummary, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoriesInput) (*mcp.CallToolResult, map[string]CategorySummary, error) {
		return nil, precomputedCategorySummaries, nil
	}
}

// listCategoryToolsHandler returns a handler that lists tools for a given category.
func listCategoryToolsHandler() func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoryToolsInput) (*mcp.CallToolResult, CategoryToolsOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListCategoryToolsInput) (*mcp.CallToolResult, CategoryToolsOut, error) {
		categoryName := in.Category
		cat, ok := ToolCategories[in.Category]
		if !ok {
			// Try case-insensitive match.
			lower := strings.ToLower(in.Category)
			for k, v := range ToolCategories {
				if strings.ToLower(k) == lower {
					cat = v
					categoryName = k
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, CategoryToolsOut{}, fmt.Errorf("unknown category %q; call autotask_list_categories to see available categories", in.Category)
		}

		tools := make([]ToolSummary, 0, len(cat.Tools))
		for _, name := range cat.Tools {
			desc := toolDescriptions[name]
			if desc == "" {
				desc = name
			}
			tools = append(tools, ToolSummary{Name: name, Description: desc})
		}

		return nil, CategoryToolsOut{
			Category:    categoryName,
			Description: cat.Description,
			Tools:       tools,
		}, nil
	}
}

// executeToolHandler returns a handler that dynamically dispatches named tools.
func executeToolHandler(dispatcher map[string]ToolRunner) func(ctx context.Context, req *mcp.CallToolRequest, in ExecuteToolInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ExecuteToolInput) (*mcp.CallToolResult, any, error) {
		if in.ToolName == "" {
			return nil, nil, fmt.Errorf("toolName is required")
		}

		runner, ok := dispatcher[in.ToolName]
		if !ok || runner == nil {
			return nil, nil, fmt.Errorf("unknown tool %q; call autotask_list_categories or autotask_router to discover available tools", in.ToolName)
		}

		return runner(ctx, in.Arguments)
	}
}

// routerHandler returns a handler that routes a natural language intent to a tool.
func routerHandler() func(ctx context.Context, req *mcp.CallToolRequest, in RouterInput) (*mcp.CallToolResult, RouterOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in RouterInput) (*mcp.CallToolResult, RouterOut, error) {
		if in.Intent == "" {
			return nil, RouterOut{}, fmt.Errorf("intent is required")
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

		return nil, RouterOut{
			Intent:        in.Intent,
			SuggestedTool: suggestedTool,
			Description:   description,
		}, nil
	}
}
