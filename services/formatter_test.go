package services

import (
	"strings"
	"testing"
)

func TestFormatCompactResponse_StripsNonSummaryFields(t *testing.T) {
	items := []map[string]any{
		{
			"id":           float64(1),
			"ticketNumber": "T-001",
			"title":        "Test ticket",
			"status":       float64(5),
			"priority":     float64(2),
			"companyID":    float64(100),
			"description":  "Long description that should be stripped",
			"internalNotes": "Internal notes stripped",
		},
	}

	opts := FormatOptions{Page: 1, PageSize: 25}
	resp := FormatCompactResponse(items, "tickets", opts)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}

	item := resp.Items[0]

	// Summary fields should be present
	if _, ok := item["id"]; !ok {
		t.Error("expected id field")
	}
	if _, ok := item["title"]; !ok {
		t.Error("expected title field")
	}

	// Non-summary fields should be stripped
	if _, ok := item["description"]; ok {
		t.Error("description should have been stripped")
	}
	if _, ok := item["internalNotes"]; ok {
		t.Error("internalNotes should have been stripped")
	}
}

func TestFormatCompactResponse_HasMoreTrue(t *testing.T) {
	// Fill items equal to pageSize — HasMore should be true
	pageSize := 3
	items := make([]map[string]any, pageSize)
	for i := range items {
		items[i] = map[string]any{"id": float64(i)}
	}

	opts := FormatOptions{Page: 1, PageSize: pageSize}
	resp := FormatCompactResponse(items, "tickets", opts)

	if !resp.Summary.HasMore {
		t.Error("expected HasMore=true when items.length >= pageSize")
	}
	if resp.Summary.Returned != pageSize {
		t.Errorf("expected Returned=%d, got %d", pageSize, resp.Summary.Returned)
	}
}

func TestFormatCompactResponse_HasMoreFalse(t *testing.T) {
	// items < pageSize — HasMore should be false
	items := []map[string]any{
		{"id": float64(1)},
	}

	opts := FormatOptions{Page: 1, PageSize: 25}
	resp := FormatCompactResponse(items, "tickets", opts)

	if resp.Summary.HasMore {
		t.Error("expected HasMore=false when items.length < pageSize")
	}
}

func TestFormatCompactResponse_EnhancementFieldsInlined(t *testing.T) {
	items := []map[string]any{
		{
			"id":        float64(1),
			"title":     "Test",
			"companyID": float64(42),
			"_enhanced": map[string]any{
				"companyName":          "Acme Corp",
				"assignedResourceName": "John Doe",
				"resourceName":         "Jane Smith",
			},
		},
	}

	opts := FormatOptions{Page: 1, PageSize: 25}
	resp := FormatCompactResponse(items, "tickets", opts)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}

	item := resp.Items[0]

	if v, ok := item["company"]; !ok || v != "Acme Corp" {
		t.Errorf("expected company=Acme Corp, got %v", v)
	}
	if v, ok := item["assignedTo"]; !ok || v != "John Doe" {
		t.Errorf("expected assignedTo=John Doe, got %v", v)
	}
	if v, ok := item["resourceName"]; !ok || v != "Jane Smith" {
		t.Errorf("expected resourceName=Jane Smith, got %v", v)
	}

	// _enhanced should not appear directly
	if _, ok := item["_enhanced"]; ok {
		t.Error("_enhanced should not appear in output")
	}
}

func TestDetectEntityType(t *testing.T) {
	tests := []struct {
		toolName string
		want     string
	}{
		{"autotask_search_tickets", "tickets"},
		{"autotask_get_ticket", "tickets"},
		{"autotask_search_companies", "companies"},
		{"autotask_search_contacts", "contacts"},
		{"autotask_search_projects", "projects"},
		{"autotask_search_tasks", "tasks"},
		{"autotask_search_resources", "resources"},
		// billing_item_approval must come before billing_item
		{"autotask_search_billing_item_approval_levels", "billingItemApprovalLevels"},
		{"autotask_search_billing_items", "billingItems"},
		{"autotask_search_time_entries", "timeEntries"},
		{"autotask_get_time_entry", "timeEntries"},
		{"autotask_unknown_tool", ""},
	}

	for _, tc := range tests {
		t.Run(tc.toolName, func(t *testing.T) {
			got := DetectEntityType(tc.toolName)
			if got != tc.want {
				t.Errorf("DetectEntityType(%q) = %q, want %q", tc.toolName, got, tc.want)
			}
		})
	}
}

func TestDetectEntityType_BillingApprovalBeforeBillingItem(t *testing.T) {
	// Ensure billing_item_approval is matched before billing_item
	toolName := "autotask_search_billing_item_approval_levels"
	got := DetectEntityType(toolName)
	if got != "billingItemApprovalLevels" {
		t.Errorf("expected billingItemApprovalLevels, got %q", got)
	}

	// A tool with just billing_item (no approval) goes to billingItems
	toolName2 := "autotask_search_billing_items"
	if !strings.Contains(toolName2, "billing_item") {
		t.Fatal("test setup error")
	}
	got2 := DetectEntityType(toolName2)
	if got2 != "billingItems" {
		t.Errorf("expected billingItems, got %q", got2)
	}
}

func TestFormatCompactResponse_UnknownEntityType(t *testing.T) {
	// Unknown entity type should pass all non-enhanced fields through
	items := []map[string]any{
		{
			"id":        float64(1),
			"someField": "value",
			"_enhanced": map[string]any{"companyName": "Acme"},
		},
	}

	opts := FormatOptions{Page: 1, PageSize: 25}
	resp := FormatCompactResponse(items, "unknownEntity", opts)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item")
	}
	item := resp.Items[0]

	if _, ok := item["id"]; !ok {
		t.Error("expected id field to pass through for unknown entity type")
	}
	if _, ok := item["someField"]; !ok {
		t.Error("expected someField to pass through for unknown entity type")
	}
	if _, ok := item["_enhanced"]; ok {
		t.Error("_enhanced should not appear in output")
	}
	// enhancement should still be inlined
	if v, ok := item["company"]; !ok || v != "Acme" {
		t.Errorf("expected company=Acme, got %v", v)
	}
}
