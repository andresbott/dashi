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

func TestNewStaticRenderer_CentresDotsWithFlex(t *testing.T) {
	// litehtml has comprehensive flexbox support
	// (docs/project/litehtml-rendering-reference.md:143). These four
	// declarations are the pre-refactor ones and they decide the rendered PNG:
	// with text-align instead, the dots sit on a text baseline, which changes
	// both the widget's height and the dots' vertical position.
	renderer := NewStaticRenderer()
	got, err := renderer(json.RawMessage(`{}`), widgets.RenderContext{TotalPages: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)

	for _, decl := range []string{
		"display: flex",
		"align-items: center",
		"justify-content: center",
		"padding: 8px",
	} {
		if !strings.Contains(html, decl) {
			t.Errorf("page-indicator CSS missing %q (changes rendered PNGs):\n%s", decl, html)
		}
	}
	if strings.Contains(html, "text-align: center") {
		t.Errorf("text-align centring was the regression; flex centring is required:\n%s", html)
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

	// Check for proper opening and closing tags (style block is now prepended)
	if !strings.Contains(html, `<style>`) {
		t.Errorf("expected HTML to contain style block, got: %s", html)
	}
	if !strings.Contains(html, `<div class="widget-page-indicator">`) {
		t.Errorf("expected HTML to contain widget-page-indicator div, got: %s", html)
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
