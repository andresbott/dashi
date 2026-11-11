package themes

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// lightRule returns the :root rule BrowserCSS emits for t, so tests can
// assert on the light block without matching the dark one by accident.
func lightRule(t *testing.T, css string) string {
	t.Helper()
	rule, _, ok := strings.Cut(css, "\n")
	if !ok || !strings.HasPrefix(rule, ":root{") {
		t.Fatalf("no light rule in %q", css)
	}
	return rule
}

func darkRule(t *testing.T, css string) string {
	t.Helper()
	_, rest, _ := strings.Cut(css, "\n")
	if !strings.HasPrefix(rest, `:root[data-color-mode="dark"]{`) {
		t.Fatalf("no dark rule in %q", css)
	}
	return rest
}

func TestBrowserCSSDefaultTheme(t *testing.T) {
	got := string(BrowserCSS(DefaultTheme()))
	if !strings.Contains(got, ":root{--dashi-bg:#ffffff;") {
		t.Errorf("missing light palette vars:\n%s", got)
	}
	if !strings.Contains(got, `:root[data-color-mode="dark"]{--dashi-bg:#1a1a2e;`) {
		t.Errorf("missing dark palette vars:\n%s", got)
	}
	if strings.Contains(got, "--dashi-page-bg") {
		t.Errorf("a plain background must not set --dashi-page-bg:\n%s", got)
	}
}

func TestBrowserCSSIsNeverNil(t *testing.T) {
	if BrowserCSS(Theme{}) == nil {
		t.Fatal("BrowserCSS must always return the palette vars")
	}
}

func TestBrowserCSSGradientAndImage(t *testing.T) {
	th := testTheme("GI")
	th.ID = "abc123"
	th.Background = Background{
		Gradient: &Gradient{Direction: "to bottom", Light: []string{"#000000", "#ffffff"}},
		Image:    &Image{Light: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
	}
	light := lightRule(t, string(BrowserCSS(th)))
	want := "--dashi-page-bg:url('/api/v0/themes/abc123/assets/a.png') center/cover no-repeat,linear-gradient(to bottom,#000000,#ffffff);"
	if !strings.Contains(light, want) {
		t.Fatalf("light rule:\n%s\nwant substring:\n%s", light, want)
	}
}

func TestBrowserCSSFallsBackPerRoleOnBadValue(t *testing.T) {
	// Review Focus 1: a hand-edited theme.json with a malformed colour still
	// renders, using the default for that role, and cannot inject CSS.
	th := testTheme("Bad")
	th.Colors.Light.Text = "red;}body{x"
	got := string(BrowserCSS(th))
	if !strings.Contains(lightRule(t, got), "--dashi-text:#333333;") {
		t.Errorf("expected the default text colour:\n%s", got)
	}
	if strings.Contains(got, "body{") || strings.Contains(got, "red;") {
		t.Errorf("malformed value leaked into the CSS:\n%s", got)
	}
}

func TestBrowserCSSUsesEachModesPalette(t *testing.T) {
	th := testTheme("Modes")
	th.Colors.Light.Primary = "#111111"
	th.Colors.Dark.Primary = "#222222"
	got := string(BrowserCSS(th))
	if !strings.Contains(lightRule(t, got), "--dashi-primary:#111111;") {
		t.Errorf("light rule lacks the light primary:\n%s", got)
	}
	if !strings.Contains(darkRule(t, got), "--dashi-primary:#222222;") {
		t.Errorf("dark rule lacks the dark primary:\n%s", got)
	}
}

func TestBrowserCSSImageOverGradient(t *testing.T) {
	th := testTheme("Sunset")
	th.ID = "abc123"
	th.Background = Background{
		Gradient: &Gradient{
			Direction: "to bottom right",
			Light:     []string{"#667eea", "#764ba2"},
			Dark:      []string{"#1e1b4b", "#0f172a"},
		},
		Image: &Image{
			Light: "asset:sunset.jpg", Dark: "shared:night.png",
			Fit: "cover", Position: "center", Repeat: "no-repeat",
		},
	}
	got := string(BrowserCSS(th))
	wantLight := "--dashi-page-bg:url('/api/v0/themes/abc123/assets/sunset.jpg') center/cover no-repeat,linear-gradient(to bottom right,#667eea,#764ba2);}"
	wantDark := "--dashi-page-bg:url('/api/v0/data/backgrounds/night.png') center/cover no-repeat,linear-gradient(to bottom right,#1e1b4b,#0f172a);}"
	if !strings.Contains(lightRule(t, got), wantLight) {
		t.Errorf("missing light page-bg\ngot:  %s\nwant: %s", got, wantLight)
	}
	if !strings.Contains(darkRule(t, got), wantDark) {
		t.Errorf("missing dark page-bg\ngot:  %s\nwant: %s", got, wantDark)
	}
}

func TestBrowserCSSDarkGradientFallsBackToLight(t *testing.T) {
	th := testTheme("G")
	th.Background.Gradient = &Gradient{Direction: "135deg", Light: []string{"#000000", "#ffffff"}}
	got := string(BrowserCSS(th))
	if strings.Count(got, "--dashi-page-bg:linear-gradient(135deg,#000000,#ffffff);") != 2 {
		t.Fatalf("expected the dark rule to reuse the light stops, got:\n%s", got)
	}
}

func TestBrowserCSSFitMapping(t *testing.T) {
	tests := map[string]string{
		"cover":    "center/cover",
		"contain":  "center/contain",
		"stretch":  "center/100% 100%",
		"original": "center/auto",
	}
	for fit, want := range tests {
		t.Run(fit, func(t *testing.T) {
			th := testTheme("I")
			th.ID = "abc123"
			th.Background.Image = &Image{Light: "shared:a.png", Fit: fit, Position: "center", Repeat: "no-repeat"}
			if got := string(BrowserCSS(th)); !strings.Contains(got, want) {
				t.Fatalf("got %s, want substring %s", got, want)
			}
		})
	}
}

func TestBrowserCSSEscapesQuoteInRef(t *testing.T) {
	// A single quote would terminate url('…'). It must be percent-encoded.
	th := testTheme("I")
	th.ID = "abc123"
	th.Background.Image = &Image{Light: "shared:it's.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}
	got := string(BrowserCSS(th))
	if strings.Contains(got, "it's") {
		t.Fatalf("raw single quote leaked into url(): %s", got)
	}
	if !strings.Contains(got, "it%27s.png") {
		t.Fatalf("expected percent-encoded quote, got %s", got)
	}
}

func TestBrowserCSSCannotBreakOutOfStyleElement(t *testing.T) {
	// html/template applies no escaping inside <style>, so nothing generated
	// here may contain a sequence that closes the element. Every input below
	// is rejected by Validate, but BrowserCSS must be safe even if a
	// hand-edited theme.json bypasses it.
	th := testTheme("X")
	th.Colors.Light.BG = "#fff;}</style><script>alert(1)</script>"
	th.Colors.Dark.Surface = "</style><script>"
	th.Background.Gradient = &Gradient{
		Direction: "to right",
		Light:     []string{"#000", "#fff);}</style><script>alert(1)</script>"},
	}
	got := strings.ToLower(string(BrowserCSS(th)))
	if strings.Contains(got, "</style") || strings.Contains(got, "<script") {
		t.Fatalf("style breakout possible: %s", got)
	}
}

func TestBrowserCSSDropsImageWithBadEnums(t *testing.T) {
	// A hand-edited theme.json can carry any string in fit/position/repeat;
	// those are written into the CSS verbatim, so the layer is dropped.
	th := testTheme("X")
	th.Background.Image = &Image{Light: "shared:a.png", Fit: "cover", Position: "center;}body{x", Repeat: "no-repeat"}
	got := string(BrowserCSS(th))
	if strings.Contains(got, "body{") || strings.Contains(got, "--dashi-page-bg") {
		t.Fatalf("unvalidated image enums reached the CSS: %s", got)
	}
}

func TestImageCSSRejectsUnsafeGradient(t *testing.T) {
	th := testTheme("X")
	th.Background.Gradient = &Gradient{Direction: "to right);}body{x", Light: []string{"#000", "#fff"}}
	if css, ref := ImageCSS(th, "light"); css != "" || ref != "" {
		t.Fatalf("unsafe gradient reached the litehtml CSS: %q %q", css, ref)
	}
}

func TestImageCSSPrefersImage(t *testing.T) {
	th := testTheme("S")
	th.Background = Background{
		Gradient: &Gradient{Direction: "to right", Light: []string{"#000000", "#ffffff"}},
		Image:    &Image{Light: "shared:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
	}
	css, ref := ImageCSS(th, "light")
	if css != "" {
		t.Errorf("expected no CSS when an image is present, got %q", css)
	}
	if ref != "shared:a.png" {
		t.Errorf("got ref %q", ref)
	}
}

func TestImageCSSBaseOnlyEmitsConcreteValue(t *testing.T) {
	th := testTheme("G")
	th.Background.Gradient = &Gradient{Direction: "135deg", Light: []string{"#000000", "#ffffff"}, Dark: []string{"#111111", "#222222"}}
	if css, _ := ImageCSS(th, "light"); css != "linear-gradient(135deg,#000000,#ffffff)" {
		t.Errorf("light: got %q", css)
	}
	if css, _ := ImageCSS(th, "dark"); css != "linear-gradient(135deg,#111111,#222222)" {
		t.Errorf("dark: got %q", css)
	}
}

func TestImageCSSPlainBackgroundIsEmpty(t *testing.T) {
	css, ref := ImageCSS(DefaultTheme(), "dark")
	if css != "" || ref != "" {
		t.Fatalf("a plain background has no image CSS, got %q %q", css, ref)
	}
}

func TestImageCSSAutoResolvesToLight(t *testing.T) {
	th := testTheme("G")
	th.Background.Gradient = &Gradient{Direction: "to right", Light: []string{"#ffffff", "#eeeeee"}, Dark: []string{"#000000", "#111111"}}
	css, _ := ImageCSS(th, "auto")
	if css != "linear-gradient(to right,#ffffff,#eeeeee)" {
		t.Fatalf("expected auto to resolve light, got %q", css)
	}
}

func TestImageCSSDarkImageFallsBackToLight(t *testing.T) {
	th := testTheme("I")
	th.Background.Image = &Image{Light: "shared:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}
	_, ref := ImageCSS(th, "dark")
	if ref != "shared:a.png" {
		t.Fatalf("got %q", ref)
	}
}

func TestRefURL(t *testing.T) {
	tests := []struct {
		ref  string
		want string
		ok   bool
	}{
		{"asset:sunset.jpg", "/api/v0/themes/abc123/assets/sunset.jpg", true},
		{"asset:deep/a b.png", "/api/v0/themes/abc123/assets/deep/a%20b.png", true},
		{"shared:night.png", "/api/v0/data/backgrounds/night.png", true},
		{"theme:default/x.png", "", false},
		{"nonsense", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got, ok := RefURL("abc123", tt.ref)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBrowserCSSMatchesFixture pins the exact bytes BrowserCSS produces.
// The TypeScript preview generator is meant to read the same file once it
// is ported to themes; Go is the contract. The older css_fixture.json still
// backs webui/src/lib/backgroundCss.test.ts until that port.
func TestBrowserCSSMatchesFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/theme_css_fixture.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var cases []struct {
		Name       string `json:"name"`
		Theme      Theme  `json:"theme"`
		BrowserCSS string `json:"browserCss"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("fixture is empty")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Theme.Validate(); err != nil {
				t.Fatalf("fixture themes must be valid: %v", err)
			}
			if got := string(BrowserCSS(c.Theme)); got != c.BrowserCSS {
				t.Fatalf("got:\n%q\nwant:\n%q", got, c.BrowserCSS)
			}
		})
	}
}
