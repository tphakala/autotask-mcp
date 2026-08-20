package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchContactsInput defines the input parameters for searching contacts.
type SearchContactsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search term for contact name or email"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	IsActive   *int   `json:"isActive,omitempty" jsonschema:"Filter by active status (1=active, 0=inactive)"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 200)"`
}

// CreateContactInput defines the input parameters for creating a new contact.
type CreateContactInput struct {
	CompanyID    int64  `json:"companyID" jsonschema:"Company ID for the contact"`
	FirstName    string `json:"firstName" jsonschema:"Contact first name"`
	LastName     string `json:"lastName" jsonschema:"Contact last name"`
	EmailAddress string `json:"emailAddress,omitempty" jsonschema:"Contact email address"`
	Phone        string `json:"phone,omitempty" jsonschema:"Contact phone number"`
	Title        string `json:"title,omitempty" jsonschema:"Contact job title"`
}

// RegisterContactTools registers all contact-related MCP tools with the server.
func RegisterContactTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_contacts",
		Description: "Find contacts by name or email substring, company ID, or active status, returning a compact summary of matching records (up to maxResults, default 25, max 200). A searchTerm matches across first name, last name, and email address. Use this to locate a contact and its ID; to add a new one instead use autotask_create_contact. Read-only.",
		Annotations: readOnlyTool("Search contacts"),
	}, searchContactsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_contact",
		Description: "Add a new contact under a company from a first and last name, with optional email, phone, and job title. Requires companyID, firstName, and lastName, and returns the created contact including its new ID; look up the companyID with autotask_search_companies. To find existing contacts instead use autotask_search_contacts. Writes to Autotask.",
		Annotations: createTool("Create contact"),
	}, createContactHandler(client))
}

// searchContactsHandler returns a handler that searches contacts using the provided filters.
func searchContactsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchContactsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchContactsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 200)

		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Or(
				autotask.Field("firstName", autotask.OpContains, in.SearchTerm),
				autotask.Field("lastName", autotask.OpContains, in.SearchTerm),
				autotask.Field("emailAddress", autotask.OpContains, in.SearchTerm),
			)
		}
		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		contacts, err := autotask.List[entities.Contact](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(contacts) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(contacts)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, mapper, maps, "autotask_search_contacts", maxResults)
	}
}

// createContactHandler returns a handler that creates a new contact.
func createContactHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateContactInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateContactInput) (*mcp.CallToolResult, map[string]any, error) {
		contact := &entities.Contact{
			CompanyID: autotask.Set(in.CompanyID),
			FirstName: autotask.Set(in.FirstName),
			LastName:  autotask.Set(in.LastName),
		}

		if in.EmailAddress != "" {
			contact.EmailAddress = autotask.Set(in.EmailAddress)
		}
		if in.Phone != "" {
			contact.Phone = autotask.Set(in.Phone)
		}
		if in.Title != "" {
			contact.Title = autotask.Set(in.Title)
		}

		created, err := autotask.Create[entities.Contact](ctx, client, contact)
		if err != nil {
			return nil, nil, err
		}

		m, err := entityToMap(created)
		if err != nil {
			return nil, nil, err
		}

		return nil, m, nil
	}
}
