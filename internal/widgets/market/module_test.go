package market

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	marketpkg "github.com/andresbott/dashi/internal/providers/market"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(nil, nil)
	if m.Type() != "market" {
		t.Errorf("Type() = %q, want %q", m.Type(), "market")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	if NewModule(marketpkg.NewClient(nil), nil).Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_MountsMarketEndpoint(t *testing.T) {
	m := NewModule(nil, nil)
	r := mux.NewRouter()
	m.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/widgets/market", nil)
	var match mux.RouteMatch
	if !r.Match(req, &match) {
		t.Fatal("expected /widgets/market route to be registered")
	}
}

func TestModule_ExtractTargets(t *testing.T) {
	cfgs := []json.RawMessage{
		json.RawMessage(`{"symbol":"AAPL","range":"3mo"}`),
		json.RawMessage(`{"symbol":"","range":"1mo"}`),     // skipped (empty symbol)
		json.RawMessage(`not json`),                         // skipped
		json.RawMessage(`{"symbol":"MSFT"}`),                // default range
	}
	got := extractTargets(cfgs)
	if len(got) != 2 {
		t.Fatalf("expected 2 targets, got %d: %v", len(got), got)
	}
	if got[0].Symbol != "AAPL" || got[0].Range != "3mo" {
		t.Errorf("unexpected first target: %+v", got[0])
	}
	if got[1].Symbol != "MSFT" || got[1].Range != "1mo" {
		t.Errorf("expected MSFT with default range, got %+v", got[1])
	}
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(marketpkg.NewClient(nil), nil)
	m.Warmup(context.Background(), nil)
}

var _ widgets.Module = (*Module)(nil)
