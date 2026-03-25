package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/autotask-mcp/services"
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

// entityToMap converts a typed entity to map[string]any for formatting/enhancement.
func entityToMap(entity any) (map[string]any, error) {
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("entityToMap: marshal: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("entityToMap: unmarshal: %w", err)
	}
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

// textResult builds a simple text CallToolResult.
func textResult(format string, args ...any) (*mcp.CallToolResult, any, error) {
	text := fmt.Sprintf(format, args...)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

// errorResult builds an error CallToolResult with IsError: true.
func errorResult(format string, args ...any) (*mcp.CallToolResult, any, error) {
	text := fmt.Sprintf(format, args...)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
		IsError: true,
	}, nil, nil
}

// searchResult builds a compact formatted search result with enhancement.
func searchResult(ctx context.Context, mapper *services.MappingCache, items []map[string]any, toolName string, page, pageSize int) (*mcp.CallToolResult, any, error) {
	if mapper != nil {
		mapper.EnhanceItems(ctx, items)
	}

	entityType := services.DetectEntityType(toolName)
	opts := services.FormatOptions{Page: page, PageSize: pageSize}
	compact := services.FormatCompactResponse(items, entityType, opts)

	data, err := json.Marshal(compact)
	if err != nil {
		return errorResult("failed to marshal response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// defaultPageSize returns the effective page size clamped to [1, maxVal].
// If requested is <= 0, defaultVal is used.
func defaultPageSize(requested, defaultVal, maxVal int) int {
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

// defaultPage returns the effective page number (minimum 1).
func defaultPage(requested int) int {
	if requested < 1 {
		return 1
	}
	return requested
}

// parseDate parses a date string in YYYY-MM-DD or RFC3339 format.
func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
