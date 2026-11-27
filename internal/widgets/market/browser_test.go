package market

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	mkt "github.com/andresbott/dashi/internal/providers/market"
	"github.com/andresbott/dashi/internal/widgets"
)

// Compile-time check: the router registers the browser renderer only when
// the module implements widgets.BrowserRenderable.
var _ widgets.BrowserRenderable = (*Module)(nil)

func browserModule(t *testing.T, meta, closes string) *Module {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{"chart":{"result":[{"meta":` + meta + `,` +
			`"timestamp":[1704067200,1704153600,1704240000],"indicators":{"quote":[{"close":` + closes + `}]}}]}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(srv.Close)
	client := mkt.NewClient(srv.Client())
	client.BaseURL = srv.URL
	return NewModule(client, nil)
}

const adbeMeta = `{"symbol":"ADBE","shortName":"Adobe Inc.","currency":"USD",` +
	`"regularMarketPrice":232.83,"chartPreviousClose":263.82}`

func renderBrowser(t *testing.T, m *Module, config string) string {
	t.Helper()
	out, err := m.RenderBrowser(json.RawMessage(config), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	return string(out)
}

func TestRenderBrowserShowsQuote(t *testing.T) {
	html := renderBrowser(t, browserModule(t, adbeMeta, `[263.82,250.0,232.83]`), `{"symbol":"ADBE"}`)

	for _, want := range []string{
		`class="dashi-market `,
		`text-lg font-semibold">ADBE</span>`,
		`text-xs text-[var(--dashi-text-muted)]">Adobe Inc.</span>`,
		`text-2xl font-semibold tabular-nums">232.83</span>`,
		`-30.99 (-11.75%)</span>`,
		`<span class="dashi-market__currency">USD</span>`,
		`<span class="dashi-market__range">1 Month</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	// A fall is coloured danger, never success; no card surface behind it.
	if !regexp.MustCompile(`dashi-market__change[^"]*text-\[var\(--dashi-danger\)\]`).MatchString(html) {
		t.Errorf("negative change must use the danger colour:\n%s", html)
	}
	if strings.Contains(html, "--dashi-success") {
		t.Errorf("a falling quote and chart must not use the success colour:\n%s", html)
	}
	if strings.Contains(html, "--dashi-surface") {
		t.Errorf("the widget must sit on the page background, not a surface:\n%s", html)
	}
}

func TestRenderBrowserRisingQuoteUsesSuccess(t *testing.T) {
	meta := `{"symbol":"VOO","shortName":"Vanguard S&P 500 ETF","currency":"USD",` +
		`"regularMarketPrice":610.5,"chartPreviousClose":600}`
	html := renderBrowser(t, browserModule(t, meta, `[600,605,610.5]`), `{"symbol":"VOO","range":"1y"}`)

	if !strings.Contains(html, `&#43;10.50 (&#43;1.75%)`) {
		t.Errorf("missing signed change in:\n%s", html)
	}
	if !regexp.MustCompile(`dashi-market__change[^"]*text-\[var\(--dashi-success\)\]`).MatchString(html) {
		t.Errorf("positive change must use the success colour:\n%s", html)
	}
	if !regexp.MustCompile(`dashi-market__chart[^"]*text-\[var\(--dashi-success\)\]`).MatchString(html) {
		t.Errorf("rising chart must use the success colour:\n%s", html)
	}
	if !strings.Contains(html, "1 Year") {
		t.Errorf("missing range label in:\n%s", html)
	}
	if strings.Contains(html, "--dashi-danger") {
		t.Errorf("a rising quote must not use the danger colour:\n%s", html)
	}
}

func TestRenderBrowserDrawsAShortSVGAreaChart(t *testing.T) {
	html := renderBrowser(t, browserModule(t, adbeMeta, `[263.82,250.0,232.83]`), `{"symbol":"ADBE"}`)

	if strings.Contains(html, "base64") || strings.Contains(html, "<img") {
		t.Errorf("the browser chart is inline SVG, not a PNG:\n%s", html)
	}
	if !strings.Contains(html, "aspect-[5/1] w-full") {
		t.Errorf("chart must be full width at 3:1:\n%s", html)
	}
	if got := strings.Count(html, `viewBox="0 0 300 100"`); got != 2 {
		t.Errorf("want fill and line SVGs, got %d viewBoxes:\n%s", got, html)
	}
	// Line: smoothed curve from the left to the right edge. Area: the same
	// curve closed down to the bottom edge.
	if !regexp.MustCompile(`<path d="M0,[0-9.]+ C[^"]* 300,[0-9.]+" fill="none" stroke="currentColor"`).MatchString(html) {
		t.Errorf("missing line path:\n%s", html)
	}
	if !regexp.MustCompile(`<path d="M0,[^"]* L300,100 L0,100 Z" fill="currentColor"`).MatchString(html) {
		t.Errorf("missing area path:\n%s", html)
	}
	if !strings.Contains(html, "[mask-image:linear-gradient(to_bottom,black,transparent)]") {
		t.Errorf("area fill must fade downwards:\n%s", html)
	}
}

func TestRenderBrowserHonoursChartAndChangeToggles(t *testing.T) {
	m := browserModule(t, adbeMeta, `[263.82,250.0,232.83]`)
	html := renderBrowser(t, m, `{"symbol":"ADBE","showChart":false,"showChange":false}`)
	if strings.Contains(html, "<svg") || strings.Contains(html, "dashi-market__chart") {
		t.Errorf("showChart=false must render no chart:\n%s", html)
	}
	if strings.Contains(html, "dashi-market__change") {
		t.Errorf("showChange=false must render no change:\n%s", html)
	}
	if !strings.Contains(html, "232.83") {
		t.Errorf("price must still render:\n%s", html)
	}
}

func TestRenderBrowserSkipsChartWithOnePoint(t *testing.T) {
	html := renderBrowser(t, browserModule(t, adbeMeta, `[232.83]`), `{"symbol":"ADBE"}`)
	if strings.Contains(html, "<svg") {
		t.Errorf("one point is no chart:\n%s", html)
	}
}

func TestRenderBrowserUnconfigured(t *testing.T) {
	m := NewModule(mkt.NewClient(nil), nil)
	for _, cfg := range []string{``, `{}`, `{"range":"1mo"}`} {
		out, err := m.RenderBrowser(json.RawMessage(cfg), widgets.RenderContext{})
		if err != nil {
			t.Fatalf("config %q: %v", cfg, err)
		}
		html := string(out)
		if !strings.Contains(html, `text-center text-sm text-[var(--dashi-text-muted)]">No symbol configured</div>`) {
			t.Errorf("config %q: missing muted centred message:\n%s", cfg, html)
		}
		if strings.Contains(html, "dashi-market__price") {
			t.Errorf("config %q: unconfigured widget must not render a quote:\n%s", cfg, html)
		}
	}
}

func TestRenderBrowserErrors(t *testing.T) {
	m := NewModule(mkt.NewClient(nil), nil)
	if _, err := m.RenderBrowser(json.RawMessage(`{invalid`), widgets.RenderContext{}); err == nil ||
		!strings.Contains(err.Error(), "market config") {
		t.Errorf("want a market config error, got %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := mkt.NewClient(srv.Client())
	client.BaseURL = srv.URL
	if _, err := NewModule(client, nil).RenderBrowser(json.RawMessage(`{"symbol":"ADBE"}`), widgets.RenderContext{}); err == nil ||
		!strings.Contains(err.Error(), "market fetch") {
		t.Errorf("want a market fetch error, got %v", err)
	}
}

func TestRenderBrowserEscapesQuoteText(t *testing.T) {
	meta := `{"symbol":"X","shortName":"<script>alert(1)</script>","currency":"USD",` +
		`"regularMarketPrice":1,"chartPreviousClose":1}`
	html := renderBrowser(t, browserModule(t, meta, `[1,1,1]`), `{"symbol":"X"}`)
	if strings.Contains(html, "<script>") {
		t.Errorf("provider text must be escaped:\n%s", html)
	}
}

func TestBuildSVGChartFlatSeriesStaysInsideTheBox(t *testing.T) {
	c := buildSVGChart([]mkt.PricePoint{{Close: 5}, {Close: 5}})
	if c.Line != "M0,88.46 L300,88.46" {
		t.Errorf("flat two-point series: got %q", c.Line)
	}
	if !c.Up {
		t.Error("an unchanged series counts as up, like the PNG chart")
	}
}
