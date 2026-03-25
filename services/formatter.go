package services

import (
	"fmt"
	"strings"
)

// SummaryFields defines the essential fields to include in compact responses per entity type.
var SummaryFields = map[string][]string{
	"tickets":                   {"id", "ticketNumber", "title", "status", "priority", "companyID", "assignedResourceID", "createDate", "dueDateTime"},
	"companies":                 {"id", "companyName", "isActive", "phone", "city", "state"},
	"contacts":                  {"id", "firstName", "lastName", "emailAddress", "companyID"},
	"projects":                  {"id", "projectName", "status", "companyID", "projectLeadResourceID", "startDate", "endDate"},
	"tasks":                     {"id", "title", "status", "projectID", "assignedResourceID", "percentComplete"},
	"resources":                 {"id", "firstName", "lastName", "email", "isActive"},
	"billingItems":              {"id", "itemName", "companyID", "ticketID", "projectID", "postedDate", "totalAmount", "invoiceID", "billingItemType"},
	"billingItemApprovalLevels": {"id", "timeEntryID", "approvalLevel", "approvalResourceID", "approvalDateTime"},
	"timeEntries":               {"id", "resourceID", "ticketID", "projectID", "taskID", "dateWorked", "hoursWorked", "summaryNotes"},
}

// CompactSearchTools is the set of tool names that use compact formatting.
var CompactSearchTools = map[string]bool{
	"autotask_search_tickets":                    true,
	"autotask_search_companies":                  true,
	"autotask_search_contacts":                   true,
	"autotask_search_projects":                   true,
	"autotask_search_tasks":                      true,
	"autotask_search_resources":                  true,
	"autotask_search_billing_items":              true,
	"autotask_search_billing_item_approval_levels": true,
	"autotask_search_time_entries":               true,
}

// FormatOptions controls paging metadata for compact responses.
type FormatOptions struct {
	Page     int
	PageSize int
}

// CompactSummary holds pagination metadata returned with a compact response.
type CompactSummary struct {
	Returned int    `json:"returned"`
	HasMore  bool   `json:"hasMore"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Hint     string `json:"hint,omitempty"`
}

// CompactResponse is the complete compact-formatted search result.
type CompactResponse struct {
	Summary CompactSummary           `json:"summary"`
	Items   []map[string]any         `json:"items"`
}

// FormatCompactResponse formats a slice of raw items into a compact response,
// picking only summary fields for the given entity type.
func FormatCompactResponse(items []map[string]any, entityType string, opts FormatOptions) CompactResponse {
	compact := make([]map[string]any, 0, len(items))
	for _, item := range items {
		compact = append(compact, pickSummaryFields(item, entityType))
	}

	hasMore := len(items) >= opts.PageSize && opts.PageSize > 0
	hint := ""
	if hasMore {
		hint = fmt.Sprintf("More results available. Use page=%d to retrieve the next page.", opts.Page+1)
	}

	return CompactResponse{
		Summary: CompactSummary{
			Returned: len(compact),
			HasMore:  hasMore,
			Page:     opts.Page,
			PageSize: opts.PageSize,
			Hint:     hint,
		},
		Items: compact,
	}
}

// DetectEntityType maps a tool name to an entity type key used in SummaryFields.
// Order matters: more specific patterns are checked first.
func DetectEntityType(toolName string) string {
	switch {
	case strings.Contains(toolName, "billing_item_approval"):
		return "billingItemApprovalLevels"
	case strings.Contains(toolName, "billing_item"):
		return "billingItems"
	case strings.Contains(toolName, "time_entr"):
		return "timeEntries"
	case strings.Contains(toolName, "ticket"):
		return "tickets"
	case strings.Contains(toolName, "compan"):
		return "companies"
	case strings.Contains(toolName, "contact"):
		return "contacts"
	case strings.Contains(toolName, "project"):
		return "projects"
	case strings.Contains(toolName, "_task"):
		return "tasks"
	case strings.Contains(toolName, "resource"):
		return "resources"
	default:
		return ""
	}
}

// pickSummaryFields returns a new map containing only the summary fields for the
// entity type. Enhancement fields from the "_enhanced" sub-map are inlined.
func pickSummaryFields(item map[string]any, entityType string) map[string]any {
	if item == nil {
		return make(map[string]any)
	}
	fields, ok := SummaryFields[entityType]
	result := make(map[string]any)

	if ok {
		for _, f := range fields {
			if v, exists := item[f]; exists {
				result[f] = v
			}
		}
	} else {
		// Unknown entity type: copy all fields except _enhanced
		for k, v := range item {
			if k != "_enhanced" {
				result[k] = v
			}
		}
	}

	// Inline enhancement fields from the _enhanced sub-map.
	if enhanced, ok := item["_enhanced"].(map[string]any); ok {
		if v, ok := enhanced["companyName"]; ok {
			result["company"] = v
		}
		if v, ok := enhanced["assignedResourceName"]; ok {
			result["assignedTo"] = v
		}
		if v, ok := enhanced["resourceName"]; ok {
			result["resourceName"] = v
		}
	}

	return result
}
