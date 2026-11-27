package themes

import (
	"errors"
	"strings"
	"testing"
)

// testTheme returns a valid theme with the default palettes and an empty
// background, ready for a test to tweak.
func testTheme(name string) Theme {
	return Theme{Name: name, DarkMode: true, Colors: DefaultTheme().Colors}
}

func TestBuiltinsHaveDarkMode(t *testing.T) {
	for id, fn := range builtins {
		if !fn().DarkMode {
			t.Errorf("builtin %q must have dark mode on", id)
		}
	}
}

func TestPaletteDarkModeOffRendersLight(t *testing.T) {
	th := testTheme("Off")
	th.DarkMode = false
	th.Colors.Light.Primary = "#111111"
	th.Colors.Dark.Primary = "#222222"
	if got := th.Palette("dark"); got != th.Palette("light") {
		t.Fatalf("dark mode off: Palette(dark) = %+v, want the light palette", got)
	}
	if got := th.Palette("dark").Primary; got != "#111111" {
		t.Fatalf("Primary = %q, want the light primary", got)
	}
}

func TestValidateDarkModeOff(t *testing.T) {
	tests := map[string]struct {
		mutate func(th *Theme)
		valid  bool
	}{
		"empty dark palette":       {func(th *Theme) { th.Colors.Dark = Palette{} }, true},
		"partial dark palette":     {func(th *Theme) { th.Colors.Dark = Palette{Primary: "#123456"} }, true},
		"bad dark role":            {func(th *Theme) { th.Colors.Dark = Palette{Primary: "red"} }, false},
		"translucent dark bg":      {func(th *Theme) { th.Colors.Dark.BG = "#00000080" }, false},
		"light palette still full": {func(th *Theme) { th.Colors.Light.Text = "" }, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			th := testTheme("Off")
			th.DarkMode = false
			tt.mutate(&th)
			err := th.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if !tt.valid && !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("expected ErrInvalidConfig, got %v", err)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	d := DefaultTheme()
	if DefaultID != "default" || d.ID != DefaultID {
		t.Errorf("ID = %q, want %q", d.ID, "default")
	}
	if d.Name != "Default" {
		t.Errorf("Name = %q, want Default", d.Name)
	}
	if d.Background.Gradient != nil || d.Background.Image != nil {
		t.Errorf("expected an empty background, got %+v", d.Background)
	}
	if d.Colors.Light != DefaultPalette("light") || d.Colors.Dark != DefaultPalette("dark") {
		t.Errorf("expected the default palettes, got %+v", d.Colors)
	}
	if err := d.Validate(); err != nil {
		t.Errorf("the default theme must validate: %v", err)
	}
}

func TestThemePaletteFallsBackPerRole(t *testing.T) {
	th := testTheme("P")
	th.Colors.Dark.Text = "nope"
	th.Colors.Dark.Primary = "#123456"
	got := th.Palette("dark")
	if got.Text != DefaultPalette("dark").Text {
		t.Errorf("Text = %q, want the dark default", got.Text)
	}
	if got.Primary != "#123456" {
		t.Errorf("Primary = %q, want the stored value", got.Primary)
	}
	if got := th.Palette("auto"); got != DefaultPalette("light") {
		t.Errorf("auto must resolve light, got %+v", got)
	}
}

func TestValidateAcceptsEmptyBackground(t *testing.T) {
	// A freshly created theme has no gradient or image. That is the legal
	// empty state and renders the palette bg.
	if err := testTheme("Empty").Validate(); err != nil {
		t.Fatalf("expected empty background to be valid, got %v", err)
	}
}

func TestValidateRequiresBothPalettes(t *testing.T) {
	tests := map[string]func(th *Theme){
		"zero light":       func(th *Theme) { th.Colors.Light = Palette{} },
		"zero dark":        func(th *Theme) { th.Colors.Dark = Palette{} },
		"bad light role":   func(th *Theme) { th.Colors.Light.Danger = "red" },
		"translucent dark": func(th *Theme) { th.Colors.Dark.BG = "#00000080" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			th := testTheme("C")
			mutate(&th)
			if err := th.Validate(); !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("expected ErrInvalidConfig, got %v", err)
			}
		})
	}
}

func TestValidateNamesTheFailingMode(t *testing.T) {
	th := testTheme("C")
	th.Colors.Dark.Text = ""
	err := th.Validate()
	if err == nil || !strings.Contains(err.Error(), "dark") || !strings.Contains(err.Error(), "text") {
		t.Fatalf("expected an error naming the dark palette and the role, got %v", err)
	}
}

func TestValidateGradient(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		stops     []string
		valid     bool
	}{
		{"keyword direction", "to bottom right", []string{"#000", "#fff"}, true},
		{"angle direction", "135deg", []string{"#000", "#fff"}, true},
		{"three stops", "to right", []string{"#000", "#888", "#fff"}, true},
		{"alpha stop", "to right", []string{"#00000080", "#fff"}, true},
		{"angle out of range", "400deg", []string{"#000", "#fff"}, false},
		{"bogus direction", "sideways", []string{"#000", "#fff"}, false},
		{"injected direction", "to right,red);}", []string{"#000", "#fff"}, false},
		{"single stop", "to right", []string{"#000"}, false},
		{"bad stop", "to right", []string{"#000", "url(x)"}, false},
		{"injected stop", "to right", []string{"#000", "#fff;}</style><script>"}, false},
		{"named stop", "to right", []string{"#000", "red"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := testTheme("G")
			th.Background.Gradient = &Gradient{Direction: tt.direction, Light: tt.stops}
			err := th.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("expected rejection")
			}
		})
	}
}

func TestValidateGradientDarkStopsWhenPresent(t *testing.T) {
	th := testTheme("G")
	th.Background.Gradient = &Gradient{
		Direction: "to right",
		Light:     []string{"#000", "#fff"},
		Dark:      []string{"#000", "nope"},
	}
	if err := th.Validate(); err == nil {
		t.Fatal("expected dark stops to be validated too")
	}
}

func TestValidateImageEnums(t *testing.T) {
	tests := []struct {
		name          string
		fit, pos, rep string
		valid         bool
	}{
		{"defaults", "cover", "center", "no-repeat", true},
		{"contain top left", "contain", "top left", "repeat", true},
		{"stretch", "stretch", "bottom", "repeat-x", true},
		{"original", "original", "right", "repeat-y", true},
		{"bad fit", "zoom", "center", "no-repeat", false},
		{"bad position", "cover", "middle", "no-repeat", false},
		{"bad repeat", "cover", "center", "tile", false},
		{"injected fit", "cover;}", "center", "no-repeat", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := testTheme("I")
			th.Background.Image = &Image{
				Light: "asset:a.png", Fit: tt.fit, Position: tt.pos, Repeat: tt.rep,
			}
			err := th.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("expected rejection")
			}
		})
	}
}

func TestValidateImageRefs(t *testing.T) {
	tests := []struct {
		ref   string
		valid bool
	}{
		{"asset:sunset.jpg", true},
		{"shared:night.png", false}, // the shared pool was removed
		{"asset:nested/deep.png", true},
		{"theme:default/bg.png", false}, // deliberately unsupported
		{"dashboard:bg.png", false},     // deliberately unsupported
		{"sunset.jpg", false},           // no scheme
		{"asset:../../etc/passwd", false},
		{"asset:script.js", false},
		{"asset:", false},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			th := testTheme("I")
			th.Background.Image = &Image{
				Light: tt.ref, Fit: "cover", Position: "center", Repeat: "no-repeat",
			}
			err := th.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected %q valid, got %v", tt.ref, err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("expected %q rejected", tt.ref)
			}
		})
	}
}

func TestValidateDarkImageRef(t *testing.T) {
	th := testTheme("I")
	th.Background.Image = &Image{
		Light: "asset:a.png", Dark: "asset:../x.png", Fit: "cover", Position: "center", Repeat: "no-repeat",
	}
	if err := th.Validate(); err == nil {
		t.Fatal("expected the dark image ref to be validated too")
	}
}

func TestValidateRequiresName(t *testing.T) {
	th := testTheme("  ")
	if err := th.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected a theme without a name to be rejected, got %v", err)
	}
}
