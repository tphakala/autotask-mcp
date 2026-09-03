package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/tphakala/autotask-mcp/services"
)

func TestDefaultMaxResults_Default(t *testing.T) {
	// When requested <= 0, use defaultVal
	got := defaultMaxResults(0, 25, 100)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}

	got = defaultMaxResults(-5, 25, 100)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}
}

func TestDefaultMaxResults_Clamped(t *testing.T) {
	// When requested > maxVal, clamp to maxVal
	got := defaultMaxResults(500, 25, 100)
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestDefaultMaxResults_Valid(t *testing.T) {
	// When requested is within bounds, return it
	got := defaultMaxResults(50, 25, 100)
	if got != 50 {
		t.Errorf("expected 50, got %d", got)
	}
}

func TestDefaultMaxResults_MinOne(t *testing.T) {
	// defaultVal=0 and maxVal=0 edge case: result should be at least 1
	got := defaultMaxResults(0, 0, 0)
	if got != 1 {
		t.Errorf("expected minimum 1, got %d", got)
	}
}

func TestDefaultMaxResults_ExactMax(t *testing.T) {
	// Exactly at max boundary
	got := defaultMaxResults(100, 25, 100)
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestEntityToMap_SimpleStruct(t *testing.T) {
	type testEntity struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	e := testEntity{ID: 42, Name: "Test"}
	m, err := entityToMap(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, ok := m["id"]; !ok {
		t.Error("expected id field")
	} else if v != float64(42) {
		t.Errorf("expected id=42, got %v", v)
	}

	if v, ok := m["name"]; !ok {
		t.Error("expected name field")
	} else if !strings.Contains(v.(string), "<untrusted_content>") || !strings.Contains(v.(string), "Test") {
		t.Errorf("expected framed name, got %v", v)
	}
}

func TestEntityToMap_Pointer(t *testing.T) {
	type testEntity struct {
		ID    int64   `json:"id"`
		Score float64 `json:"score"`
	}

	e := &testEntity{ID: 7, Score: 9.5}
	m, err := entityToMap(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m["id"] != float64(7) {
		t.Errorf("expected id=7, got %v", m["id"])
	}
	if m["score"] != 9.5 {
		t.Errorf("expected score=9.5, got %v", m["score"])
	}
}

func TestEntitiesToMaps_MultipleEntities(t *testing.T) {
	type testEntity struct {
		ID int64 `json:"id"`
	}

	entities := []*testEntity{{ID: 1}, {ID: 2}, {ID: 3}}
	maps, err := entitiesToMaps(entities)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(maps) != 3 {
		t.Errorf("expected 3 maps, got %d", len(maps))
	}
	for i, m := range maps {
		expectedID := float64(i + 1)
		if m["id"] != expectedID {
			t.Errorf("maps[%d]: expected id=%v, got %v", i, expectedID, m["id"])
		}
	}
}

func TestEntitiesToMaps_Empty(t *testing.T) {
	type testEntity struct {
		ID int64 `json:"id"`
	}

	maps, err := entitiesToMaps([]*testEntity{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maps) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(maps))
	}
}

func TestSearchResult(t *testing.T) {
	items := []map[string]any{
		{"id": float64(101), "companyName": "Acme Corp", "phone": "555-1234"},
		{"id": float64(102), "companyName": "Beta Inc", "phone": "555-5678"},
	}

	result, compact, err := searchResult(context.Background(), nil, items, "autotask_search_companies", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil *CallToolResult so SDK generates it, got %v", result)
	}
	if compact.Summary.Returned != 2 {
		t.Errorf("expected returned=2, got %d", compact.Summary.Returned)
	}
	if compact.Summary.HasMore {
		t.Errorf("expected hasMore=false, got %v", compact.Summary.HasMore)
	}
	if len(compact.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(compact.Items))
	}
}

func TestEmptySearchResult(t *testing.T) {
	result, compact, err := emptySearchResult()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil *CallToolResult, got %v", result)
	}
	if compact.Summary.Returned != 0 {
		t.Errorf("expected returned=0, got %d", compact.Summary.Returned)
	}
	if compact.Summary.HasMore {
		t.Errorf("expected hasMore=false, got %v", compact.Summary.HasMore)
	}
	if len(compact.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(compact.Items))
	}
}

// TestSearchResult_HasMoreBoundaries pins every branch of searchResult's accurate
// hasMore: over the limit (trim + hasMore), exactly the limit below the ceiling
// (accurate false, the case the old len>=maxResults heuristic got wrong), exactly the
// ceiling (conservative hasMore fallback), and under the limit (false). The ceiling
// branch is otherwise unreachable in a wire test without seeding 500 records.
func TestSearchResult_HasMoreBoundaries(t *testing.T) {
	mk := func(n int) []map[string]any {
		items := make([]map[string]any, n)
		for i := range items {
			items[i] = map[string]any{"id": float64(i)}
		}
		return items
	}

	cases := []struct {
		name        string
		count       int
		maxResults  int
		wantRet     int
		wantHasMore bool
	}{
		{"over limit", 4, 3, 3, true},
		{"exactly limit below ceiling", 3, 3, 3, false},
		{"exactly ceiling", autotaskMaxRecords, autotaskMaxRecords, autotaskMaxRecords, true},
		{"under limit", 2, 3, 2, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, resp, err := searchResult(context.Background(), nil, mk(tc.count), "autotask_search_tickets", tc.maxResults)
			if err != nil {
				t.Fatalf("searchResult: %v", err)
			}
			if resp.Summary.Returned != tc.wantRet {
				t.Errorf("Returned = %d, want %d", resp.Summary.Returned, tc.wantRet)
			}
			if resp.Summary.HasMore != tc.wantHasMore {
				t.Errorf("HasMore = %v, want %v", resp.Summary.HasMore, tc.wantHasMore)
			}
		})
	}
}

// TestRegisterAll_ToolSetMatchesDispatcherAndCategories pins the three
// hand-maintained tool lists together: the tools RegisterAll actually registers
// on the server (full mode), the buildToolDispatcher runners (lazy execution via
// autotask_execute_tool), and ToolCategories (lazy discovery and routing). A tool
// added to one but not the others would compile and pass every other test, yet be
// silently uncallable or undiscoverable in lazy mode. This asserts set-equality of
// all three so that gap cannot ship.
func TestRegisterAll_ToolSetMatchesDispatcherAndCategories(t *testing.T) {
	cs, client := setupWireTest(t)
	ctx := context.Background()

	// The set the built server actually exposes in full mode.
	registered := map[string]struct{}{}
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("enumerating registered tools: %v", err)
		}
		registered[tool.Name] = struct{}{}
	}
	if len(registered) == 0 {
		t.Fatal("RegisterAll registered no tools")
	}

	// The lazy dispatcher set.
	mapper := services.NewMappingCache(client)
	picklist := services.NewPicklistCache(client)
	dispatched := map[string]struct{}{}
	for name := range buildToolDispatcher(client, mapper, picklist) {
		dispatched[name] = struct{}{}
	}

	// The lazy discovery/routing set.
	categorized := map[string]struct{}{}
	for _, cat := range ToolCategories {
		for _, name := range cat.Tools {
			categorized[name] = struct{}{}
		}
	}

	for name := range registered {
		if _, ok := dispatched[name]; !ok {
			t.Errorf("tool %q is registered by RegisterAll but missing from buildToolDispatcher (uncallable via autotask_execute_tool in lazy mode)", name)
		}
		if _, ok := categorized[name]; !ok {
			t.Errorf("tool %q is registered by RegisterAll but missing from ToolCategories (undiscoverable in lazy mode)", name)
		}
	}
	for name := range dispatched {
		if _, ok := registered[name]; !ok {
			t.Errorf("tool %q is in buildToolDispatcher but not registered by RegisterAll", name)
		}
	}
	for name := range categorized {
		if _, ok := registered[name]; !ok {
			t.Errorf("tool %q is in ToolCategories but not registered by RegisterAll", name)
		}
	}
}
