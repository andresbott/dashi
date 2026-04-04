package battery

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderStatic_ValidValue(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{QueryParams: map[string]string{"battery": "75"}}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "75%") {
		t.Errorf("expected '75%%' in output, got: %s", html)
	}
}

func TestRenderStatic_ClampAbove100(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{QueryParams: map[string]string{"battery": "150"}}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "100%") {
		t.Errorf("expected '100%%' in output, got: %s", html)
	}
}

func TestRenderStatic_ClampBelow0(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{QueryParams: map[string]string{"battery": "-10"}}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "0%") {
		t.Errorf("expected '0%%' in output, got: %s", html)
	}
}

func TestRenderStatic_MissingParam(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{QueryParams: map[string]string{}}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "0%") {
		t.Errorf("expected '0%%' for missing param, got: %s", html)
	}
}

func TestRenderStatic_InvalidParam(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{QueryParams: map[string]string{"battery": "abc"}}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "0%") {
		t.Errorf("expected '0%%' for invalid param, got: %s", html)
	}
}

func TestRenderStatic_NilQueryParams(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{}
	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "0%") {
		t.Errorf("expected '0%%' for nil query params, got: %s", html)
	}
}
