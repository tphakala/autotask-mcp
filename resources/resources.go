package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

const (
	mimeTypeJSON        = "application/json"
	defaultResourceLimit = 50
	statusCompleted      = 5
)

// RegisterAll registers all MCP resource endpoints with the server.
func RegisterAll(s *mcp.Server, client *autotask.Client) {
	// Static list resources.
	s.AddResource(&mcp.Resource{
		URI:         "autotask://companies",
		Name:        "companies",
		Description: "List companies in Autotask (up to 50)",
		MIMEType:    mimeTypeJSON,
	}, listCompaniesHandler(client))

	s.AddResource(&mcp.Resource{
		URI:         "autotask://contacts",
		Name:        "contacts",
		Description: "List contacts in Autotask (up to 50)",
		MIMEType:    mimeTypeJSON,
	}, listContactsHandler(client))

	s.AddResource(&mcp.Resource{
		URI:         "autotask://tickets",
		Name:        "tickets",
		Description: "List open tickets in Autotask (up to 50)",
		MIMEType:    mimeTypeJSON,
	}, listTicketsHandler(client))

	s.AddResource(&mcp.Resource{
		URI:         "autotask://time-entries",
		Name:        "time-entries",
		Description: "List time entries in Autotask (up to 50)",
		MIMEType:    mimeTypeJSON,
	}, listTimeEntriesHandler(client))

	// URI template resources (by ID).
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "autotask://companies/{id}",
		Name:        "company",
		Description: "Get a company by ID",
		MIMEType:    mimeTypeJSON,
	}, getCompanyHandler(client))

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "autotask://contacts/{id}",
		Name:        "contact",
		Description: "Get a contact by ID",
		MIMEType:    mimeTypeJSON,
	}, getContactHandler(client))

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "autotask://tickets/{id}",
		Name:        "ticket",
		Description: "Get a ticket by ID",
		MIMEType:    mimeTypeJSON,
	}, getTicketHandler(client))
}

// parseIDFromURI extracts a numeric ID from the last path segment of a URI.
// For example, "autotask://companies/123" returns 123.
func parseIDFromURI(uri string) (int64, error) {
	idx := strings.LastIndex(uri, "/")
	if idx < 0 || idx == len(uri)-1 {
		return 0, fmt.Errorf("no ID found in URI %q", uri)
	}
	return strconv.ParseInt(uri[idx+1:], 10, 64)
}

// jsonResult wraps any value as a JSON ResourceContents.
func jsonResult(uri string, v any) (*mcp.ReadResourceResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: mimeTypeJSON,
			Text:     string(data),
		}},
	}, nil
}

// listCompaniesHandler returns a ResourceHandler that lists all companies.
func listCompaniesHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		q := autotask.NewQuery().Limit(defaultResourceLimit)
		companies, err := autotask.List[entities.Company](ctx, client, q)
		if err != nil {
			return nil, fmt.Errorf("list companies: %w", err)
		}
		return jsonResult(req.Params.URI, companies)
	}
}

// getCompanyHandler returns a ResourceHandler that fetches a single company by ID.
func getCompanyHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id, err := parseIDFromURI(req.Params.URI)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		company, err := autotask.Get[entities.Company](ctx, client, id)
		if err != nil {
			return nil, fmt.Errorf("get company %d: %w", id, err)
		}
		return jsonResult(req.Params.URI, company)
	}
}

// listContactsHandler returns a ResourceHandler that lists all contacts.
func listContactsHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		q := autotask.NewQuery().Limit(defaultResourceLimit)
		contacts, err := autotask.List[entities.Contact](ctx, client, q)
		if err != nil {
			return nil, fmt.Errorf("list contacts: %w", err)
		}
		return jsonResult(req.Params.URI, contacts)
	}
}

// getContactHandler returns a ResourceHandler that fetches a single contact by ID.
func getContactHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id, err := parseIDFromURI(req.Params.URI)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		contact, err := autotask.Get[entities.Contact](ctx, client, id)
		if err != nil {
			return nil, fmt.Errorf("get contact %d: %w", id, err)
		}
		return jsonResult(req.Params.URI, contact)
	}
}

// listTicketsHandler returns a ResourceHandler that lists open tickets.
func listTicketsHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		q := autotask.NewQuery().Limit(defaultResourceLimit)
		q.Where("status", autotask.OpNotEq, statusCompleted)
		tickets, err := autotask.List[entities.Ticket](ctx, client, q)
		if err != nil {
			return nil, fmt.Errorf("list tickets: %w", err)
		}
		return jsonResult(req.Params.URI, tickets)
	}
}

// getTicketHandler returns a ResourceHandler that fetches a single ticket by ID.
func getTicketHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id, err := parseIDFromURI(req.Params.URI)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		ticket, err := autotask.Get[entities.Ticket](ctx, client, id)
		if err != nil {
			return nil, fmt.Errorf("get ticket %d: %w", id, err)
		}
		return jsonResult(req.Params.URI, ticket)
	}
}

// listTimeEntriesHandler returns a ResourceHandler that lists time entries.
func listTimeEntriesHandler(client *autotask.Client) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		q := autotask.NewQuery().Limit(defaultResourceLimit)
		timeEntries, err := autotask.List[entities.TimeEntry](ctx, client, q)
		if err != nil {
			return nil, fmt.Errorf("list time entries: %w", err)
		}
		return jsonResult(req.Params.URI, timeEntries)
	}
}
