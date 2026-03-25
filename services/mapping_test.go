package services

import (
	"context"
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
)

func TestGetCompanyName_ZeroID(t *testing.T) {
	client := autotasktest.NewMockClient(t)
	cache := NewMappingCache(client)

	name := cache.GetCompanyName(context.Background(), 0)
	if name != "" {
		t.Errorf("expected empty string for id=0, got %q", name)
	}
}

func TestGetCompanyName_FromAPI(t *testing.T) {
	body := map[string]any{
		"item": map[string]any{
			"id":          float64(42),
			"companyName": "Acme Corp",
		},
	}
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Companies/42", 200, body),
	)
	cache := NewMappingCache(client)

	name := cache.GetCompanyName(context.Background(), 42)
	if name != "Acme Corp" {
		t.Errorf("expected Acme Corp, got %q", name)
	}
}

func TestGetCompanyName_CacheHit(t *testing.T) {
	body := map[string]any{
		"item": map[string]any{
			"id":          float64(42),
			"companyName": "Acme Corp",
		},
	}
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Companies/42", 200, body),
	)
	cache := NewMappingCache(client)

	// First call populates cache
	name1 := cache.GetCompanyName(context.Background(), 42)
	if name1 != "Acme Corp" {
		t.Fatalf("first call: expected Acme Corp, got %q", name1)
	}

	// Second call should hit cache (no new HTTP request needed, fixture only served once but mock doesn't error on reuse)
	name2 := cache.GetCompanyName(context.Background(), 42)
	if name2 != "Acme Corp" {
		t.Errorf("cache hit: expected Acme Corp, got %q", name2)
	}
}

func TestGetCompanyName_ErrorFallback(t *testing.T) {
	// No fixture registered — will return 404 / connection error
	client := autotasktest.NewMockClient(t)
	cache := NewMappingCache(client)

	name := cache.GetCompanyName(context.Background(), 99)
	if name != "Unknown (99)" {
		t.Errorf("expected 'Unknown (99)', got %q", name)
	}
}

func TestGetResourceName_ZeroID(t *testing.T) {
	client := autotasktest.NewMockClient(t)
	cache := NewMappingCache(client)

	name := cache.GetResourceName(context.Background(), 0)
	if name != "" {
		t.Errorf("expected empty string for id=0, got %q", name)
	}
}

func TestGetResourceName_FromAPI(t *testing.T) {
	body := map[string]any{
		"item": map[string]any{
			"id":        float64(7),
			"firstName": "John",
			"lastName":  "Doe",
		},
	}
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Resources/7", 200, body),
	)
	cache := NewMappingCache(client)

	name := cache.GetResourceName(context.Background(), 7)
	if name != "John Doe" {
		t.Errorf("expected John Doe, got %q", name)
	}
}

func TestGetResourceName_ErrorFallback(t *testing.T) {
	client := autotasktest.NewMockClient(t)
	cache := NewMappingCache(client)

	name := cache.GetResourceName(context.Background(), 55)
	if name != "Unknown (55)" {
		t.Errorf("expected 'Unknown (55)', got %q", name)
	}
}

func TestEnhanceItems(t *testing.T) {
	companyBody := map[string]any{
		"item": map[string]any{
			"id":          float64(10),
			"companyName": "Widget Co",
		},
	}
	resourceBody := map[string]any{
		"item": map[string]any{
			"id":        float64(20),
			"firstName": "Jane",
			"lastName":  "Smith",
		},
	}
	client := autotasktest.NewMockClient(t,
		autotasktest.WithFixture("GET", "/v1.0/Companies/10", 200, companyBody),
		autotasktest.WithFixture("GET", "/v1.0/Resources/20", 200, resourceBody),
	)
	cache := NewMappingCache(client)

	items := []map[string]any{
		{
			"id":                 float64(1),
			"companyID":          float64(10),
			"assignedResourceID": float64(20),
		},
	}

	cache.EnhanceItems(context.Background(), items)

	enhanced, ok := items[0]["_enhanced"].(map[string]any)
	if !ok {
		t.Fatal("expected _enhanced map to be added")
	}
	if v := enhanced["companyName"]; v != "Widget Co" {
		t.Errorf("expected companyName=Widget Co, got %v", v)
	}
	if v := enhanced["assignedResourceName"]; v != "Jane Smith" {
		t.Errorf("expected assignedResourceName=Jane Smith, got %v", v)
	}
}

func TestEnhanceItems_NoEnhancedWhenNoIDs(t *testing.T) {
	client := autotasktest.NewMockClient(t)
	cache := NewMappingCache(client)

	items := []map[string]any{
		{"id": float64(1), "title": "No IDs here"},
	}

	cache.EnhanceItems(context.Background(), items)

	if _, ok := items[0]["_enhanced"]; ok {
		t.Error("expected no _enhanced map when no ID fields present")
	}
}

func TestEnhanceItems_BatchPreload(t *testing.T) {
	// Use NewServer which supports query-based batch fetching.
	company := autotasktest.CompanyFixture()
	resource := autotasktest.ResourceFixture()
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntity(company),
		autotasktest.WithEntity(resource),
	)

	companyID, _ := company.ID.Get()
	resourceID, _ := resource.ID.Get()

	cache := NewMappingCache(client)
	items := []map[string]any{
		{"companyID": float64(companyID), "assignedResourceID": float64(resourceID)},
		{"companyID": float64(companyID)}, // same company, should be deduplicated
	}

	cache.EnhanceItems(context.Background(), items)

	// Both items should have enhancement.
	for i, item := range items {
		enhanced, ok := item["_enhanced"].(map[string]any)
		if !ok {
			t.Errorf("item[%d]: expected _enhanced map", i)
			continue
		}
		if enhanced["companyName"] == nil {
			t.Errorf("item[%d]: expected companyName", i)
		}
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		input any
		want  int64
		ok    bool
	}{
		{int64(42), 42, true},
		{float64(3.0), 3, true},
		{int(7), 7, true},
		{"string", 0, false},
		{nil, 0, false},
	}

	for _, tc := range tests {
		got, ok := toInt64(tc.input)
		if ok != tc.ok {
			t.Errorf("toInt64(%v): ok=%v, want %v", tc.input, ok, tc.ok)
		}
		if ok && got != tc.want {
			t.Errorf("toInt64(%v): got %d, want %d", tc.input, got, tc.want)
		}
	}
}
