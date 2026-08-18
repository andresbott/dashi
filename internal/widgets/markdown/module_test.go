package markdown

import (
	"context"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil)
	if m.Type() != "markdown" {
		t.Errorf("Type() = %q, want %q", m.Type(), "markdown")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	if NewModule(nil).Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	// The widget no longer owns any routes — RegisterRoutes is inherited
	// from NoopModule as a no-op. This just confirms it doesn't panic.
	m := NewModule(nil)
	m.RegisterRoutes(mux.NewRouter())
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(nil)
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
