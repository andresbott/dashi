package backgrounds

import (
	"errors"
	"testing"
)

func TestValidateAcceptsEmptyBackground(t *testing.T) {
	// A freshly created background has no colour, gradient or image. That is
	// the legal empty state and renders the theme background.
	b := Background{ID: "abc123", Name: "Empty"}
	if err := b.Validate(); err != nil {
		t.Fatalf("expected empty background to be valid, got %v", err)
	}
}

func TestValidateRejectsColorAndGradientTogether(t *testing.T) {
	b := Background{
		ID:       "abc123",
		Name:     "Both",
		Color:    &Color{Light: "#ffffff"},
		Gradient: &Gradient{Direction: "to right", Light: []string{"#000000", "#ffffff"}},
	}
	err := b.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestValidateColors(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"short hex", "#fff", true},
		{"long hex", "#ff00aa", true},
		{"hex with alpha", "#ff00aa80", true},
		{"uppercase hex", "#FF00AA", true},
		{"missing hash", "ff00aa", false},
		{"named colour", "red", false},
		{"css function", "rgb(1,2,3)", false},
		{"injection attempt", "#fff;}</style><script>", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Background{ID: "abc123", Name: "C", Color: &Color{Light: tt.value}}
			err := b.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected %q to be valid, got %v", tt.value, err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("expected %q to be rejected", tt.value)
			}
		})
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
		{"angle out of range", "400deg", []string{"#000", "#fff"}, false},
		{"bogus direction", "sideways", []string{"#000", "#fff"}, false},
		{"injected direction", "to right,red);}", []string{"#000", "#fff"}, false},
		{"single stop", "to right", []string{"#000"}, false},
		{"bad stop", "to right", []string{"#000", "url(x)"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Background{
				ID:       "abc123",
				Name:     "G",
				Gradient: &Gradient{Direction: tt.direction, Light: tt.stops},
			}
			err := b.Validate()
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
	b := Background{
		ID:   "abc123",
		Name: "G",
		Gradient: &Gradient{
			Direction: "to right",
			Light:     []string{"#000", "#fff"},
			Dark:      []string{"#000", "nope"},
		},
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected dark stops to be validated too")
	}
}

func TestValidateImageEnums(t *testing.T) {
	tests := []struct {
		name           string
		fit, pos, rep  string
		valid          bool
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
			b := Background{
				ID:   "abc123",
				Name: "I",
				Image: &Image{
					Light: "shared:a.png", Fit: tt.fit, Position: tt.pos, Repeat: tt.rep,
				},
			}
			err := b.Validate()
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
		{"shared:night.png", true},
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
			b := Background{
				ID:   "abc123",
				Name: "I",
				Image: &Image{
					Light: tt.ref, Fit: "cover", Position: "center", Repeat: "no-repeat",
				},
			}
			err := b.Validate()
			if tt.valid && err != nil {
				t.Fatalf("expected %q valid, got %v", tt.ref, err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("expected %q rejected", tt.ref)
			}
		})
	}
}

func TestValidateRequiresName(t *testing.T) {
	b := Background{ID: "abc123"}
	if err := b.Validate(); err == nil {
		t.Fatal("expected a background without a name to be rejected")
	}
}
