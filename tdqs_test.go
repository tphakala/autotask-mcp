package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
)

// TestTDQS_ToolDefinitionsQuality asserts, across every registered tool, the
// invariants that keep the server's Tool Definition Quality Score in tier A:
// a substantive description, annotations carrying a title, a trailing
// read/write marker consistent with the readOnly hint, every input parameter
// carrying a schema description, and no em or en dashes. It exercises the real
// tools/list path over the in-memory transport, so it also confirms
// annotations and the input schema serialize end to end.
func TestTDQS_ToolDefinitionsQuality(t *testing.T) {
	ctx := context.Background()
	_, client := autotasktest.NewServer(t)
	cs := connectMCP(t, client)

	// em dash and en dash code points, built from integers to keep the literal
	// characters out of the source entirely.
	emDash, enDash := rune(0x2014), rune(0x2013)

	count := 0
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		count++
		name := tool.Name

		if len(strings.TrimSpace(tool.Description)) < 40 {
			t.Errorf("%s: description too short to be specific: %q", name, tool.Description)
		}
		if tool.Annotations == nil {
			t.Errorf("%s: missing annotations", name)
			continue
		}
		if strings.TrimSpace(tool.Annotations.Title) == "" {
			t.Errorf("%s: missing annotation title", name)
		}

		// No em dashes or en dashes anywhere (project rule).
		for _, s := range []string{tool.Description, tool.Annotations.Title} {
			if strings.ContainsRune(s, emDash) || strings.ContainsRune(s, enDash) {
				t.Errorf("%s: contains em/en dash: %q", name, s)
			}
		}

		// Parameter Semantics (TDQS): every declared input parameter must carry a
		// description, so the schema conveys meaning beyond the bare field name.
		// InputSchema arrives as decoded JSON over the tools/list path, so parse
		// it generically rather than depending on the server-side struct type.
		var schema struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		}
		if raw, err := json.Marshal(tool.InputSchema); err != nil {
			t.Errorf("%s: cannot marshal input schema: %v", name, err)
		} else if err := json.Unmarshal(raw, &schema); err != nil {
			t.Errorf("%s: cannot parse input schema: %v", name, err)
		} else {
			for pname, p := range schema.Properties {
				if strings.TrimSpace(p.Description) == "" {
					t.Errorf("%s: parameter %q lacks a schema description", name, pname)
				}
			}
		}

		// The trailing marker must match the readOnly hint, so the description
		// never contradicts the annotation (a TDQS hard-gate failure).
		readOnly := tool.Annotations.ReadOnlyHint
		endsReadOnly := strings.HasSuffix(tool.Description, "Read-only.")
		endsWrites := strings.HasSuffix(tool.Description, "Writes to Autotask.")
		switch {
		case !endsReadOnly && !endsWrites:
			t.Errorf("%s: description missing Read-only./Writes to Autotask. marker", name)
		case readOnly && endsWrites:
			t.Errorf("%s: readOnly tool claims Writes to Autotask", name)
		case !readOnly && endsReadOnly:
			t.Errorf("%s: write tool claims Read-only", name)
		}
	}

	if count < 59 {
		t.Errorf("expected at least 59 registered tools, got %d", count)
	}
}
