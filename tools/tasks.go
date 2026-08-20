package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
)

// SearchTasksInput defines the input parameters for searching tasks.
type SearchTasksInput struct {
	SearchTerm         string `json:"searchTerm,omitempty" jsonschema:"Search term for task title"`
	ProjectID          int64  `json:"projectID,omitempty" jsonschema:"Filter by project ID"`
	Status             int    `json:"status,omitempty" jsonschema:"Filter by task status (1=New, 2=In Progress, 5=Complete)"`
	AssignedResourceID int64  `json:"assignedResourceID,omitempty" jsonschema:"Filter by assigned resource ID"`
	MaxResults         int    `json:"maxResults,omitempty" jsonschema:"Maximum results to return (default 25, max 100)"`
}

// CreateTaskInput defines the input parameters for creating a new task.
type CreateTaskInput struct {
	ProjectID          int64    `json:"projectID" jsonschema:"Project ID"`
	Title              string   `json:"title" jsonschema:"Task title"`
	Status             int      `json:"status" jsonschema:"Task status (1=New, 2=In Progress, 5=Complete)"`
	Description        string   `json:"description,omitempty" jsonschema:"Task description"`
	AssignedResourceID int64    `json:"assignedResourceID,omitempty" jsonschema:"Assigned resource ID"`
	EstimatedHours     *float64 `json:"estimatedHours,omitempty" jsonschema:"Estimated hours"`
	StartDateTime      string   `json:"startDateTime,omitempty" jsonschema:"Start date/time (ISO format)"`
	EndDateTime        string   `json:"endDateTime,omitempty" jsonschema:"End date/time (ISO format)"`
}

// RegisterTaskTools registers all task-related MCP tools with the server.
func RegisterTaskTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_tasks",
		Description: "Find project tasks by title substring, parent project ID, status, or assigned resource, returning a compact summary of matching records (up to maxResults, default 25, max 100). Use this to locate tasks within a project; to add a new task use autotask_create_task instead. Returns tasks of every status when no status filter is given, including completed tasks. Read-only.",
		Annotations: readOnlyTool("Search tasks"),
	}, searchTasksHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_task",
		Description: "Add a task to an existing project from a title and status, with optional description, resource assignment, estimated hours, and start and end dates. Requires projectID, title, and status (1=New, 2=In Progress, 5=Complete); returns the created task including its new ID. To find existing tasks use autotask_search_tasks instead. Writes to Autotask.",
		Annotations: createTool("Create task"),
	}, createTaskHandler(client))
}

// searchTasksHandler returns a handler that searches tasks using the provided filters.
func searchTasksHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTasksInput) (*mcp.CallToolResult, services.CompactResponse, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTasksInput) (*mcp.CallToolResult, services.CompactResponse, error) {
		maxResults := defaultMaxResults(in.MaxResults, 25, 100)

		q := autotask.NewQuery().Limit(maxResults)

		if in.SearchTerm != "" {
			q.Where("title", autotask.OpContains, in.SearchTerm)
		}
		if in.ProjectID != 0 {
			q.Where("projectID", autotask.OpEq, in.ProjectID)
		}
		if in.Status != 0 {
			q.Where("status", autotask.OpEq, in.Status)
		}
		if in.AssignedResourceID != 0 {
			q.Where("assignedResourceID", autotask.OpEq, in.AssignedResourceID)
		}

		tasks, err := autotask.List[entities.Task](ctx, client, q)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		if len(tasks) == 0 {
			return emptySearchResult()
		}

		maps, err := entitiesToMaps(tasks)
		if err != nil {
			return nil, services.CompactResponse{}, err
		}

		return searchResult(ctx, mapper, maps, "autotask_search_tasks", maxResults)
	}
}

// createTaskHandler returns a handler that creates a new task.
func createTaskHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTaskInput) (*mcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTaskInput) (*mcp.CallToolResult, map[string]any, error) {
		task := &entities.Task{
			ProjectID: autotask.Set(in.ProjectID),
			Title:     autotask.Set(in.Title),
			Status:    autotask.Set(int64(in.Status)),
		}

		if in.Description != "" {
			task.Description = autotask.Set(in.Description)
		}
		if in.AssignedResourceID != 0 {
			task.AssignedResourceID = autotask.Set(in.AssignedResourceID)
		}
		if in.EstimatedHours != nil {
			task.EstimatedHours = autotask.Set(*in.EstimatedHours)
		}
		if in.StartDateTime != "" {
			t, err := parseDate(in.StartDateTime)
			if err != nil {
				return nil, nil, err
			}
			task.StartDateTime = autotask.Set(t)
		}
		if in.EndDateTime != "" {
			t, err := parseDate(in.EndDateTime)
			if err != nil {
				return nil, nil, err
			}
			task.EndDateTime = autotask.Set(t)
		}

		created, err := autotask.Create[entities.Task](ctx, client, task)
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
