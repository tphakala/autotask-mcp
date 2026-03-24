package resources

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tphakala/go-autotask/autotasktest"
)

func TestRegisterAll_DoesNotPanic(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)

	// Should not panic.
	RegisterAll(s, client)
}

func TestParseIDFromURI(t *testing.T) {
	tests := []struct {
		uri     string
		wantID  int64
		wantErr bool
	}{
		{"autotask://companies/123", 123, false},
		{"autotask://tickets/456", 456, false},
		{"autotask://contacts/1", 1, false},
		{"autotask://companies/", 0, true},
		{"autotask://companies", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got, err := parseIDFromURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIDFromURI(%q) error = %v, wantErr %v", tt.uri, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantID {
				t.Errorf("parseIDFromURI(%q) = %d, want %d", tt.uri, got, tt.wantID)
			}
		})
	}
}
