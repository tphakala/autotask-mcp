package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchProjectsInput defines the input parameters for searching projects.
type SearchProjectsInput struct {
	SearchTerm string `json:"searchTerm,omitempty" jsonschema:"Search term for project name"`
	CompanyID  int64  `json:"companyID,omitempty" jsonschema:"Filter by company ID"`
	Status     int    `json:"status,omitempty" jsonschema:"Filter by project status"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 100)"`
}

// CreateProjectInput defines the input parameters for creating a new project.
type CreateProjectInput struct {
	CompanyID      int64    `json:"companyID" jsonschema:"Company ID"`
	ProjectName    string   `json:"projectName" jsonschema:"Project name"`
	Status         int      `json:"status" jsonschema:"Project status (1=New, 2=In Progress, 5=Complete)"`
	Description    string   `json:"description,omitempty" jsonschema:"Project description"`
	StartDate      string   `json:"startDate,omitempty" jsonschema:"Start date (YYYY-MM-DD or ISO format)"`
	EndDate        string   `json:"endDate,omitempty" jsonschema:"End date (YYYY-MM-DD or ISO format)"`
	EstimatedHours *float64 `json:"estimatedHours,omitempty" jsonschema:"Estimated hours"`
}

// RegisterProjectTools registers all project-related MCP tools with the server.
func RegisterProjectTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_projects",
		Description: "Find projects by name substring, company ID, or status, returning a compact summary of matching records (up to maxResults, default 25, max 100). Use this to locate existing projects; to add a new one use autotask_create_project instead. Returns projects of every status when no status filter is given, including completed projects. Read-only.",
		Annotations: readOnlyTool("Search projects"),
	}, searchProjectsHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_project",
		Description: "Create a project for a company from a project name and status, with optional description, start and end dates, and estimated hours. Requires companyID, projectName, and status (1=New, 2=In Progress, 5=Complete); returns the created project including its new ID. To find existing projects use autotask_search_projects instead. Writes to Autotask.",
		Annotations: createTool("Create project"),
	}, createProjectHandler(client))
}

// searchProjectsHandler returns a handler that searches projects using the provided filters.
func searchProjectsHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchProjectsInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 100)

		q := autotask.NewQuery().Limit(maxResults + 1)

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
			return nil, services.CompactResponse{}, err
		}

		if len(projects) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(projects)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, mapper, maps, "autotask_search_projects", maxResults)
	}
}

// createProjectHandler returns a handler that creates a new project.
func createProjectHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateProjectInput) (*mcp.CallToolResult, map[string]any, error) {
		project := &entities.Project{
			CompanyID:   autotask.Set(in.CompanyID),
			ProjectName: autotask.Set(in.ProjectName),
			Status:      autotask.Set(int64(in.Status)),
		}

		if in.Description != "" {
			project.Description = autotask.Set(in.Description)
		}
		if in.StartDate != "" {
			t, err := parseDate(in.StartDate)
			if err != nil {
				return nil, nil, err
			}
			project.StartDateTime = autotask.Set(t)
		}
		if in.EndDate != "" {
			t, err := parseDate(in.EndDate)
			if err != nil {
				return nil, nil, err
			}
			project.EndDateTime = autotask.Set(t)
		}
		if in.EstimatedHours != nil {
			project.EstimatedHours = autotask.Set(*in.EstimatedHours)
		}

		created, err := autotask.Create[entities.Project](ctx, client, project)
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
