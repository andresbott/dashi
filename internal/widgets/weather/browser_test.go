package weather

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	weatherpkg "github.com/andresbott/dashi/internal/providers/weather"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// newBrowserTestModules wires both modules to the canned test servers, so
// no test reaches the network.
func newBrowserTestModules(t *testing.T, themeDir string) (*Module, *CompactModule) {
	t.Helper()
	forecastSrv, aqSrv := newTestWeatherServer()
	t.Cleanup(forecastSrv.Close)
	t.Cleanup(aqSrv.Close)
	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	store := widgetthemes.NewStore(themeDir)
	return NewModule(client, store, nil), NewCompactModule(client, store, nil)
}

func renderBrowserHTML(t *testing.T, r widgets.BrowserRenderable, config string, ctx widgets.RenderContext) string {
	t.Helper()
	out, err := r.RenderBrowser(json.RawMessage(config), ctx)
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	return string(out)
}

const fullBrowserConfig = `{"city":"Basel","latitude":47.5,"longitude":7.5,
	"showDetails":true,"showHourly":true,"showForecast":true,"showGraph":true,"graphHours":3,
	"showSunrise":true,"showSunset":true,"showWind":true,"showHumidity":true,
	"showPressure":true,"showUV":true,"showVisibility":true,"showAirQuality":true}`

func TestRenderBrowserFullStructure(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, m, fullBrowserConfig, widgets.RenderContext{})

	for _, want := range []string{
		`class="dashi-weather `,
		"dashi-weather__current", "Basel", "18.5&deg;C", "Partly cloudy",
		"dashi-weather__summary", "Feels like 16.2", "Humidity 65%",
		"dashi-weather__details", "Sunrise", "06:30", "1013 hPa", "AQI", "42",
		"dashi-weather__graph", "data:image/png;base64,", "100%", "0%", "12:00",
		"dashi-weather__hourly",
		"dashi-weather__forecast", "Fri", "Sat",
		"divide-y divide-[var(--dashi-text)]/40",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	// Forecast shows high then low, no boxes.
	if !strings.Contains(html, "20&deg; <span") {
		t.Errorf("forecast should show the high first:\n%s", html)
	}
	for _, bad := range []string{"forecast-box", "hourly-box", "rounded", "bg-[", "ZgotmplZ", "<style"} {
		if strings.Contains(html, bad) {
			t.Errorf("unexpected %q in browser output", bad)
		}
	}
}

func TestRenderBrowserUsesThemeFontIcons(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, m, fullBrowserConfig, widgets.RenderContext{WidgetTheme: "default"})

	if !strings.Contains(html, "font-family: icon-font-default") {
		t.Errorf("expected icon font family in output:\n%s", html)
	}
	if strings.Contains(html, "ti-cloud-question") {
		t.Errorf("every icon should resolve through the default theme:\n%s", html)
	}
	// Icons inherit their slot's font size instead of the image stack's 1.4em.
	if strings.Contains(html, "1.4em") {
		t.Error("browser icons must not hard-code the image stack's 1.4em size")
	}
	// current + 8 details + hourly + 2 forecast days
	if got := strings.Count(html, "font-family: icon-font-default"); got < 12 {
		t.Errorf("expected at least 12 font icons, got %d", got)
	}
}

func TestRenderBrowserImageIconsUseTheIconRoute(t *testing.T) {
	themeDir := t.TempDir()
	iconsDir := filepath.Join(themeDir, "imgtheme", "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "partly-cloudy.svg"), []byte("<svg/>"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := []byte("name: imgtheme\ndescription: test image theme\nicons:\n  type: image\n")
	if err := os.WriteFile(filepath.Join(themeDir, "imgtheme", "widget_theme.yaml"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}

	m, _ := newBrowserTestModules(t, themeDir)
	html := renderBrowserHTML(t, m, `{"city":"Basel","latitude":47.5,"longitude":7.5}`, widgets.RenderContext{WidgetTheme: "imgtheme"})

	if !strings.Contains(html, `src="/api/v0/widget-themes/imgtheme/icons/partly-cloudy"`) {
		t.Errorf("expected icon route in img src:\n%s", html)
	}
	if strings.Contains(html, iconsDir) {
		t.Error("a filesystem path does not resolve in a browser")
	}
}

func TestRenderBrowserHonoursToggles(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, m, `{"city":"Basel","latitude":47.5,"longitude":7.5,"showCurrent":false,"showForecast":true,"forecastDays":1}`, widgets.RenderContext{})

	for _, absent := range []string{
		"Basel", "18.5", "dashi-weather__current", "dashi-weather__details",
		"dashi-weather__graph", "dashi-weather__hourly", "Sat",
	} {
		if strings.Contains(html, absent) {
			t.Errorf("%q must be absent:\n%s", absent, html)
		}
	}
	if got := strings.Count(html, "dashi-weather__day "); got != 1 {
		t.Errorf("forecastDays=1 should render 1 day, got %d", got)
	}
}

func TestRenderBrowserGraphAxisToggles(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, m, `{"latitude":47.5,"longitude":7.5,"showCurrent":false,"showGraph":true,"graphHours":3,"graphShowRain":false}`, widgets.RenderContext{})
	if strings.Contains(html, "100%") {
		t.Error("rain axis must be hidden when graphShowRain is false")
	}
	if !strings.Contains(html, "20°") || !strings.Contains(html, "18°") {
		t.Errorf("temperature axis labels missing:\n%s", html)
	}
}

func TestRenderBrowserCompact(t *testing.T) {
	_, cm := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, cm, `{"city":"Basel","latitude":47.5,"longitude":7.5,"showForecast":true,"showGraph":true,"compactFeelsLike":true,"compactAlign":"right"}`, widgets.RenderContext{})

	for _, want := range []string{"dashi-weather__compact", "justify-end", "Basel", "18.5&deg;C", "Partly cloudy", "Feels like 16.2", "font-family: icon-font-"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	for _, bad := range []string{"dashi-weather__forecast", "dashi-weather__graph"} {
		if strings.Contains(html, bad) {
			t.Errorf("compact must not render %q", bad)
		}
	}

	hidden := renderBrowserHTML(t, cm, `{"city":"Basel","latitude":47.5,"longitude":7.5,"compactCity":false,"compactDescription":false}`, widgets.RenderContext{})
	for _, bad := range []string{"Basel", "Partly cloudy", "Feels like"} {
		if strings.Contains(hidden, bad) {
			t.Errorf("%q must be hidden by its compact toggle", bad)
		}
	}
	if !strings.Contains(hidden, "justify-start") {
		t.Error("compact alignment defaults to left")
	}
}

func TestRenderBrowserFetchErrorIsReturned(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	if _, err := m.RenderBrowser(json.RawMessage(`{not json`), widgets.RenderContext{}); err == nil {
		t.Error("malformed config must fail like the image renderer")
	}
}

// Tailwind extracts classes from browser.html statically: a class built
// from template data would silently produce no CSS.
func TestBrowserTemplateHasNoComputedClasses(t *testing.T) {
	// Control actions choosing between literal classes are fine; only value
	// output ({{.X}}, {{printf ...}}) inside a class attribute is not.
	control := regexp.MustCompile(`\{\{(if|else if) [^}]*\}\}|\{\{else\}\}|\{\{end\}\}`)
	cleaned := control.ReplaceAllString(browserHTML, "")
	for _, attr := range regexp.MustCompile(`class="[^"]*"`).FindAllString(cleaned, -1) {
		if strings.Contains(attr, "{{") {
			t.Errorf("class attribute interpolates data: %s", attr)
		}
	}
}

func TestBrowserModulesShipNoAssets(t *testing.T) {
	// Styling is pure Tailwind; there is no weather.css to scope.
	var m any = &Module{}
	if _, ok := m.(widgets.BrowserAssets); ok {
		t.Error("weather ships no CSS/JS; add a scoping test if that changes")
	}
}

// The axis labels overlay the graph's corners (temperature left, rain
// right) instead of taking side columns, so the curve spans the full width.
func TestRenderBrowserGraphLabelsOverlayThePlot(t *testing.T) {
	m, _ := newBrowserTestModules(t, "")
	html := renderBrowserHTML(t, m, fullBrowserConfig, widgets.RenderContext{})

	plotStart := strings.Index(html, "dashi-weather__plot relative")
	if plotStart < 0 {
		t.Fatalf("missing the relative plot wrapper:\n%s", html)
	}
	plot := html[plotStart:]
	plot = plot[:strings.Index(plot, "</div>")]
	for _, want := range []string{
		`<img class="block h-auto w-full"`,
		`<span class="absolute top-0 left-0">`,
		`<span class="absolute bottom-0 left-0">`,
		`<span class="absolute top-0 right-0">100%</span>`,
		`<span class="absolute right-0 bottom-0">0%</span>`,
	} {
		if !strings.Contains(plot, want) {
			t.Errorf("missing %q inside the plot:\n%s", want, plot)
		}
	}
	if strings.Contains(html, "grid-cols-[auto_minmax(0,1fr)_auto]") {
		t.Error("the side-column grid must be gone")
	}
}
