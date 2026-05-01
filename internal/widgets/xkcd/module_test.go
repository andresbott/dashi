package xkcd

import (
	"context"
	"net/http/httptest"
	"testing"

	xkcdclient "github.com/andresbott/dashi/internal/xkcd"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil, nil)
	if m.Type() != "xkcd" {
		t.Errorf("Type() = %q, want %q", m.Type(), "xkcd")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	m := NewModule(&xkcdclient.Client{}, nil)
	if m.Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_MountsGetComic(t *testing.T) {
	m := NewModule(&xkcdclient.Client{}, nil)
	r := mux.NewRouter()
	m.RegisterRoutes(r)

	// Assert the route exists by matching the path.
	req := httptest.NewRequest("GET", "/widgets/xkcd", nil)
	var match mux.RouteMatch
	if !r.Match(req, &match) {
		t.Fatal("expected /widgets/xkcd route to be registered")
	}
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(nil, nil)
	m.Warmup(context.Background(), nil) // no-op; inherits NoopModule
}

var _ widgets.Module = (*Module)(nil)
