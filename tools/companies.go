package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// SearchCompaniesInput defines the input parameters for searching companies.
type SearchCompaniesInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search term for company name"`
	IsActive   *bool  `json:"isActive,omitempty" jsonschema:"Filter by active status"`
	Page       int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 200)"`
}

// CreateCompanyInput defines the input parameters for creating a new company.
type CreateCompanyInput struct {
	CompanyName     string `json:"companyName" jsonschema:"Company name"`
	CompanyType     int    `json:"companyType" jsonschema:"Company type ID"`
	Phone           string `json:"phone,omitempty" jsonschema:"Company phone number"`
	Address1        string `json:"address1,omitempty" jsonschema:"Company address line 1"`
	City            string `json:"city,omitempty" jsonschema:"Company city"`
	State           string `json:"state,omitempty" jsonschema:"Company state/province"`
	PostalCode      string `json:"postalCode,omitempty" jsonschema:"Company postal/ZIP code"`
	OwnerResourceID int64  `json:"ownerResourceID,omitempty" jsonschema:"Owner resource ID"`
	IsActive        *bool  `json:"isActive,omitempty" jsonschema:"Whether the company is active"`
}

// UpdateCompanyInput defines the input parameters for updating an existing company.
type UpdateCompanyInput struct {
	ID          int64  `json:"id" jsonschema:"Company ID to update"`
	CompanyName string `json:"companyName,omitempty" jsonschema:"Company name"`
	Phone       string `json:"phone,omitempty" jsonschema:"Company phone number"`
	Address1    string `json:"address1,omitempty" jsonschema:"Company address line 1"`
	City        string `json:"city,omitempty" jsonschema:"Company city"`
	State       string `json:"state,omitempty" jsonschema:"Company state/province"`
	PostalCode  string `json:"postalCode,omitempty" jsonschema:"Company postal/ZIP code"`
	IsActive    *bool  `json:"isActive,omitempty" jsonschema:"Whether the company is active"`
}

// RegisterCompanyTools registers all company-related MCP tools with the server.
func RegisterCompanyTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_companies",
		Description: "Search for companies in Autotask. Returns 25 results per page by default.",
	}, searchCompaniesHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_company",
		Description: "Create a new company in Autotask.",
	}, createCompanyHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_update_company",
		Description: "Update an existing company. Only provided fields are changed.",
	}, updateCompanyHandler(client))
}

// searchCompaniesHandler returns a handler that searches companies using the provided filters.
func searchCompaniesHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompaniesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchCompaniesInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 200)

		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("companyName", autotask.OpContains, in.SearchTerm)
		}
		if in.IsActive != nil {
			q.Where("isActive", autotask.OpEq, *in.IsActive)
		}

		companies, err := autotask.List[entities.Company](ctx, client, q)
		if err != nil {
			return errorResult("failed to search companies: %v", err)
		}

		if len(companies) == 0 {
			return textResult("No companies found")
		}

		maps, err := entitiesToMaps(companies)
		if err != nil {
			return errorResult("failed to convert companies: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_companies", page, pageSize)
	}
}

// createCompanyHandler returns a handler that creates a new company.
func createCompanyHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateCompanyInput) (*mcp.CallToolResult, any, error) {
		company := &entities.Company{
			CompanyName: autotask.Set(in.CompanyName),
			CompanyType: autotask.Set(int64(in.CompanyType)),
		}

		if in.Phone != "" {
			company.Phone = autotask.Set(in.Phone)
		}
		if in.Address1 != "" {
			company.Address1 = autotask.Set(in.Address1)
		}
		if in.City != "" {
			company.City = autotask.Set(in.City)
		}
		if in.State != "" {
			company.State = autotask.Set(in.State)
		}
		if in.PostalCode != "" {
			company.PostalCode = autotask.Set(in.PostalCode)
		}
		if in.OwnerResourceID != 0 {
			company.OwnerResourceID = autotask.Set(in.OwnerResourceID)
		}
		if in.IsActive != nil {
			company.IsActive = autotask.Set(*in.IsActive)
		}

		created, err := autotask.Create[entities.Company](ctx, client, company)
		if err != nil {
			return errorResult("failed to create company: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created company: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created company: %v", err)
		}

		return textResult("%s", string(data))
	}
}

// updateCompanyHandler returns a handler that updates an existing company.
func updateCompanyHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in UpdateCompanyInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UpdateCompanyInput) (*mcp.CallToolResult, any, error) {
		company := &entities.Company{
			ID: autotask.Set(in.ID),
		}

		if in.CompanyName != "" {
			company.CompanyName = autotask.Set(in.CompanyName)
		}
		if in.Phone != "" {
			company.Phone = autotask.Set(in.Phone)
		}
		if in.Address1 != "" {
			company.Address1 = autotask.Set(in.Address1)
		}
		if in.City != "" {
			company.City = autotask.Set(in.City)
		}
		if in.State != "" {
			company.State = autotask.Set(in.State)
		}
		if in.PostalCode != "" {
			company.PostalCode = autotask.Set(in.PostalCode)
		}
		if in.IsActive != nil {
			company.IsActive = autotask.Set(*in.IsActive)
		}

		updated, err := autotask.Update[entities.Company](ctx, client, company)
		if err != nil {
			return errorResult("failed to update company %d: %v", in.ID, err)
		}

		m, err := entityToMap(updated)
		if err != nil {
			return errorResult("failed to convert updated company: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal updated company: %v", err)
		}

		return textResult("%s", string(data))
	}
}
