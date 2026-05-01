package search

import (
	"context"
	"encoding/json"
	"strings"
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

func TestModule_Renderer_EmitsPlaceholder(t *testing.T) {
	m := NewModule()
	got, err := m.Renderer()(json.RawMessage(`{}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), "widget-search-placeholder") {
		t.Errorf("expected placeholder HTML, got %q", got)
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

var _ widgets.Module = (*Module)(nil)
