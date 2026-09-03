package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupPromptTest wires a full in-memory MCP server with all prompts registered
// and connected to an MCP client.
func setupPromptTest(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "prompts-test", Version: "v0.0.1"}, nil)
	RegisterAll(server)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

// promptText concatenates the text of every message in a prompt result.
func promptText(t *testing.T, res *mcp.GetPromptResult) string {
	t.Helper()
	if len(res.Messages) == 0 {
		t.Fatal("expected at least one prompt message")
	}
	var b strings.Builder
	for _, m := range res.Messages {
		tc, ok := m.Content.(*mcp.TextContent)
		if !ok {
			t.Fatalf("expected *mcp.TextContent, got %T", m.Content)
		}
		if m.Role != "user" {
			t.Errorf("expected user role, got %q", m.Role)
		}
		b.WriteString(tc.Text)
	}
	return b.String()
}

func TestListPrompts_RegistersAllFour(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.ListPrompts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	want := map[string]int{ // name -> required argument count
		"autotask_triage_ticket":           0,
		"autotask_summarize_ticket":        1,
		"autotask_draft_time_entry":        3,
		"autotask_weekly_timesheet_review": 3,
	}
	if len(res.Prompts) != len(want) {
		t.Fatalf("expected %d prompts, got %d: %+v", len(want), len(res.Prompts), res.Prompts)
	}
	for _, p := range res.Prompts {
		reqCount, known := want[p.Name]
		if !known {
			t.Errorf("unexpected prompt %q", p.Name)
			continue
		}
		if p.Description == "" {
			t.Errorf("prompt %q has empty description", p.Name)
		}
		got := 0
		for _, a := range p.Arguments {
			if a.Required {
				got++
			}
			if a.Description == "" {
				t.Errorf("prompt %q argument %q has empty description", p.Name, a.Name)
			}
		}
		if got != reqCount {
			t.Errorf("prompt %q: expected %d required args, got %d", p.Name, reqCount, got)
		}
	}
}

func TestGetPrompt_TriageByID(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_triage_ticket",
		Arguments: map[string]string{"ticketId": "778899"},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	for _, want := range []string{"778899", "autotask_get_ticket_details", "autotask_search_companies", "autotask_list_queues", "autotask_update_ticket"} {
		if !strings.Contains(text, want) {
			t.Errorf("triage-by-id text missing %q; got:\n%s", want, text)
		}
	}
}

func TestGetPrompt_TriageByDescription(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_triage_ticket",
		Arguments: map[string]string{"description": "VPN keeps dropping every hour"},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	if !strings.Contains(text, "VPN keeps dropping every hour") {
		t.Errorf("triage-by-description text missing the description; got:\n%s", text)
	}
	if !strings.Contains(text, "autotask_search_tickets") {
		t.Errorf("triage-by-description should point at autotask_search_tickets; got:\n%s", text)
	}
}

func TestGetPrompt_TriageRequiresOneArgument(t *testing.T) {
	cs := setupPromptTest(t)
	_, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_triage_ticket",
		Arguments: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error when neither ticketId nor description is provided")
	}
}

func TestGetPrompt_SummarizeEmbedsArgsAndTools(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_summarize_ticket",
		Arguments: map[string]string{"ticketId": "424242"},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	for _, want := range []string{"424242", "autotask_get_ticket_details", "autotask_search_ticket_notes", "autotask_search_time_entries"} {
		if !strings.Contains(text, want) {
			t.Errorf("summarize text missing %q; got:\n%s", want, text)
		}
	}
}

func TestGetPrompt_SummarizeMissingRequiredArg(t *testing.T) {
	cs := setupPromptTest(t)
	_, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_summarize_ticket",
		Arguments: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error when required ticketId is missing")
	}
}

func TestGetPrompt_DraftTimeEntryEmbedsAllInputs(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "autotask_draft_time_entry",
		Arguments: map[string]string{
			"ticketId":    "5150",
			"hoursWorked": "2.5",
			"summary":     "Replaced faulty switch and verified uplinks",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	for _, want := range []string{"5150", "2.5", "Replaced faulty switch and verified uplinks", "autotask_search_resources", "autotask_create_time_entry"} {
		if !strings.Contains(text, want) {
			t.Errorf("draft time entry text missing %q; got:\n%s", want, text)
		}
	}
}

func TestGetPrompt_DraftTimeEntryMissingArg(t *testing.T) {
	cs := setupPromptTest(t)
	_, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_draft_time_entry",
		Arguments: map[string]string{"ticketId": "5150", "hoursWorked": "2.5"}, // summary missing
	})
	if err == nil {
		t.Fatal("expected error when required summary is missing")
	}
}

func TestGetPrompt_WeeklyTimesheetEmbedsRange(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "autotask_weekly_timesheet_review",
		Arguments: map[string]string{
			"resourceId": "312",
			"startDate":  "2026-08-24",
			"endDate":    "2026-08-30",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	for _, want := range []string{"312", "2026-08-24", "2026-08-30", "autotask_search_time_entries"} {
		if !strings.Contains(text, want) {
			t.Errorf("weekly timesheet text missing %q; got:\n%s", want, text)
		}
	}
}

// TestPrompts_ReferenceOnlyRealCompanyTool guards against the model being told to
// use a non-existent autotask_get_company tool: the guidance must steer to
// autotask_search_companies instead.
func TestPrompts_ReferenceOnlyRealCompanyTool(t *testing.T) {
	cs := setupPromptTest(t)
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      "autotask_triage_ticket",
		Arguments: map[string]string{"ticketId": "1"},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	text := promptText(t, res)
	if !strings.Contains(text, "autotask_search_companies") {
		t.Errorf("expected guidance to use autotask_search_companies; got:\n%s", text)
	}
}
