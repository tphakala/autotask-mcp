package main

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

func TestBuildServer(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := buildServer(client, "", false)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestBuildServer_LazyLoading(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := buildServer(client, "", true)
	if s == nil {
		t.Fatal("expected non-nil server in lazy loading mode")
	}
}

// connectInMemory wires an in-memory MCP client to the given server.
func connectInMemory(t *testing.T, s *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := t.Context()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := s.Connect(ctx, serverTransport, nil)
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

// TestBuildServer_PromptsRegisteredInBothModes verifies that the PSA workflow
// prompts register unconditionally, including in lazy-loading mode where only
// meta-tools are exposed. In lazy mode the direct tools the prompts name are not
// registered, so the guidance routes through autotask_execute_tool.
func TestBuildServer_PromptsRegisteredInBothModes(t *testing.T) {
	for _, lazy := range []bool{false, true} {
		name := "full"
		if lazy {
			name = "lazy"
		}
		t.Run(name, func(t *testing.T) {
			verifyServerPromptsAndTools(t, lazy)
		})
	}
}

func verifyServerPromptsAndTools(t *testing.T, lazy bool) {
	t.Helper()
	_, client := autotasktest.NewServer(t)
	cs := connectInMemory(t, buildServer(client, "", lazy))

	verifyPromptCount(t, cs, lazy)
	verifyToolExposure(t, cs, lazy)
}

func verifyPromptCount(t *testing.T, cs *mcp.ClientSession, lazy bool) {
	t.Helper()
	pr, err := cs.ListPrompts(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	if len(pr.Prompts) != 4 {
		t.Fatalf("expected 4 prompts (lazy=%v), got %d", lazy, len(pr.Prompts))
	}
}

func verifyToolExposure(t *testing.T, cs *mcp.ClientSession, lazy bool) {
	t.Helper()
	tr, err := cs.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := make(map[string]bool, len(tr.Tools))
	for _, tool := range tr.Tools {
		names[tool.Name] = true
	}
	if lazy {
		if !names["autotask_execute_tool"] {
			t.Errorf("lazy mode should expose autotask_execute_tool meta-tool; tools: %v", names)
		}
		if names["autotask_get_ticket_details"] {
			t.Errorf("lazy mode should not register direct tool autotask_get_ticket_details")
		}
	} else if !names["autotask_get_ticket_details"] {
		t.Errorf("full mode should register direct tool autotask_get_ticket_details")
	}
}
