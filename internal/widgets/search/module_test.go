package search

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule()
	if m.Type() != "search" {
		t.Errorf("Type() = %q, want %q", m.Type(), "search")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	m := NewModule()
	m.RegisterRoutes(mux.NewRouter())
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule()
	m.Warmup(context.Background(), nil)
}

var (
	_ widgets.Module            = (*Module)(nil)
	_ widgets.BrowserRenderable = (*Module)(nil)
)
