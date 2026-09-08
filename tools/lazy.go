package tools

import (
	"context"
	"encoding/json"
	"errors"
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
		Tools:       []string{toolTestConnection, toolListQueues, toolListTicketStatuses, toolListTicketPriorities, toolGetFieldInfo},
	},
	"companies": {
		Description: "Search, create, and update companies",
		Tools:       []string{toolSearchCompanies, toolCreateCompany, toolUpdateCompany},
	},
	"contacts": {
		Description: "Search and create contacts",
		Tools:       []string{toolSearchContacts, toolCreateContact},
	},
	"tickets": {
		Description: "Search, create, update tickets and manage notes/attachments",
		Tools:       []string{toolSearchTickets, toolGetTicketDetails, toolCreateTicket, toolUpdateTicket, toolGetTicketNote, toolSearchTicketNotes, toolCreateTicketNote, toolGetTicketAttachment, toolSearchTicketAttachments},
	},
	"projects": {
		Description: "Search and create projects, tasks, and project notes",
		Tools:       []string{toolSearchProjects, toolCreateProject, toolSearchTasks, toolCreateTask, toolGetProjectNote, toolSearchProjectNotes, toolCreateProjectNote},
	},
	"time_and_billing": {
		Description: "Time entries, billing items, and expense management",
		Tools:       []string{toolCreateTimeEntry, toolSearchTimeEntries, toolSearchBillingItems, toolGetBillingItem, toolSearchBillingItemApprovalLevels, toolGetExpenseReport, toolSearchExpenseReports, toolCreateExpenseReport, toolCreateExpenseItem},
	},
	"financial": {
		Description: "Quotes, quote items, opportunities, invoices, and contracts",
		Tools:       []string{toolGetQuote, toolSearchQuotes, toolCreateQuote, toolGetQuoteItem, toolSearchQuoteItems, toolCreateQuoteItem, toolUpdateQuoteItem, toolDeleteQuoteItem, toolGetOpportunity, toolSearchOpportunities, toolCreateOpportunity, toolSearchInvoices, toolSearchContracts},
	},
	"products_and_services": {
		Description: "Products, services, and service bundles catalog",
		Tools:       []string{toolGetProduct, toolSearchProducts, toolGetService, toolSearchServices, toolGetServiceBundle, toolSearchServiceBundles},
	},
	"resources": {
		Description: "Search for Autotask resources",
		Tools:       []string{toolSearchResources},
	},
	"configuration_items": {
		Description: "Search configuration items",
		Tools:       []string{toolSearchConfigurationItems},
	},
	"company_notes": {
		Description: "Get, search, and create company notes",
		Tools:       []string{toolGetCompanyNote, toolSearchCompanyNotes, toolCreateCompanyNote},
	},
}

// toolDescriptions maps tool names to their human-readable descriptions.
// Used by autotask_list_category_tools.
var toolDescriptions = map[string]string{
	toolTestConnection:                  "Test connectivity to the Autotask API",
	toolListQueues:                      "List available ticket queues",
	toolListTicketStatuses:              "List available ticket status values",
	toolListTicketPriorities:            "List available ticket priority values",
	toolGetFieldInfo:                    "Get field metadata for an entity type",
	toolSearchCompanies:                 "Search for companies",
	toolCreateCompany:                   "Create a new company",
	toolUpdateCompany:                   "Update an existing company",
	toolSearchContacts:                  "Search for contacts",
	toolCreateContact:                   "Create a new contact",
	toolSearchTickets:                   "Search for tickets",
	toolGetTicketDetails:                "Get detailed information for a ticket",
	toolCreateTicket:                    "Create a new ticket",
	toolUpdateTicket:                    "Update an existing ticket",
	toolGetTicketNote:                   "Get a specific ticket note by ID",
	toolSearchTicketNotes:               "Search ticket notes",
	toolCreateTicketNote:                "Create a new note on a ticket",
	toolGetTicketAttachment:             "Get a ticket attachment by ID",
	toolSearchTicketAttachments:         "Search ticket attachments",
	toolSearchProjects:                  "Search for projects",
	toolCreateProject:                   "Create a new project",
	toolSearchTasks:                     "Search for project tasks",
	toolCreateTask:                      "Create a new project task",
	toolGetProjectNote:                  "Get a project note by ID",
	toolSearchProjectNotes:              "Search project notes",
	toolCreateProjectNote:               "Create a new project note",
	toolCreateTimeEntry:                 "Create a new time entry",
	toolSearchTimeEntries:               "Search time entries",
	toolSearchBillingItems:              "Search billing items",
	toolGetBillingItem:                  "Get a billing item by ID",
	toolSearchBillingItemApprovalLevels: "List billing item approval levels",
	toolGetExpenseReport:                "Get an expense report by ID",
	toolSearchExpenseReports:            "Search expense reports",
	toolCreateExpenseReport:             "Create a new expense report",
	toolCreateExpenseItem:               "Create a new expense item",
	toolGetQuote:                        "Get a quote by ID",
	toolSearchQuotes:                    "Search quotes",
	toolCreateQuote:                     "Create a new quote",
	toolGetQuoteItem:                    "Get a quote item by ID",
	toolSearchQuoteItems:                "Search quote items",
	toolCreateQuoteItem:                 "Create a new quote item",
	toolUpdateQuoteItem:                 "Update an existing quote item",
	toolDeleteQuoteItem:                 "Delete a quote item",
	toolGetOpportunity:                  "Get an opportunity by ID",
	toolSearchOpportunities:             "Search opportunities",
	toolCreateOpportunity:               "Create a new opportunity",
	toolSearchInvoices:                  "Search invoices",
	toolSearchContracts:                 "Search contracts",
	toolGetProduct:                      "Get a product by ID",
	toolSearchProducts:                  "Search products",
	toolGetService:                      "Get a service by ID",
	toolSearchServices:                  "Search services",
	toolGetServiceBundle:                "Get a service bundle by ID",
	toolSearchServiceBundles:            "Search service bundles",
	toolSearchResources:                 "Search for Autotask resources (employees/contacts)",
	toolSearchConfigurationItems:        descSearchConfigurationItems,
	toolGetCompanyNote:                  "Get a company note by ID",
	toolSearchCompanyNotes:              "Search company notes",
	toolCreateCompanyNote:               "Create a new company note",
}

// routingRules maps keywords to suggested tools for autotask_router.
var routingRules = []struct {
	keywords    []string
	tool        string
	description string
}{
	{[]string{"create ticket", "new ticket", "open ticket"}, toolCreateTicket, "Create a new ticket"},
	{[]string{"ticket", "issue", "problem", "request"}, toolSearchTickets, "Search for tickets"},
	{[]string{"company", "companies", "account", "client", "customer"}, toolSearchCompanies, "Search for companies"},
	{[]string{"contact", "contacts", "person", "user"}, toolSearchContacts, "Search for contacts"},
	{[]string{"time", "hours", "timesheet"}, toolSearchTimeEntries, "Search time entries"},
	{[]string{"project"}, toolSearchProjects, "Search for projects"},
	{[]string{"task"}, toolSearchTasks, "Search for tasks"},
	{[]string{"billing", "invoice", "charge"}, toolSearchBillingItems, "Search billing items"},
	{[]string{"expense", "cost"}, toolSearchExpenseReports, "Search expense reports"},
	{[]string{"quote", "proposal"}, toolSearchQuotes, "Search quotes"},
	{[]string{"opportunity", "deal", "sale"}, toolSearchOpportunities, "Search opportunities"},
	{[]string{"contract"}, toolSearchContracts, "Search contracts"},
	{[]string{"product", "item", "sku"}, toolSearchProducts, "Search products"},
	{[]string{"service"}, toolSearchServices, "Search services"},
	{[]string{"resource", "employee", "tech"}, toolSearchResources, "Search resources"},
	{[]string{"config", "configuration", "device", "asset"}, toolSearchConfigurationItems, descSearchConfigurationItems},
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
			return nil, nil, errors.New("tool handler is not configured")
		}

		args := rawArgs
		if args == nil {
			args = map[string]any{}
		}

		validatedArgs, err := validateRunnerArgs(resolved, args)
		if err != nil {
			return nil, nil, err
		}

		in, err := unmarshalRunnerArgs[In](validatedArgs)
		if err != nil {
			return nil, nil, err
		}

		res, out, err := handler(ctx, &mcp.CallToolRequest{}, in)
		if err != nil {
			return nil, nil, err
		}
		return res, out, nil
	}
}

func validateRunnerArgs(resolved *jsonschema.Resolved, args map[string]any) (map[string]any, error) {
	if resolved == nil {
		return args, nil
	}
	var v any = args
	if err := resolved.ApplyDefaults(&v); err != nil {
		return nil, fmt.Errorf("applying argument defaults: %w", err)
	}
	if err := resolved.Validate(&v); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if m, ok := v.(map[string]any); ok {
		return m, nil
	}
	return args, nil
}

func unmarshalRunnerArgs[In any](args map[string]any) (In, error) {
	var in In
	if len(args) == 0 {
		return in, nil
	}
	data, err := json.Marshal(args)
	if err != nil {
		return in, fmt.Errorf("failed to encode tool arguments: %w", err)
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return in, fmt.Errorf("failed to decode arguments into target parameter schema: %w", err)
	}
	return in, nil
}

// buildToolDispatcher builds the internal dispatcher mapping tool names to their execution runners.
func buildToolDispatcher(client *autotask.Client, mapper *services.MappingCache, picklist *services.PicklistCache) map[string]ToolRunner {
	return map[string]ToolRunner{
		// Connection
		toolTestConnection: makeRunner(testConnectionHandler(client)),

		// Picklists and metadata
		toolListQueues:           makeRunner(listQueuesHandler(picklist)),
		toolListTicketStatuses:   makeRunner(listTicketStatusesHandler(picklist)),
		toolListTicketPriorities: makeRunner(listTicketPrioritiesHandler(picklist)),
		toolGetFieldInfo:         makeRunner(getFieldInfoHandler(picklist)),

		// Tickets
		toolSearchTickets:    makeRunner(searchTicketsHandler(client, mapper)),
		toolGetTicketDetails: makeRunner(getTicketDetailsHandler(client, mapper)),
		toolCreateTicket:     makeRunner(createTicketHandler(client)),
		toolUpdateTicket:     makeRunner(updateTicketHandler(client)),

		// Companies
		toolSearchCompanies: makeRunner(searchCompaniesHandler(client, mapper)),
		toolCreateCompany:   makeRunner(createCompanyHandler(client)),
		toolUpdateCompany:   makeRunner(updateCompanyHandler(client)),

		// Contacts
		toolSearchContacts: makeRunner(searchContactsHandler(client, mapper)),
		toolCreateContact:  makeRunner(createContactHandler(client)),

		// Projects
		toolSearchProjects: makeRunner(searchProjectsHandler(client, mapper)),
		toolCreateProject:  makeRunner(createProjectHandler(client)),

		// Tasks
		toolSearchTasks: makeRunner(searchTasksHandler(client, mapper)),
		toolCreateTask:  makeRunner(createTaskHandler(client)),

		// Time entries
		toolSearchTimeEntries: makeRunner(searchTimeEntriesHandler(client, mapper)),
		toolCreateTimeEntry:   makeRunner(createTimeEntryHandler(client)),

		// Resources
		toolSearchResources: makeRunner(searchResourcesHandler(client)),

		// Configuration items
		toolSearchConfigurationItems: makeRunner(searchConfigurationItemsHandler(client, mapper)),

		// Company, ticket, and project notes
		toolGetTicketNote:      makeRunner(getTicketNoteHandler(client)),
		toolSearchTicketNotes:  makeRunner(searchTicketNotesHandler(client)),
		toolCreateTicketNote:   makeRunner(createTicketNoteHandler(client)),
		toolGetProjectNote:     makeRunner(getProjectNoteHandler(client)),
		toolSearchProjectNotes: makeRunner(searchProjectNotesHandler(client)),
		toolCreateProjectNote:  makeRunner(createProjectNoteHandler(client)),
		toolGetCompanyNote:     makeRunner(getCompanyNoteHandler(client)),
		toolSearchCompanyNotes: makeRunner(searchCompanyNotesHandler(client)),
		toolCreateCompanyNote:  makeRunner(createCompanyNoteHandler(client)),

		// Ticket attachments
		toolGetTicketAttachment:     makeRunner(getTicketAttachmentHandler(client)),
		toolSearchTicketAttachments: makeRunner(searchTicketAttachmentsHandler(client)),

		// Billing
		toolGetBillingItem:                  makeRunner(getBillingItemHandler(client)),
		toolSearchBillingItems:              makeRunner(searchBillingItemsHandler(client, mapper)),
		toolSearchBillingItemApprovalLevels: makeRunner(searchBillingItemApprovalLevelsHandler(client)),

		// Expenses
		toolGetExpenseReport:     makeRunner(getExpenseReportHandler(client)),
		toolSearchExpenseReports: makeRunner(searchExpenseReportsHandler(client)),
		toolCreateExpenseReport:  makeRunner(createExpenseReportHandler(client)),
		toolCreateExpenseItem:    makeRunner(createExpenseItemHandler(client)),

		// Sales
		toolGetProduct:           makeRunner(getProductHandler(client)),
		toolSearchProducts:       makeRunner(searchProductsHandler(client)),
		toolGetService:           makeRunner(getServiceHandler(client)),
		toolSearchServices:       makeRunner(searchServicesHandler(client)),
		toolGetServiceBundle:     makeRunner(getServiceBundleHandler(client)),
		toolSearchServiceBundles: makeRunner(searchServiceBundlesHandler(client)),

		// Financial
		toolGetQuote:            makeRunner(getQuoteHandler(client)),
		toolSearchQuotes:        makeRunner(searchQuotesHandler(client)),
		toolCreateQuote:         makeRunner(createQuoteHandler(client)),
		toolGetQuoteItem:        makeRunner(getQuoteItemHandler(client)),
		toolSearchQuoteItems:    makeRunner(searchQuoteItemsHandler(client)),
		toolCreateQuoteItem:     makeRunner(createQuoteItemHandler(client)),
		toolUpdateQuoteItem:     makeRunner(updateQuoteItemHandler(client)),
		toolDeleteQuoteItem:     makeRunner(deleteQuoteItemHandler(client)),
		toolGetOpportunity:      makeRunner(getOpportunityHandler(client)),
		toolSearchOpportunities: makeRunner(searchOpportunitiesHandler(client)),
		toolCreateOpportunity:   makeRunner(createOpportunityHandler(client)),
		toolSearchContracts:     makeRunner(searchContractsHandler(client, mapper)),
		toolSearchInvoices:      makeRunner(searchInvoicesHandler(client)),
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

const (
	toolListCategories    = "autotask_list_categories"
	toolListCategoryTools = "autotask_list_category_tools"
	toolExecuteTool       = "autotask_execute_tool"
	toolRouter            = "autotask_router"
)

// RegisterLazyTools registers the 4 lazy-loading meta-tools with the server.
func RegisterLazyTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache, picklist *services.PicklistCache) {
	dispatcher := buildToolDispatcher(client, mapper, picklist)

	mcp.AddTool(s, &mcp.Tool{
		Name:        toolListCategories,
		Description: "Enumerate the Autotask tool categories (tickets, companies, projects, financial, and more), each with its description, member tool count, and tool names. Start here in lazy-loading mode to discover which domains exist, then call autotask_list_category_tools to see the tools within one category. Reads the static in-process registry with no Autotask API call. Read-only.",
		Annotations: localReadTool("List categories"),
	}, listCategoriesHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        toolListCategoryTools,
		Description: "List the tool names and one-line descriptions belonging to one category, matched case-insensitively by category name. Call autotask_list_categories first to obtain valid category names; to invoke a listed tool use autotask_execute_tool or call it directly through the MCP client. Requires category and reads the static in-process registry with no Autotask API call. Read-only.",
		Annotations: localReadTool("List category tools"),
	}, listCategoryToolsHandler())

	mcp.AddTool(s, &mcp.Tool{
		Name:        toolExecuteTool,
		Description: "Execute a named Autotask tool with arguments in lazy-loading mode. Dispatches the tool call to the internal handler and returns the execution result. Use autotask_router or autotask_list_category_tools to discover tool names and argument schemas. Requires toolName. Open world.",
		// This proxy can dispatch create/update/delete tools, so it must advertise
		// itself as destructive: a host that gates confirmation on DestructiveHint
		// would otherwise grant blanket mutation access after one read-only call.
		Annotations: &mcp.ToolAnnotations{Title: "Execute tool", DestructiveHint: new(true), OpenWorldHint: new(true)},
	}, executeToolHandler(dispatcher))

	mcp.AddTool(s, &mcp.Tool{
		Name:        toolRouter,
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
				if strings.EqualFold(k, lower) {
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
			return nil, nil, errors.New("toolName is required")
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
			return nil, RouterOut{}, errors.New("intent is required")
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
			suggestedTool = toolListCategories
			description = "No specific match found. Use autotask_list_categories to browse available tools."
		}

		return nil, RouterOut{
			Intent:        in.Intent,
			SuggestedTool: suggestedTool,
			Description:   description,
		}, nil
	}
}
