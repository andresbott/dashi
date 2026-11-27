package search

import (
	"encoding/json"
	"os"
	"testing"
)

// The engine list is shared with the editor: webui/src/widgets/search/
// types.test.ts reads the same fixture, so a label or URL changed on one
// side only fails a test instead of drifting.
func TestEnginesMatchSharedFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/engines.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture struct {
		Default string `json:"default"`
		Engines []struct {
			Engine string `json:"engine"`
			Label  string `json:"label"`
			Action string `json:"action"`
		} `json:"engines"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	if fixture.Default != defaultEngine {
		t.Errorf("default engine = %q, fixture says %q", defaultEngine, fixture.Default)
	}
	if len(fixture.Engines) != len(engines) {
		t.Errorf("Go allowlists %d engines, fixture lists %d", len(engines), len(fixture.Engines))
	}
	for _, want := range fixture.Engines {
		got, ok := engines[want.Engine]
		if !ok {
			t.Errorf("engine %q missing from the Go allowlist", want.Engine)
			continue
		}
		if got.Label != want.Label || got.Action != want.Action {
			t.Errorf("engine %q = %+v, fixture wants label %q action %q", want.Engine, got, want.Label, want.Action)
		}
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   view
	}{
		{
			name:   "no config uses google and the default placeholder",
			config: ``,
			want:   view{Action: "https://www.google.com/search", Label: "Google", Placeholder: "Search..."},
		},
		{
			name:   "empty object uses the defaults",
			config: `{}`,
			want:   view{Action: "https://www.google.com/search", Label: "Google", Placeholder: "Search..."},
		},
		{
			name:   "duckduckgo with a custom placeholder",
			config: `{"engine":"duckduckgo","placeholder":"Search the web..."}`,
			want:   view{Action: "https://duckduckgo.com/", Label: "DuckDuckGo", Placeholder: "Search the web..."},
		},
		{
			name:   "bing",
			config: `{"engine":"bing"}`,
			want:   view{Action: "https://www.bing.com/search", Label: "Bing", Placeholder: "Search..."},
		},
		{
			name:   "unknown engine falls back to google",
			config: `{"engine":"https://evil.example/collect"}`,
			want:   view{Action: "https://www.google.com/search", Label: "Google", Placeholder: "Search..."},
		},
		{
			name:   "engine names are exact, not case-folded",
			config: `{"engine":"Bing"}`,
			want:   view{Action: "https://www.google.com/search", Label: "Google", Placeholder: "Search..."},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolve(json.RawMessage(tc.config))
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if got != tc.want {
				t.Errorf("resolve() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestResolveRejectsMalformedConfig(t *testing.T) {
	if _, err := resolve(json.RawMessage(`{"engine":`)); err == nil {
		t.Fatal("expected an error for malformed config")
	}
}
