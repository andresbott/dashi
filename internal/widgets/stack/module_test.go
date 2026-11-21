package stack

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

func TestModule_Type(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	if m.Type() != "stack" {
		t.Errorf("Type() = %q, want %q", m.Type(), "stack")
	}
}

func TestModule_RendererNotNil(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	if m.Renderer() == nil {
		t.Fatal("Renderer() is nil")
	}
}

func TestModule_RegisterRoutes_NoPanic(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	m.RegisterRoutes(mux.NewRouter())
}

func TestModule_Warmup_NoPanic(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	m.Warmup(context.Background(), nil)
}

// A stack delegates to its children's validators, including stacks nested
// in stacks, so wrapping a widget in a stack cannot skip its rule.
func TestModule_ValidateConfig(t *testing.T) {
	registry := widgets.NewRegistry()
	m := NewModule(registry)
	registry.RegisterValidator("stack", m.ValidateConfig)
	registry.RegisterValidator("strict", func(config json.RawMessage) error {
		if string(config) == `"bad"` {
			return errors.New("strict says no")
		}
		return nil
	})

	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "empty config", config: ``},
		{name: "no children", config: `{"widgets":[]}`},
		{name: "valid child", config: `{"widgets":[{"id":"c1","type":"strict","config":"ok"}]}`},
		{name: "child without validator", config: `{"widgets":[{"id":"c1","type":"clock","config":"bad"}]}`},
		{
			name:    "invalid child",
			config:  `{"widgets":[{"id":"c1","type":"strict","config":"ok"},{"id":"c2","type":"strict","config":"bad"}]}`,
			wantErr: `stack child "c2" (strict): strict says no`,
		},
		{
			name:    "invalid grandchild",
			config:  `{"widgets":[{"id":"inner","type":"stack","config":{"widgets":[{"id":"g1","type":"strict","config":"bad"}]}}]}`,
			wantErr: `stack child "inner" (stack): stack child "g1" (strict): strict says no`,
		},
		{name: "malformed", config: `{"widgets":5}`, wantErr: "stack config"},
	}
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
