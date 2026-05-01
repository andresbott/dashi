package battery

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule()
	if m.Type() != "battery" {
		t.Errorf("Type() = %q, want %q", m.Type(), "battery")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	m := NewModule()
	if m.Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	m := NewModule()
	r := mux.NewRouter()
	m.RegisterRoutes(r)
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule()
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
