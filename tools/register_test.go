package tools

import (
	"testing"
)

func TestDefaultPageSize_Default(t *testing.T) {
	// When requested <= 0, use defaultVal
	got := defaultPageSize(0, 25, 100)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}

	got = defaultPageSize(-5, 25, 100)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}
}

func TestDefaultPageSize_Clamped(t *testing.T) {
	// When requested > maxVal, clamp to maxVal
	got := defaultPageSize(500, 25, 100)
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestDefaultPageSize_Valid(t *testing.T) {
	// When requested is within bounds, return it
	got := defaultPageSize(50, 25, 100)
	if got != 50 {
		t.Errorf("expected 50, got %d", got)
	}
}

func TestDefaultPageSize_MinOne(t *testing.T) {
	// defaultVal=0 and maxVal=0 edge case: result should be at least 1
	got := defaultPageSize(0, 0, 0)
	if got != 1 {
		t.Errorf("expected minimum 1, got %d", got)
	}
}

func TestDefaultPageSize_ExactMax(t *testing.T) {
	// Exactly at max boundary
	got := defaultPageSize(100, 25, 100)
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestDefaultPage_Default(t *testing.T) {
	// page < 1 should return 1
	if got := defaultPage(0); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := defaultPage(-1); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestDefaultPage_Valid(t *testing.T) {
	if got := defaultPage(1); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := defaultPage(5); got != 5 {
		t.Errorf("expected 5, got %d", got)
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
	} else if v != "Test" {
		t.Errorf("expected name=Test, got %v", v)
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
