package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
)

// RegisterAll registers every tool category with the MCP server.
func RegisterAll(s *mcp.Server, client *autotask.Client, mapper *services.MappingCache, picklist *services.PicklistCache) {
	RegisterConnectionTools(s, client)
	RegisterTicketTools(s, client, mapper)
	RegisterCompanyTools(s, client, mapper)
	RegisterContactTools(s, client, mapper)
	RegisterResourceTools(s, client)
	RegisterTimeEntryTools(s, client, mapper)
	RegisterProjectTools(s, client, mapper)
	RegisterTaskTools(s, client, mapper)
	RegisterNoteTools(s, client)
	RegisterAttachmentTools(s, client)
	RegisterFinancialTools(s, client, mapper)
	RegisterSalesTools(s, client)
	RegisterConfigItemTools(s, client, mapper)
	RegisterBillingTools(s, client, mapper)
	RegisterExpenseTools(s, client)
	RegisterPicklistTools(s, client, picklist)
}

// readOnlyTool annotates a tool that only reads from the Autotask API (open world).
// DestructiveHint is set false explicitly (redundant under ReadOnlyHint, but unambiguous for scanners).
func readOnlyTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, DestructiveHint: new(false), OpenWorldHint: new(true)}
}

// createTool annotates a tool that additively creates a new Autotask record.
func createTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, DestructiveHint: new(false), IdempotentHint: false, OpenWorldHint: new(true)}
}

// updateTool annotates a tool that overwrites fields on an existing Autotask record.
func updateTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, DestructiveHint: new(true), IdempotentHint: true, OpenWorldHint: new(true)}
}

// deleteTool annotates a tool that deletes an Autotask record.
func deleteTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, DestructiveHint: new(true), IdempotentHint: true, OpenWorldHint: new(true)}
}

// localReadTool annotates a read-only tool that operates on local metadata (closed world).
func localReadTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, OpenWorldHint: new(false)}
}

// entityToMap converts a typed entity to map[string]any for formatting/enhancement,
// and wraps untrusted user-supplied text fields in boundary markers.
func entityToMap(entity any) (map[string]any, error) {
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("entityToMap: marshal: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("entityToMap: unmarshal: %w", err)
	}
	services.FrameUntrustedMapFields(m)
	return m, nil
}

// entitiesToMaps converts a slice of typed entities to []map[string]any.
func entitiesToMaps[T any](entities []*T) ([]map[string]any, error) {
	maps := make([]map[string]any, 0, len(entities))
	for _, e := range entities {
		if e == nil {
			continue
		}
		m, err := entityToMap(e)
		if err != nil {
			return nil, err
		}
		maps = append(maps, m)
	}
	return maps, nil
}

// ConnectionStatusOut is the structured output for autotask_test_connection.
type ConnectionStatusOut struct {
	Success   bool   `json:"success" jsonschema:"Whether the connection test succeeded"`
	Entity    string `json:"entity" jsonschema:"Target entity used for verification"`
	CanCreate bool   `json:"canCreate" jsonschema:"Permission flag for creating records"`
	CanUpdate bool   `json:"canUpdate" jsonschema:"Permission flag for updating records"`
	CanQuery  bool   `json:"canQuery" jsonschema:"Permission flag for querying records"`
	Message   string `json:"message" jsonschema:"Human-readable connection status message"`
}

// DeleteResultOut is the structured output for delete operations.
type DeleteResultOut struct {
	Success bool   `json:"success" jsonschema:"Whether the deletion succeeded"`
	ID      int64  `json:"id" jsonschema:"ID of deleted entity"`
	Message string `json:"message" jsonschema:"Status message"`
}

// CategorySummary describes a tool category in list_categories output.
type CategorySummary struct {
	Description string   `json:"description" jsonschema:"Summary of tool category purpose"`
	ToolCount   int      `json:"toolCount" jsonschema:"Number of tools in this category"`
	Tools       []string `json:"tools" jsonschema:"Names of tools in this category"`
}

// ToolSummary describes a tool in list_category_tools output.
type ToolSummary struct {
	Name        string `json:"name" jsonschema:"Tool identifier"`
	Description string `json:"description" jsonschema:"Human-readable tool description"`
}

// CategoryToolsOut is the structured output for autotask_list_category_tools.
type CategoryToolsOut struct {
	Category    string        `json:"category" jsonschema:"Category identifier"`
	Description string        `json:"description" jsonschema:"Category purpose description"`
	Tools       []ToolSummary `json:"tools" jsonschema:"List of tools in category"`
}

// RouterOut is the structured output for autotask_router.
type RouterOut struct {
	Intent        string `json:"intent" jsonschema:"Original input intent query"`
	SuggestedTool string `json:"suggestedTool" jsonschema:"Best matching tool name"`
	Description   string `json:"description" jsonschema:"Guidance or tool description"`
}

// searchResult builds a compact formatted search result with enhancement.
func searchResult(ctx context.Context, mapper *services.MappingCache, items []map[string]any, toolName string, maxResults int) (*mcp.CallToolResult, services.CompactResponse, error) {
	if mapper != nil {
		mapper.EnhanceItems(ctx, items)
	}

	entityType := services.DetectEntityType(toolName)
	opts := services.FormatOptions{MaxResults: maxResults}
	compact := services.FormatCompactResponse(items, entityType, opts)

	return nil, compact, nil
}

// emptySearchResult returns a zero-item compact response for searches matching no records.
func emptySearchResult() (*mcp.CallToolResult, services.CompactResponse, error) {
	return nil, services.CompactResponse{
		Summary: services.CompactSummary{
			Returned: 0,
			HasMore:  false,
		},
		Items: []map[string]any{},
	}, nil
}

// defaultMaxResults returns the effective max results limit clamped to [1, maxVal].
// If requested is <= 0, defaultVal is used.
func defaultMaxResults(requested, defaultVal, maxVal int) int {
	size := requested
	if size <= 0 {
		size = defaultVal
	}
	if size > maxVal {
		size = maxVal
	}
	if size < 1 {
		size = 1
	}
	return size
}

// parseDate parses a date string in YYYY-MM-DD or RFC3339 format.
func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
