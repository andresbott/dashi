package backgrounds

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestBrowserCSSNilWhenEmpty(t *testing.T) {
	if got := BrowserCSS(Background{ID: "abc123", Name: "Empty"}); got != nil {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestBrowserCSSColorOnly(t *testing.T) {
	b := Background{ID: "abc123", Name: "C", Color: &Color{Light: "#ffffff", Dark: "#101014"}}
	want := ":root{--dashi-page-bg:#ffffff;}\n" +
		":root[data-color-mode=\"dark\"]{--dashi-page-bg:#101014;}\n"
	if got := string(BrowserCSS(b)); got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestBrowserCSSDarkFallsBackToLight(t *testing.T) {
	b := Background{ID: "abc123", Name: "C", Color: &Color{Light: "#ffffff"}}
	got := string(BrowserCSS(b))
	if strings.Count(got, "#ffffff") != 2 {
		t.Fatalf("expected the dark rule to reuse the light colour, got:\n%s", got)
	}
}

func TestBrowserCSSImageOverGradient(t *testing.T) {
	b := Background{
		ID:   "abc123",
		Name: "Sunset",
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
	got := string(BrowserCSS(b))
	wantLight := ":root{--dashi-page-bg:url('/api/v0/backgrounds/abc123/assets/sunset.jpg') center/cover no-repeat,linear-gradient(to bottom right,#667eea,#764ba2);}"
	wantDark := ":root[data-color-mode=\"dark\"]{--dashi-page-bg:url('/api/v0/data/backgrounds/night.png') center/cover no-repeat,linear-gradient(to bottom right,#1e1b4b,#0f172a);}"
	if !strings.Contains(got, wantLight) {
		t.Errorf("missing light rule\ngot:  %s\nwant: %s", got, wantLight)
	}
	if !strings.Contains(got, wantDark) {
		t.Errorf("missing dark rule\ngot:  %s\nwant: %s", got, wantDark)
	}
}

func TestBrowserCSSColorIsTheFinalLayer(t *testing.T) {
	// Only the last layer of the background shorthand may carry a colour.
	b := Background{
		ID:    "abc123",
		Name:  "I",
		Color: &Color{Light: "#ffffff"},
		Image: &Image{Light: "shared:a.png", Fit: "contain", Position: "center", Repeat: "no-repeat"},
	}
	got := string(BrowserCSS(b))
	want := "url('/api/v0/data/backgrounds/a.png') center/contain no-repeat,#ffffff"
	if !strings.Contains(got, want) {
		t.Fatalf("got:\n%s\nwant substring:\n%s", got, want)
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
			b := Background{
				ID:    "abc123",
				Name:  "I",
				Image: &Image{Light: "shared:a.png", Fit: fit, Position: "center", Repeat: "no-repeat"},
			}
			if got := string(BrowserCSS(b)); !strings.Contains(got, want) {
				t.Fatalf("got %s, want substring %s", got, want)
			}
		})
	}
}

func TestBrowserCSSEscapesQuoteInRef(t *testing.T) {
	// A single quote would terminate url('…'). It must be percent-encoded.
	b := Background{
		ID:    "abc123",
		Name:  "I",
		Image: &Image{Light: "shared:it's.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
	}
	got := string(BrowserCSS(b))
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
	// hand-edited background.json bypasses it.
	b := Background{
		ID:    "abc123",
		Name:  "X",
		Color: &Color{Light: "#fff;}</style><script>alert(1)</script>"},
	}
	got := strings.ToLower(string(BrowserCSS(b)))
	if strings.Contains(got, "</style") || strings.Contains(got, "<script") {
		t.Fatalf("style breakout possible: %s", got)
	}
}

func TestImageCSSSuppressesBaseWhenImagePresent(t *testing.T) {
	b := Background{
		ID:       "abc123",
		Name:     "S",
		Gradient: &Gradient{Direction: "to right", Light: []string{"#000000", "#ffffff"}},
		Image:    &Image{Light: "shared:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
	}
	css, ref := ImageCSS(b, "light")
	if css != "" {
		t.Errorf("expected no CSS when an image is present, got %q", css)
	}
	if ref != "shared:a.png" {
		t.Errorf("got ref %q", ref)
	}
}

func TestImageCSSBaseOnlyEmitsConcreteValue(t *testing.T) {
	b := Background{
		ID:       "abc123",
		Name:     "G",
		Gradient: &Gradient{Direction: "135deg", Light: []string{"#000000", "#ffffff"}, Dark: []string{"#111111", "#222222"}},
	}
	if css, _ := ImageCSS(b, "light"); css != "linear-gradient(135deg,#000000,#ffffff)" {
		t.Errorf("light: got %q", css)
	}
	if css, _ := ImageCSS(b, "dark"); css != "linear-gradient(135deg,#111111,#222222)" {
		t.Errorf("dark: got %q", css)
	}
}

func TestImageCSSAutoResolvesToLight(t *testing.T) {
	b := Background{ID: "abc123", Name: "C", Color: &Color{Light: "#ffffff", Dark: "#000000"}}
	css, _ := ImageCSS(b, "auto")
	if css != "#ffffff" {
		t.Fatalf("expected auto to resolve light, got %q", css)
	}
}

func TestImageCSSDarkImageFallsBackToLight(t *testing.T) {
	b := Background{
		ID:    "abc123",
		Name:  "I",
		Image: &Image{Light: "shared:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
	}
	_, ref := ImageCSS(b, "dark")
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
		{"asset:sunset.jpg", "/api/v0/backgrounds/abc123/assets/sunset.jpg", true},
		{"asset:deep/a b.png", "/api/v0/backgrounds/abc123/assets/deep/a%20b.png", true},
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

// TestBrowserCSSMatchesFixture pins the exact bytes the TypeScript preview
// generator must also produce. webui/src/lib/backgroundCss.test.ts reads the
// same file. Go is the contract: if this test and that one disagree, the
// TypeScript side is wrong.
func TestBrowserCSSMatchesFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/css_fixture.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var cases []struct {
		Name       string     `json:"name"`
		Background Background `json:"background"`
		BrowserCSS string     `json:"browserCss"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("fixture is empty")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := string(BrowserCSS(c.Background)); got != c.BrowserCSS {
				t.Fatalf("got:\n%q\nwant:\n%q", got, c.BrowserCSS)
			}
		})
	}
}
