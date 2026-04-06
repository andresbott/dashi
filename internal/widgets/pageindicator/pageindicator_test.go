package pageindicator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestNewStaticRenderer_SinglePage(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  0,
		TotalPages: 1,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `<div class="widget-page-indicator">`) {
		t.Errorf("expected page indicator div, got: %s", html)
	}

	// Should have exactly one dot
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 1 {
		t.Errorf("expected 1 dot for single page, got %d", dotCount)
	}

	// The single dot should be active
	if !strings.Contains(html, `class="dot active"`) {
		t.Errorf("expected single dot to be active, got: %s", html)
	}
}

func TestNewStaticRenderer_MultiplePages_FirstActive(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  0,
		TotalPages: 3,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 3 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 3 {
		t.Errorf("expected 3 dots, got %d", dotCount)
	}

	// Should have exactly 1 active dot
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 1 {
		t.Errorf("expected 1 active dot, got %d", activeCount)
	}

	// Should have 2 inactive dots
	inactiveCount := strings.Count(html, `class="dot"`)
	if inactiveCount != 2 {
		t.Errorf("expected 2 inactive dots, got %d", inactiveCount)
	}
}

func TestNewStaticRenderer_MultiplePages_MiddleActive(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  1,
		TotalPages: 3,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 3 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 3 {
		t.Errorf("expected 3 dots, got %d", dotCount)
	}

	// Should have exactly 1 active dot
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 1 {
		t.Errorf("expected 1 active dot, got %d", activeCount)
	}
}

func TestNewStaticRenderer_MultiplePages_LastActive(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  4,
		TotalPages: 5,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 5 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 5 {
		t.Errorf("expected 5 dots, got %d", dotCount)
	}

	// Should have exactly 1 active dot
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 1 {
		t.Errorf("expected 1 active dot, got %d", activeCount)
	}
}

func TestNewStaticRenderer_ZeroPages_DefaultsToOne(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  0,
		TotalPages: 0, // Invalid, should default to 1
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 1 dot when TotalPages is 0
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 1 {
		t.Errorf("expected 1 dot when TotalPages=0, got %d", dotCount)
	}

	// The single dot should be active
	if !strings.Contains(html, `class="dot active"`) {
		t.Errorf("expected single dot to be active, got: %s", html)
	}
}

func TestNewStaticRenderer_NegativePages_DefaultsToOne(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  0,
		TotalPages: -5, // Invalid, should default to 1
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 1 dot when TotalPages is negative
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 1 {
		t.Errorf("expected 1 dot when TotalPages is negative, got %d", dotCount)
	}
}

func TestNewStaticRenderer_PageIndexOutOfBounds_NoActiveDot(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  10, // Out of bounds
		TotalPages: 3,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 3 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 3 {
		t.Errorf("expected 3 dots, got %d", dotCount)
	}

	// Should have 0 active dots (all inactive)
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 0 {
		t.Errorf("expected 0 active dots when PageIndex out of bounds, got %d", activeCount)
	}

	// All dots should be inactive
	inactiveCount := strings.Count(html, `class="dot"`)
	if inactiveCount != 3 {
		t.Errorf("expected 3 inactive dots, got %d", inactiveCount)
	}
}

func TestNewStaticRenderer_NegativePageIndex_NoActiveDot(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  -1, // Negative index
		TotalPages: 3,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 3 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 3 {
		t.Errorf("expected 3 dots, got %d", dotCount)
	}

	// Should have 0 active dots
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 0 {
		t.Errorf("expected 0 active dots when PageIndex is negative, got %d", activeCount)
	}
}

func TestNewStaticRenderer_ManyPages(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  5,
		TotalPages: 10,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should have exactly 10 dots
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 10 {
		t.Errorf("expected 10 dots, got %d", dotCount)
	}

	// Should have exactly 1 active dot
	activeCount := strings.Count(html, `class="dot active"`)
	if activeCount != 1 {
		t.Errorf("expected 1 active dot, got %d", activeCount)
	}
}

func TestNewStaticRenderer_IgnoresConfig(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  0,
		TotalPages: 2,
	}

	// Config is ignored, so arbitrary JSON should work
	config := json.RawMessage(`{"foo": "bar", "baz": 123}`)
	got, err := renderer(config, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Should still render correctly
	dotCount := strings.Count(html, `<span class="dot`)
	if dotCount != 2 {
		t.Errorf("expected 2 dots regardless of config, got %d", dotCount)
	}
}

func TestNewStaticRenderer_ValidHTMLStructure(t *testing.T) {
	renderer := NewStaticRenderer()
	ctx := widgets.RenderContext{
		PageIndex:  1,
		TotalPages: 3,
	}

	got, err := renderer(json.RawMessage(`{}`), ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Check for proper opening and closing tags
	if !strings.HasPrefix(strings.TrimSpace(html), `<div class="widget-page-indicator">`) {
		t.Errorf("expected HTML to start with widget-page-indicator div, got: %s", html)
	}

	if !strings.HasSuffix(strings.TrimSpace(html), `</div>`) {
		t.Errorf("expected HTML to end with closing div tag, got: %s", html)
	}

	// Check that spans are properly closed
	openSpans := strings.Count(html, "<span")
	closeSpans := strings.Count(html, "</span>")
	if openSpans != closeSpans {
		t.Errorf("mismatched span tags: %d open, %d close", openSpans, closeSpans)
	}
}
