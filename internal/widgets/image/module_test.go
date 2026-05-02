package image

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil)
	if m.Type() != "image" {
		t.Errorf("Type() = %q, want %q", m.Type(), "image")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	store := dashboard.NewStore("")
	m := NewModule(store)
	if m.Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	m := NewModule(nil)
	m.RegisterRoutes(mux.NewRouter())
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(nil)
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
