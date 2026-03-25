package services

import (
	"context"
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/metadata"
)

func fieldsFixture() map[string]any {
	return map[string]any{
		"fields": []map[string]any{
			{
				"name":       "status",
				"label":      "Status",
				"dataType":   "integer",
				"isRequired": false,
				"isReadOnly": false,
				"isPickList": true,
				"picklistValues": []map[string]any{
					{"value": "1", "label": "New", "isActive": true},
					{"value": "5", "label": "Complete", "isActive": true},
				},
			},
			{
				"name":       "title",
				"label":      "Title",
				"dataType":   "string",
				"isRequired": true,
				"isReadOnly": false,
				"isPickList": false,
			},
		},
	}
}

func TestGetFields_FromAPI(t *testing.T) {
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	fields, err := cache.GetFields(context.Background(), "Tickets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

func TestGetFields_CacheHit(t *testing.T) {
	// The fixture is only registered once; second call must use cache
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	// First call
	fields1, err := cache.GetFields(context.Background(), "Tickets")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Second call — should hit cache
	fields2, err := cache.GetFields(context.Background(), "Tickets")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if len(fields1) != len(fields2) {
		t.Errorf("cache hit: field counts differ %d vs %d", len(fields1), len(fields2))
	}
}

func TestGetPicklistValues_Success(t *testing.T) {
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	values, err := cache.GetPicklistValues(context.Background(), "Tickets", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 picklist values, got %d", len(values))
	}

	var labels []string
	for _, v := range values {
		labels = append(labels, v.Label)
	}
	if labels[0] != "New" || labels[1] != "Complete" {
		t.Errorf("unexpected labels: %v", labels)
	}
}

func TestGetPicklistValues_NotAPicklist(t *testing.T) {
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	_, err := cache.GetPicklistValues(context.Background(), "Tickets", "title")
	if err == nil {
		t.Error("expected error for non-picklist field")
	}
}

func TestGetPicklistValues_FieldNotFound(t *testing.T) {
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	_, err := cache.GetPicklistValues(context.Background(), "Tickets", "nonExistentField")
	if err == nil {
		t.Error("expected error for non-existent field")
	}
}

func TestGetFields_APIError(t *testing.T) {
	// No fixture — will get a 404 or similar error
	client := autotasktest.NewMockClient(t)
	cache := NewPicklistCache(client)

	_, err := cache.GetFields(context.Background(), "Tickets")
	if err == nil {
		t.Error("expected error when API returns no fixture")
	}
}

// Verify the PicklistCache properly uses the metadata.FieldInfo type
func TestPicklistCacheTypes(t *testing.T) {
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Tickets/entityInformation/fields", 200, fieldsFixture()),
	)
	cache := NewPicklistCache(client)

	fields, err := cache.GetFields(context.Background(), "Tickets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the type is correct
	var _ []metadata.FieldInfo = fields
	if len(fields) == 0 {
		t.Fatal("expected non-empty fields slice")
	}
	if fields[0].Name != "status" {
		t.Errorf("expected first field name=status, got %q", fields[0].Name)
	}
	if !fields[0].IsPickList {
		t.Error("expected status to be a picklist field")
	}
}
