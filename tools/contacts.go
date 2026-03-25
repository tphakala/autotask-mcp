package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// SearchContactsInput defines the input parameters for searching contacts.
type SearchContactsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search term for contact name or email"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	IsActive   *int   `json:"isActive,omitempty" jsonschema:"Filter by active status (1=active, 0=inactive)"`
	Page       int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 200)"`
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
		Description: "Search for contacts in Autotask. Returns 25 results per page by default.",
	}, searchContactsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_contact",
		Description: "Create a new contact in Autotask.",
	}, createContactHandler(client))
}

// searchContactsHandler returns a handler that searches contacts using the provided filters.
func searchContactsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchContactsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchContactsInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 200)

		q := autotask.NewQuery().Limit(pageSize)

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
			return errorResult("failed to search contacts: %v", err)
		}

		if len(contacts) == 0 {
			return textResult("No contacts found")
		}

		maps, err := entitiesToMaps(contacts)
		if err != nil {
			return errorResult("failed to convert contacts: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_contacts", page, pageSize)
	}
}

// createContactHandler returns a handler that creates a new contact.
func createContactHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateContactInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateContactInput) (*mcp.CallToolResult, any, error) {
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
			return errorResult("failed to create contact: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created contact: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created contact: %v", err)
		}

		return textResult("%s", string(data))
	}
}
