package themes

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultPaletteValues(t *testing.T) {
	light := Palette{
		BG: "#ffffff", Surface: "#f5f5f7", Border: "#e0e0e5",
		Text: "#333333", TextMuted: "#666666",
		Primary: "#0a84ff", OnPrimary: "#ffffff",
		Secondary: "#8b5cf6", Success: "#16a34a", Warning: "#d97706", Danger: "#dc2626",
	}
	dark := Palette{
		BG: "#1a1a2e", Surface: "#23233d", Border: "#33334d",
		Text: "#e0e0e0", TextMuted: "#999999",
		Primary: "#4a90d9", OnPrimary: "#ffffff",
		Secondary: "#a78bfa", Success: "#22c55e", Warning: "#f59e0b", Danger: "#ef4444",
	}
	if got := DefaultPalette("light"); got != light {
		t.Errorf("light = %+v", got)
	}
	if got := DefaultPalette("dark"); got != dark {
		t.Errorf("dark = %+v", got)
	}
	if got := DefaultPalette(""); got != light {
		t.Errorf("empty = %+v, want light", got)
	}
}

func TestPaletteValidate(t *testing.T) {
	with := func(f func(p *Palette)) Palette {
		p := DefaultPalette("light")
		f(&p)
		return p
	}
	tests := []struct {
		name    string
		p       Palette
		wantErr string // empty means no error
	}{
		{"default", DefaultPalette("light"), ""},
		{"translucent surface", with(func(p *Palette) { p.Surface = "#ffffff80" }), ""},
		{"translucent bg", with(func(p *Palette) { p.BG = "#ffffff80" }), "bg"},
		{"empty text", with(func(p *Palette) { p.Text = "" }), "text"},
		{"named danger", with(func(p *Palette) { p.Danger = "red" }), "danger"},
		{"short primary", with(func(p *Palette) { p.Primary = "#12345" }), "primary"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("error %v does not wrap ErrInvalidConfig", err)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestPaletteWithDefaults(t *testing.T) {
	custom := Palette{
		BG: "#101010", Surface: "#202020", Border: "#303030",
		Text: "red;}", TextMuted: "#404040",
		Primary: "#505050", OnPrimary: "#606060",
		Secondary: "", Success: "#707070", Warning: "#808080", Danger: "#909090",
	}
	def := DefaultPalette("dark")
	got := custom.withDefaults("dark")
	if got.Text != def.Text {
		t.Errorf("Text = %q, want %q", got.Text, def.Text)
	}
	if got.Secondary != def.Secondary {
		t.Errorf("Secondary = %q, want %q", got.Secondary, def.Secondary)
	}
	want := custom
	want.Text, want.Secondary = def.Text, def.Secondary
	if got != want {
		t.Errorf("other roles changed: %+v", got)
	}
}

func TestPaletteFlatten(t *testing.T) {
	p := DefaultPalette("light")
	p.BG, p.Surface = "#000000", "#ffffff80"
	if got := p.Flatten().Surface; got != "#808080" {
		t.Errorf("Surface = %q, want #808080", got)
	}

	p = DefaultPalette("light")
	p.BG, p.Border = "#ffffff", "#00000000"
	if got := p.Flatten().Border; got != "#ffffff" {
		t.Errorf("Border = %q, want #ffffff", got)
	}

	p = DefaultPalette("light")
	p.Text, p.Primary = "#abc", "#a1b2c3"
	got := p.Flatten()
	if got.Text != "#abc" || got.Primary != "#a1b2c3" {
		t.Errorf("opaque values changed: %q %q", got.Text, got.Primary)
	}
}

func TestWritePaletteVarsOrder(t *testing.T) {
	var b strings.Builder
	writePaletteVars(&b, DefaultPalette("light"))
	want := "--dashi-bg:#ffffff;--dashi-surface:#f5f5f7;--dashi-border:#e0e0e5;" +
		"--dashi-text:#333333;--dashi-text-muted:#666666;" +
		"--dashi-primary:#0a84ff;--dashi-on-primary:#ffffff;" +
		"--dashi-secondary:#8b5cf6;--dashi-success:#16a34a;" +
		"--dashi-warning:#d97706;--dashi-danger:#dc2626;"
	if got := b.String(); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
