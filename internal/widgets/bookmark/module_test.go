package bookmark

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule()
	if m.Type() != "bookmark" {
		t.Errorf("Type() = %q, want %q", m.Type(), "bookmark")
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

func TestModule_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "https url", config: `{"url":"https://example.com","title":"x"}`},
		{name: "relative path", config: `{"url":"/abc123"}`},
		{name: "empty url", config: `{"url":"","title":"x"}`},
		{name: "empty config", config: ``},
		{name: "null config", config: `null`},
		{name: "javascript url", config: `{"url":"javascript:alert(1)"}`, wantErr: "bookmark url must start with"},
		{name: "protocol relative", config: `{"url":"//evil.com"}`, wantErr: "bookmark url must start with"},
		{name: "wrong type", config: `{"url":5}`, wantErr: "bookmark config"},
	}
	m := NewModule()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := m.ValidateConfig(json.RawMessage(tc.config))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

var (
	_ widgets.Module          = (*Module)(nil)
	_ widgets.ConfigValidator = (*Module)(nil)
)
