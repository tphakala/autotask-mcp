package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/autotask-mcp/services"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/go-autotask/metadata"
)

// TestRegisterPicklistTools_NoPanic verifies registration does not panic.
func TestRegisterPicklistTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	picklist := services.NewPicklistCache(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterPicklistTools(s, client, picklist)
}

// TestListQueuesHandler_ReturnsResult tests that list_queues returns structured list over wire.
func TestListQueuesHandler_ReturnsResult(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "queueID", Label: "Queue", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_list_queues",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected IsError=false; content: %v", result.Content)
	}

	values := parseStructuredContent[[]metadata.PickListValue](t, result)
	_ = values
}

// TestListTicketStatusesHandler_ReturnsResult tests that list_ticket_statuses returns structured list over wire.
func TestListTicketStatusesHandler_ReturnsResult(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "status", Label: "Status", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_list_ticket_statuses",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected IsError=false; content: %v", result.Content)
	}

	values := parseStructuredContent[[]metadata.PickListValue](t, result)
	_ = values
}

// TestListTicketPrioritiesHandler_ReturnsResult tests that list_ticket_priorities returns structured list over wire.
func TestListTicketPrioritiesHandler_ReturnsResult(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "priority", Label: "Priority", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "autotask_list_ticket_priorities",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected IsError=false; content: %v", result.Content)
	}

	values := parseStructuredContent[[]metadata.PickListValue](t, result)
	_ = values
}

// TestGetFieldInfoHandler_AllFields tests that all fields for an entity are returned over wire.
func TestGetFieldInfoHandler_AllFields(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Companies",
			autotasktest.EntityInfoResponse{Name: "Companies", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "companyName", Label: "Company Name", DataType: "string"},
				{Name: "isActive", Label: "Is Active", DataType: "boolean"},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_field_info",
		Arguments: map[string]any{
			"entityType": "Companies",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	fields := parseStructuredContent[[]metadata.FieldInfo](t, result)
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

// TestGetFieldInfoHandler_SpecificField tests filtering to a specific field over wire.
func TestGetFieldInfoHandler_SpecificField(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Companies",
			autotasktest.EntityInfoResponse{Name: "Companies", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "companyName", Label: "Company Name", DataType: "string"},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_field_info",
		Arguments: map[string]any{
			"entityType": "Companies",
			"fieldName":  "companyName",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}

	field := parseStructuredContent[metadata.FieldInfo](t, result)
	if field.Name != "companyName" {
		t.Errorf("expected fieldName='companyName', got %q", field.Name)
	}
}

// TestGetFieldInfoHandler_NotFoundField tests that a non-existent field returns an error over wire.
func TestGetFieldInfoHandler_NotFoundField(t *testing.T) {
	cs, _ := setupWireTest(t,
		autotasktest.WithEntityMetadata(
			"Companies",
			autotasktest.EntityInfoResponse{Name: "Companies"},
			[]autotasktest.FieldInfoResponse{
				{Name: "companyName", Label: "Company Name", DataType: "string"},
			},
			nil,
		),
	)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "autotask_get_field_info",
		Arguments: map[string]any{
			"entityType": "Companies",
			"fieldName":  "nonexistentField",
		},
	})
	if err != nil {
		t.Fatalf("unexpected wire protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for non-existent field")
	}
}
