package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/autotasktest"
	"github.com/tphakala/autotask-mcp/services"
)

// newTestPicklist creates a PicklistCache backed by the provided client.
func newTestPicklist(client *autotask.Client) *services.PicklistCache {
	return services.NewPicklistCache(client)
}

// TestRegisterPicklistTools_NoPanic verifies registration does not panic.
func TestRegisterPicklistTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	picklist := newTestPicklist(client)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterPicklistTools(s, client, picklist)
}

// TestListQueuesHandler_ReturnsResult tests that the handler returns a result without panicking.
func TestListQueuesHandler_ReturnsResult(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "queueID", Label: "Queue", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	picklist := newTestPicklist(client)
	handler := listQueuesHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
}

// TestListTicketStatusesHandler_ReturnsResult tests that the handler returns a result.
func TestListTicketStatusesHandler_ReturnsResult(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "status", Label: "Status", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	picklist := newTestPicklist(client)
	handler := listTicketStatusesHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
}

// TestListTicketPrioritiesHandler_ReturnsResult tests that the handler returns a result.
func TestListTicketPrioritiesHandler_ReturnsResult(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntityMetadata(
			"Tickets",
			autotasktest.EntityInfoResponse{Name: "Tickets", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "priority", Label: "Priority", DataType: "integer", IsPickList: true},
			},
			nil,
		),
	)
	picklist := newTestPicklist(client)
	handler := listTicketPrioritiesHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
}

// TestGetFieldInfoHandler_AllFields tests that all fields for an entity are returned.
func TestGetFieldInfoHandler_AllFields(t *testing.T) {
	_, client := autotasktest.NewServer(t,
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
	picklist := newTestPicklist(client)
	handler := getFieldInfoHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetFieldInfoInput{EntityType: "Companies"})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestGetFieldInfoHandler_SpecificField tests filtering to a specific field.
func TestGetFieldInfoHandler_SpecificField(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntityMetadata(
			"Companies",
			autotasktest.EntityInfoResponse{Name: "Companies", CanCreate: true, CanQuery: true},
			[]autotasktest.FieldInfoResponse{
				{Name: "companyName", Label: "Company Name", DataType: "string"},
			},
			nil,
		),
	)
	picklist := newTestPicklist(client)
	handler := getFieldInfoHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetFieldInfoInput{EntityType: "Companies", FieldName: "companyName"})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Errorf("expected no error result, got IsError=true; content: %v", result.Content)
	}
}

// TestGetFieldInfoHandler_NotFoundField tests that a non-existent field returns a text result.
func TestGetFieldInfoHandler_NotFoundField(t *testing.T) {
	_, client := autotasktest.NewServer(t,
		autotasktest.WithEntityMetadata(
			"Companies",
			autotasktest.EntityInfoResponse{Name: "Companies"},
			[]autotasktest.FieldInfoResponse{
				{Name: "companyName", Label: "Company Name", DataType: "string"},
			},
			nil,
		),
	)
	picklist := newTestPicklist(client)
	handler := getFieldInfoHandler(picklist)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetFieldInfoInput{EntityType: "Companies", FieldName: "nonexistentField"})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected no error for not-found field (returns text message)")
	}
}
