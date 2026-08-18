package browser

import (
	"context"
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// fakeModule implements widgets.Module plus BrowserAssets.
type fakeModule struct {
	typ string
	css []byte
	js  []byte
}

func (m fakeModule) Type() string { return m.typ }
func (m fakeModule) Renderer() widgets.StaticRenderer {
	return func(json.RawMessage, widgets.RenderContext) (template.HTML, error) { return "", nil }
}
func (m fakeModule) RegisterRoutes(*mux.Router)                        {}
func (m fakeModule) Warmup(context.Context, []json.RawMessage)         {}
func (m fakeModule) CSS() []byte                                       { return m.css }
func (m fakeModule) JS() []byte                                        { return m.js }

// bareModule implements widgets.Module but not widgets.BrowserAssets, to
// prove NewAssets skips modules that ship nothing.
type bareModule struct{ typ string }

func (m bareModule) Type() string { return m.typ }
func (m bareModule) Renderer() widgets.StaticRenderer {
	return func(json.RawMessage, widgets.RenderContext) (template.HTML, error) { return "", nil }
}
func (m bareModule) RegisterRoutes(*mux.Router)                {}
func (m bareModule) Warmup(context.Context, []json.RawMessage) {}

func TestNewAssetsSkipsModulesWithoutBrowserAssets(t *testing.T) {
	a := NewAssets([]widgets.Module{bareModule{typ: "search"}})
	if len(a.WidgetsCSS()) != 0 {
		t.Errorf("WidgetsCSS = %q, want empty", a.WidgetsCSS())
	}
	if len(a.JSTypes()) != 0 {
		t.Errorf("JSTypes = %v, want empty", a.JSTypes())
	}
}

func TestWidgetsCSSConcatenatesInTypeOrder(t *testing.T) {
	a := NewAssets([]widgets.Module{
		fakeModule{typ: "zebra", css: []byte(".dashi-zebra{color:red}")},
		fakeModule{typ: "apple", css: []byte(".dashi-apple{color:blue}")},
	})

	got := string(a.WidgetsCSS())
	if strings.Index(got, "dashi-apple") > strings.Index(got, "dashi-zebra") {
		t.Errorf("expected type-sorted output for deterministic caching, got:\n%s", got)
	}
}

func TestWidgetJSLookup(t *testing.T) {
	a := NewAssets([]widgets.Module{
		fakeModule{typ: "clock", js: []byte("export const x = 1")},
		fakeModule{typ: "bookmark"}, // no JS
	})

	if _, ok := a.WidgetJS("clock"); !ok {
		t.Error("clock JS should be present")
	}
	if _, ok := a.WidgetJS("bookmark"); ok {
		t.Error("bookmark ships no JS, lookup should miss")
	}
	if _, ok := a.WidgetJS("nope"); ok {
		t.Error("unknown type lookup should miss")
	}

	types := a.JSTypes()
	if !types["clock"] || types["bookmark"] {
		t.Errorf("JSTypes = %v, want only clock", types)
	}
}

func TestEmbeddedAssetsArePresent(t *testing.T) {
	a := NewAssets(nil)
	if len(a.ViewerCSS()) == 0 {
		t.Error("viewer.css is empty — run make viewer-css")
	}
	if !strings.Contains(string(a.DashiJS()), "export function poll") {
		t.Error("dashi.js does not export poll")
	}
}
