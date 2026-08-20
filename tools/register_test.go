package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestTextResult(t *testing.T) {
	result, out, err := textResult("Hello %s", "world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil out, got %v", out)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestErrorResult(t *testing.T) {
	result, out, err := errorResult("something went wrong: %d", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil out, got %v", out)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestSearchResult(t *testing.T) {
	items := []map[string]any{
		{"id": float64(101), "companyName": "Acme Corp", "phone": "555-1234"},
		{"id": float64(102), "companyName": "Beta Inc", "phone": "555-5678"},
	}

	result, _, err := searchResult(context.Background(), nil, items, "autotask_search_companies", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatalf("expected valid non-error result: %v", result)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	summary, ok := resp["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary object: %v", resp)
	}
	if summary["returned"] != float64(2) {
		t.Errorf("expected returned=2, got %v", summary["returned"])
	}
	if summary["hasMore"] != false {
		t.Errorf("expected hasMore=false, got %v", summary["hasMore"])
	}
}
