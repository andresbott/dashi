package markdown

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil, nil)
	if m.Type() != "markdown" {
		t.Errorf("Type() = %q, want %q", m.Type(), "markdown")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	if NewModule(nil, nil).Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_MountsBothEndpoints(t *testing.T) {
	m := NewModule(nil, nil)
	r := mux.NewRouter()
	m.RegisterRoutes(r)

	for _, path := range []string{"/dashboards/abc/markdown", "/dashboards/abc/markdown/readme.md"} {
		req := httptest.NewRequest("GET", path, nil)
		var match mux.RouteMatch
		if !r.Match(req, &match) {
			t.Errorf("expected %s to be registered", path)
		}
	}
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(nil, nil)
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
