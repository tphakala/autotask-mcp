package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/entities"
	"github.com/tphakala/autotask-mcp/services"
)

// SearchTasksInput defines the input parameters for searching tasks.
type SearchTasksInput struct {
	SearchTerm         string `json:"searchTerm,omitempty" jsonschema:"Search term for task title"`
	ProjectID          int64  `json:"projectID,omitempty" jsonschema:"Filter by project ID"`
	Status             int    `json:"status,omitempty" jsonschema:"Filter by task status (1=New, 2=In Progress, 5=Complete)"`
	AssignedResourceID int64  `json:"assignedResourceID,omitempty" jsonschema:"Filter by assigned resource ID"`
	Page               int    `json:"page,omitempty" jsonschema:"Page number (default 1)"`
	PageSize           int    `json:"pageSize,omitempty" jsonschema:"Results per page (default 25, max 100)"`
}

// CreateTaskInput defines the input parameters for creating a new task.
type CreateTaskInput struct {
	ProjectID          int64   `json:"projectID" jsonschema:"Project ID"`
	Title              string  `json:"title" jsonschema:"Task title"`
	Status             int     `json:"status" jsonschema:"Task status (1=New, 2=In Progress, 5=Complete)"`
	Description        string  `json:"description,omitempty" jsonschema:"Task description"`
	AssignedResourceID int64   `json:"assignedResourceID,omitempty" jsonschema:"Assigned resource ID"`
	EstimatedHours     *float64 `json:"estimatedHours,omitempty" jsonschema:"Estimated hours"`
	StartDateTime      string  `json:"startDateTime,omitempty" jsonschema:"Start date/time (ISO format)"`
	EndDateTime        string  `json:"endDateTime,omitempty" jsonschema:"End date/time (ISO format)"`
}

// RegisterTaskTools registers all task-related MCP tools with the server.
func RegisterTaskTools(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_search_tasks",
		Description: "Search for project tasks in Autotask. Returns 25 results per page by default.",
	}, searchTasksHandler(client, mapper))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "autotask_create_task",
		Description: "Create a new project task in Autotask.",
	}, createTaskHandler(client))
}

// searchTasksHandler returns a handler that searches tasks using the provided filters.
func searchTasksHandler(client *autotask.Client, mapper *services.MappingCache) func(ctx context.Context, req *mcp.CallToolRequest, in SearchTasksInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchTasksInput) (*mcp.CallToolResult, any, error) {
		page := defaultPage(in.Page)
		pageSize := defaultPageSize(in.PageSize, 25, 100)

		q := autotask.NewQuery().Limit(pageSize)

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
			return errorResult("failed to search tasks: %v", err)
		}

		if len(tasks) == 0 {
			return textResult("No tasks found")
		}

		maps, err := entitiesToMaps(tasks)
		if err != nil {
			return errorResult("failed to convert tasks: %v", err)
		}

		return searchResult(ctx, mapper, maps, "autotask_search_tasks", page, pageSize)
	}
}

// createTaskHandler returns a handler that creates a new task.
func createTaskHandler(client *autotask.Client) func(ctx context.Context, req *mcp.CallToolRequest, in CreateTaskInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CreateTaskInput) (*mcp.CallToolResult, any, error) {
		task := &entities.Task{
			ProjectID: autotask.Set(in.ProjectID),
			Title:     autotask.Set(in.Title),
			Status:    autotask.Set(in.Status),
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
				return errorResult("invalid startDateTime format (expected ISO format): %v", err)
			}
			task.StartDateTime = autotask.Set(t)
		}
		if in.EndDateTime != "" {
			t, err := parseDate(in.EndDateTime)
			if err != nil {
				return errorResult("invalid endDateTime format (expected ISO format): %v", err)
			}
			task.EndDateTime = autotask.Set(t)
		}

		created, err := autotask.Create[entities.Task](ctx, client, task)
		if err != nil {
			return errorResult("failed to create task: %v", err)
		}

		m, err := entityToMap(created)
		if err != nil {
			return errorResult("failed to convert created task: %v", err)
		}

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return errorResult("failed to marshal created task: %v", err)
		}

		return textResult("%s", string(data))
	}
}
