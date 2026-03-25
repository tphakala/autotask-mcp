package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// SearchProjectsInput defines the input parameters for searching projects.
type SearchProjectsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search term for project name"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	Status     int    `json:"status,omitempty" jsonschema:"Filter by project status"`
	Page       int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize   int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 100)"`
}

// CreateProjectInput defines the input parameters for creating a new project.
type CreateProjectInput struct {
	CompanyID      int64   `json:"companyID" jsonschema:"Company ID"`
	ProjectName    string  `json:"projectName" jsonschema:"Project name"`
	Status         int     `json:"status" jsonschema:"Project status (1=New, 2=In Progress, 5=Complete)"`
	Description    string  `json:"description,omitempty" jsonschema:"Project description"`
	StartDate      string  `json:"startDate,omitempty" jsonschema:"Start date (YYYY-MM-DD or ISO format)"`
	EndDate        string  `json:"endDate,omitempty" jsonschema:"End date (YYYY-MM-DD or ISO format)"`
	EstimatedHours *float64 `json:"estimatedHours,omitempty" jsonschema:"Estimated hours"`
}

// RegisterProjectTools registers all project-related MCP tools with the server.
func RegisterProjectTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_projects",
		Description: "Search for projects in Autotask. Returns 25 results per page by default.",
	}, searchProjectsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_project",
		Description: "Create a new project in Autotask.",
	}, createProjectHandler(client))
}

// searchProjectsHandler returns a handler that searches projects using the provided filters.
func searchProjectsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectsInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 100)

		q := autotask.NewQuery().Limit(pageSize)

		if in.SearchTerm != "" {
			q.Where("projectName", autotask.OpContains, in.SearchTerm)
		}
		if in.CompanyID != 0 {
			q.Where("companyID", autotask.OpEq, in.CompanyID)
		}
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		}

		projects, err := autotask.List[entities.Project](ctx, client, q)
		if err != nil {
			return errorResult("failed to search projects: %v", err)
		}

		if len(projects) == 0 {
			return textResult("No projects found")
		}

		maps, err := entitiesToMaps(projects)
		if err != nil {
			return errorResult("failed to convert projects: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_projects", page, pageSize)
	}
}

// createProjectHandler returns a handler that creates a new project.
func createProjectHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectInput) (*mcp.CallToolResult, any, error) {
		project := &entities.Project{
			CompanyID:   autotask.Set(in.CompanyID),
			ProjectName: autotask.Set(in.ProjectName),
			Status:      autotask.Set(in.Status),
		}

		if in.Description != "" {
			project.Description = autotask.Set(in.Description)
		}
		if in.StartDate != "" {
			t, err := parseDate(in.StartDate)
			if err != nil {
				return errorResult("invalid startDate format (expected YYYY-MM-DD or ISO format): %v", err)
			}
			project.StartDateTime = autotask.Set(t)
		}
		if in.EndDate != "" {
			t, err := parseDate(in.EndDate)
			if err != nil {
				return errorResult("invalid endDate format (expected YYYY-MM-DD or ISO format): %v", err)
			}
			project.EndDateTime = autotask.Set(t)
		}
		if in.EstimatedHours != nil {
			project.EstimatedHours = autotask.Set(*in.EstimatedHours)
		}

		created, err := autotask.Create[entities.Project](ctx, client, project)
		if err != nil {
			return errorResult("failed to create project: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created project: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created project: %v", err)
		}

		return textResult("%s", string(data))
	}
}

