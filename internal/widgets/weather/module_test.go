package weather

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	weatherpkg "github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_TypesDistinct(t *testing.T) {
	m := NewModule(nil, nil, nil)
	mc := NewCompactModule(nil, nil, nil)
	if m.Type() == mc.Type() {
		t.Fatalf("weather and weather-compact must have different type strings, both are %q", m.Type())
	}
	if m.Type() != "weather" || mc.Type() != "weather-compact" {
		t.Errorf("unexpected types: %q / %q", m.Type(), mc.Type())
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	client := weatherpkg.NewClient(nil)
	store := themes.NewStore("")
	if NewModule(client, store, nil).Renderer() == nil {
		t.Fatal("weather renderer nil")
	}
	if NewCompactModule(client, store, nil).Renderer() == nil {
		t.Fatal("weather-compact renderer nil")
	}
}

func TestModule_RegisterRoutes_MountsBothEndpoints(t *testing.T) {
	m := NewModule(weatherpkg.NewClient(nil), themes.NewStore(""), nil)
	r := mux.NewRouter()
	m.RegisterRoutes(r)

	for _, path := range []string{"/widgets/weather", "/widgets/weather/geocode"} {
		req := httptest.NewRequest("GET", path, nil)
		var match mux.RouteMatch
		if !r.Match(req, &match) {
			t.Errorf("expected %s to be registered", path)
		}
	}
}

func TestModule_CompactRegisterRoutes_NoOp(t *testing.T) {
	// weather-compact does not mount any HTTP routes — the weather module
	// owns them. Confirm Compact does not register /widgets/weather again,
	// which would cause a duplicate-route panic when both modules run.
	mc := NewCompactModule(weatherpkg.NewClient(nil), themes.NewStore(""), nil)
	r := mux.NewRouter()
	mc.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/widgets/weather", nil)
	var match mux.RouteMatch
	if r.Match(req, &match) {
		t.Error("weather-compact must not mount /widgets/weather")
	}
}

func TestModule_Warmup_ExtractsLocations(t *testing.T) {
	m := NewModule(weatherpkg.NewClient(nil), themes.NewStore(""), nil)
	cfgs := []json.RawMessage{
		json.RawMessage(`{"latitude":1.5,"longitude":2.5}`),
		json.RawMessage(`{"latitude":0,"longitude":0}`),     // skipped (both zero)
		json.RawMessage(`not json`),                         // skipped
		json.RawMessage(`{"latitude":3.5,"longitude":4.5}`),
	}
	// Warmup is fire-and-forget; we only assert it doesn't panic and
	// returns in reasonable time. A deeper test would stub the client.
	m.Warmup(context.Background(), cfgs)
}

var (
	_ widgets.Module = (*Module)(nil)
	_ widgets.Module = (*CompactModule)(nil)
)
