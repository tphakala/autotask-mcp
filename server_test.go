package main

import (
	"testing"

	"github.com/tphakala/go-autotask/autotasktest"
)

func TestBuildServer(t *testing.T) {
	_, client := autotasktest.NewServer(t)
	s := buildServer(client)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}
