package stack

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	if m.Type() != "stack" {
		t.Errorf("Type() = %q, want %q", m.Type(), "stack")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	if m.Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	m.RegisterRoutes(mux.NewRouter())
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
