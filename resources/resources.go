package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

const (
	mimeTypeJSON         = "application/json"
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
		Description: "List non-completed tickets in Autotask (up to 50)",
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
	id, err := strconv.ParseInt(uri[idx+1:], 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, fmt.Errorf("invalid non-positive ID %d in URI %q", id, uri)
	}
	return id, nil
}

// jsonResult wraps any value as a JSON ResourceContents. Entity data is routed
// through the same untrusted-content framing the tools apply (see frameUntrusted),
// so customer-controlled free text read via a resource URI is bounded consistently.
func jsonResult(uri string, v any) (*mcp.ReadResourceResult, error) {
	framed, err := frameUntrusted(v)
	if err != nil {
		return nil, fmt.Errorf("frame: %w", err)
	}
	// Encode with HTML escaping OFF so the untrusted-content boundary markers appear
	// as literal "<untrusted_content>" rather than "<...". This matches the tool
	// path (the SDK transport encodes tool output with SetEscapeHTML(false)); the
	// default json.MarshalIndent would escape the markers and blunt the defense.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(framed); err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: mimeTypeJSON,
			// Encoder.Encode appends a trailing newline; trim it to match the prior output.
			Text: strings.TrimRight(buf.String(), "\n"),
		}},
	}, nil
}

// frameUntrusted converts a typed entity (or slice of entities) to its generic
// JSON form and wraps customer-controlled free-text fields in untrusted-content
// boundary markers. It mirrors the tools path, which marshals the same entities
// to a map and calls services.FrameUntrustedMapFields, so resources and tools
// present untrusted content identically.
func frameUntrusted(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("frame marshal: %w", err)
	}
	// UseNumber keeps entity IDs and other integers as json.Number rather than
	// float64, so large int64 IDs re-marshal exactly instead of losing precision
	// or printing in scientific notation.
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var decoded any
	if err := dec.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("frame unmarshal: %w", err)
	}
	frameValue(decoded)
	return decoded, nil
}

// frameValue applies field framing to each entity map reachable from v. A list
// resource decodes to []any of maps; a get resource decodes to a single map.
// Framing matches the tools path: top-level known fields only, no deeper
// recursion, so the two surfaces stay consistent.
func frameValue(v any) {
	switch t := v.(type) {
	case map[string]any:
		services.FrameUntrustedMapFields(t)
	case []any:
		for _, item := range t {
			frameValue(item)
		}
	}
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
