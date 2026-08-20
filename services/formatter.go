package services

import (
	"regexp"
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

// FormatOptions controls result limits for compact responses.
type FormatOptions struct {
	MaxResults int
}

// CompactSummary holds bounded search metadata returned with a compact response.
type CompactSummary struct {
	Returned   int    `json:"returned" jsonschema:"Number of items returned in this response"`
	HasMore    bool   `json:"hasMore" jsonschema:"Whether more items match the search criteria"`
	MaxResults int    `json:"maxResults,omitempty" jsonschema:"Effective max results limit used"`
	Hint       string `json:"hint,omitempty" jsonschema:"Guidance when results are capped"`
}

// CompactResponse is the complete compact-formatted search result.
type CompactResponse struct {
	Summary CompactSummary   `json:"summary" jsonschema:"Search metadata summary"`
	Items   []map[string]any `json:"items" jsonschema:"Matching entity records with framed untrusted fields"`
}

// untrustedFields lists the known external user-supplied fields that must be framed.
var untrustedFields = []string{
	"description", "title", "note", "summaryNotes", "details",
	"problemDescription", "resolution", "internalNotes", "cause", "name",
	"projectName", "companyName", "referenceTitle", "referenceName",
	"fileName", "serviceName", "serviceBundleName",
	// Short free-text identity fields that are also externally supplied
	// (contacts are routinely auto-created from inbound customer email).
	"firstName", "lastName", "emailAddress", "email", "itemName", "contractName",
	// Derived reference names inlined from the _enhanced sub-map by pickSummaryFields;
	// these carry the same customer-controlled text as their source *Name fields.
	"company", "assignedTo", "resourceName", "projectLead",
}

// untrustedTagPattern matches a boundary tag tolerating whitespace and case
// variants (e.g. "</untrusted_content >"), so an attacker cannot slip a
// near-miss closing tag past the exact-match escaping.
var untrustedTagPattern = regexp.MustCompile(`(?i)<\s*(/?)\s*untrusted_content\s*>`)

const (
	untrustedPrefix = "<untrusted_content>\n"
	untrustedSuffix = "\n</untrusted_content>"
)

// FrameUntrustedContent wraps external user-supplied text with untrusted data boundary markers,
// escaping any inner boundary tags to prevent breakout.
func FrameUntrustedContent(content string) string {
	if content == "" {
		return content
	}
	// Strip one layer of our own prior framing for idempotency. The length guard
	// prevents an inverted slice when the value is a bare wrapper with no inner
	// content (prefix and suffix share the newline, so a 40-byte overlap would
	// otherwise compute content[20:19] and panic).
	if len(content) >= len(untrustedPrefix)+len(untrustedSuffix) &&
		strings.HasPrefix(content, untrustedPrefix) && strings.HasSuffix(content, untrustedSuffix) {
		content = content[len(untrustedPrefix) : len(content)-len(untrustedSuffix)]
	}
	sanitized := untrustedTagPattern.ReplaceAllString(content, "&lt;${1}untrusted_content&gt;")
	return untrustedPrefix + sanitized + untrustedSuffix
}

// FrameUntrustedMapFields inspects known external user-supplied fields on an entity map
// and wraps their string values in untrusted content boundary markers.
func FrameUntrustedMapFields(m map[string]any) {
	if m == nil {
		return
	}
	for _, field := range untrustedFields {
		if val, ok := m[field].(string); ok && val != "" {
			m[field] = FrameUntrustedContent(val)
		}
	}
	// The _enhanced sub-map (emitted by get-detail handlers before flattening)
	// holds resolved human-readable names, all customer/staff-controlled text.
	// Frame every string value so the get path matches the search path, which
	// inlines and frames these same names via pickSummaryFields.
	if enhanced, ok := m["_enhanced"].(map[string]any); ok {
		for k, v := range enhanced {
			if s, ok := v.(string); ok && s != "" {
				enhanced[k] = FrameUntrustedContent(s)
			}
		}
	}
}

// FormatCompactResponse formats a slice of raw items into a compact response,
// picking only summary fields for the given entity type and framing untrusted text.
func FormatCompactResponse(items []map[string]any, entityType string, opts FormatOptions) CompactResponse {
	compact := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summaryItem := pickSummaryFields(item, entityType)
		FrameUntrustedMapFields(summaryItem)
		compact = append(compact, summaryItem)
	}

	hasMore := len(items) >= opts.MaxResults && opts.MaxResults > 0
	hint := ""
	if hasMore {
		hint = "Maximum result limit reached. Use narrower search filters to find specific records."
	}

	return CompactResponse{
		Summary: CompactSummary{
			Returned:   len(compact),
			HasMore:    hasMore,
			MaxResults: opts.MaxResults,
			Hint:       hint,
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
		if v, ok := enhanced["projectLeadResourceName"]; ok {
			result["projectLead"] = v
		}
	}

	return result
}
