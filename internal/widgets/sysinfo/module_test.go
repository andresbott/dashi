package sysinfo

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil)
	if m.Type() != "sysinfo" {
		t.Errorf("Type() = %q, want %q", m.Type(), "sysinfo")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	if NewModule(nil).Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_MountsGetSysinfo(t *testing.T) {
	m := NewModule(nil)
	r := mux.NewRouter()
	m.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/widgets/sysinfo", nil)
	var match mux.RouteMatch
	if !r.Match(req, &match) {
		t.Fatal("expected /widgets/sysinfo route to be registered")
	}
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(nil)
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
