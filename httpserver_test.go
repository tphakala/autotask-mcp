package main

import (
	"net/http"
	"testing"
	"time"
)

// TestNewMCPHTTPServer_NoWriteTimeout pins the fix for issue #51: the MCP HTTP server
// must not carry a WriteTimeout, since that would forcibly close long-lived SSE streams.
// Slowloris protection is provided by ReadHeaderTimeout and IdleTimeout instead.
func TestNewMCPHTTPServer_NoWriteTimeout(t *testing.T) {
	srv := newMCPHTTPServer("127.0.0.1:0", http.NewServeMux())

	if srv.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %v, want 0 so long-lived SSE streams are not aborted (issue #51)", srv.WriteTimeout)
	}
	if srv.ReadTimeout != 0 {
		t.Errorf("ReadTimeout = %v, want 0; a full-request read deadline is intentionally not set (ReadHeaderTimeout guards slowloris)", srv.ReadTimeout)
	}
	if srv.ReadHeaderTimeout != 30*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 30s (slowloris protection)", srv.ReadHeaderTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("IdleTimeout = %v, want 120s", srv.IdleTimeout)
	}
	if srv.Addr != "127.0.0.1:0" {
		t.Errorf("Addr = %q, want 127.0.0.1:0", srv.Addr)
	}
	if srv.Handler == nil {
		t.Error("Handler must be set")
	}
}
