package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

// TestRegisterSalesTools_NoPanic verifies registration does not panic.
func TestRegisterSalesTools_NoPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	RegisterSalesTools(s, client)
}

// TestGetProductHandler_NotFound tests that a missing product returns an error result.
func TestGetProductHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getProductHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetProductInput{ProductID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing product")
	}
}

// TestSearchProductsHandler_NoResults tests the empty-result case.
func TestSearchProductsHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchProductsHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchProductsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestGetServiceHandler_NotFound tests that a missing service returns an error result.
func TestGetServiceHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getServiceHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetServiceInput{ServiceID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing service")
	}
}

// TestSearchServicesHandler_NoResults tests the empty-result case.
func TestSearchServicesHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchServicesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchServicesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestGetServiceBundleHandler_NotFound tests that a missing service bundle returns an error result.
func TestGetServiceBundleHandler_NotFound(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := getServiceBundleHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, GetServiceBundleInput{ServiceBundleID: 99999})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true for missing service bundle")
	}
}

// TestSearchServiceBundlesHandler_NoResults tests the empty-result case.
func TestSearchServiceBundlesHandler_NoResults(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	handler := searchServiceBundlesHandler(client)
	ctx := context.Background()

	result, _, err := handler(ctx, nil, SearchServiceBundlesInput{SearchTerm: "acme"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
