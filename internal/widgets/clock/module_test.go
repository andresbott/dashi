package clock

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule()
	if m.Type() != "clock" {
		t.Errorf("Type() = %q, want %q", m.Type(), "clock")
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
	m.RegisterRoutes(r) // inherits NoopModule; must not panic
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule()
	m.Warmup(context.Background(), nil) // inherits NoopModule; must not panic
}

// Confirm the module implements the interface at compile time.
var _ widgets.Module = (*Module)(nil)
