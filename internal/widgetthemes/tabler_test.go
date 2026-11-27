package widgetthemes

import "testing"

func TestTablerCodepoint(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		wantOK bool
	}{
		// Code points from @tabler/icons-webfont, checked against the font.
		{name: "ti-brand-github", want: "ec1c", wantOK: true},
		{name: "brand-github", want: "ec1c", wantOK: true},
		{name: "ti-device-floppy", want: "eb62", wantOK: true},
		// Past U+F8FF, outside the private-use block.
		{name: "ti-brand-wikipedia", want: "fa79", wantOK: true},
		{name: "ti-no-such-icon-anywhere", wantOK: false},
		{name: "", wantOK: false},
		{name: "ti-", wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tablerCodepoint(tc.name)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("tablerCodepoint(%q) = %q, %v; want %q, %v", tc.name, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// Any Tabler icon the editor offers resolves through the default theme,
// not only the ~36 names its manifest maps (weather and details).
func TestResolveIconFallsBackToTablerGlyphNames(t *testing.T) {
	s := NewStore("")

	got, err := s.ResolveIcon(DefaultName, "ti-brand-github")
	if err != nil {
		t.Fatalf("ResolveIcon: %v", err)
	}
	if got.Type != ThemeTypeFont || got.Codepoint != "ec1c" || got.FontFamily != IconFontFamily(DefaultName) {
		t.Errorf("got %+v, want font icon ec1c in %s", got, IconFontFamily(DefaultName))
	}

	// Manifest names still win.
	if got, err := s.ResolveIcon(DefaultName, "clear-sky"); err != nil || got.Codepoint != "eb30" {
		t.Errorf("manifest icon: got %+v, %v", got, err)
	}

	if _, err := s.ResolveIcon(DefaultName, "ti-no-such-icon-anywhere"); err == nil {
		t.Error("expected an error for an unknown icon")
	}
}
