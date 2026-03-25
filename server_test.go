package main

import (
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
)

func TestBuildServer(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := buildServer(client, false)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestBuildServer_LazyLoading(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := buildServer(client, true)
	if s == nil {
		t.Fatal("expected non-nil server in lazy loading mode")
	}
}
