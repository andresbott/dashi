package einkimage

import (
	"math"
	"testing"
)

func TestDiffusionKernelsAllExist(t *testing.T) {
	names := []string{
		"floydSteinberg", "falseFloydSteinberg", "atkinson",
		"jarvis", "stucki", "burkes",
		"sierra3", "sierra2", "sierra2-4a",
	}
	for _, name := range names {
		if k := getDiffusionKernel(name); len(k) == 0 {
			t.Errorf("kernel %q missing or empty", name)
		}
	}
}

func TestDiffusionKernelsSumToOne(t *testing.T) {
	names := []string{
		"floydSteinberg", "falseFloydSteinberg", "atkinson",
		"jarvis", "stucki", "burkes",
		"sierra3", "sierra2", "sierra2-4a",
	}
	for _, name := range names {
		var total float64
		for _, p := range getDiffusionKernel(name) {
			total += p.Factor
		}
		// Atkinson famously does NOT sum to 1 (sums to 6/8 = 0.75).
		// All others sum to exactly 1.
		want := 1.0
		if name == "atkinson" {
			want = 6.0 / 8.0
		}
		if math.Abs(total-want) > 1e-9 {
			t.Errorf("%s kernel sum=%v, want %v", name, total, want)
		}
	}
}

func TestDiffusionKernelUnknownFallsBackToFloyd(t *testing.T) {
	fs := getDiffusionKernel("floydSteinberg")
	unknown := getDiffusionKernel("totally-made-up-name")
	if len(unknown) != len(fs) {
		t.Errorf("unknown kernel should fall back to Floyd-Steinberg")
	}
}

func TestApplyErrorDiffusionMapsToNearestPaletteColor(t *testing.T) {
	// Solid gray should dither to a mix of black and white, not to a single color.
	w, h := 16, 16
	buf := make([]uint8, w*h*4)
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = 128, 128, 128, 255
	}
	palette := DefaultPalette
	applyErrorDiffusion(buf, w, h, palette, "floydSteinberg", MatchRGB, false)

	var blacks, whites int
	for i := 0; i < len(buf); i += 4 {
		// The palette's Color values are [0,0,0] and [255,255,255].
		if buf[i] == 0 && buf[i+1] == 0 && buf[i+2] == 0 {
			blacks++
		}
		if buf[i] == 255 && buf[i+1] == 255 && buf[i+2] == 255 {
			whites++
		}
	}
	total := w * h
	if blacks+whites != total {
		t.Errorf("all pixels should have snapped to palette colors; got b=%d w=%d/%d", blacks, whites, total)
	}
	if blacks == 0 || whites == 0 {
		t.Errorf("gray should dither to both blacks and whites; got b=%d w=%d", blacks, whites)
	}
}

func TestApplyErrorDiffusionSerpentineProducesDifferentPattern(t *testing.T) {
	w, h := 16, 16
	a := make([]uint8, w*h*4)
	b := make([]uint8, w*h*4)
	// Use a gradient instead of uniform gray to break symmetry
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			val := uint8(100 + x*5) // Gradient from left to right
			a[i], a[i+1], a[i+2], a[i+3] = val, val, val, 255
			b[i], b[i+1], b[i+2], b[i+3] = val, val, val, 255
		}
	}
	palette := DefaultPalette
	applyErrorDiffusion(a, w, h, palette, "floydSteinberg", MatchRGB, false)
	applyErrorDiffusion(b, w, h, palette, "floydSteinberg", MatchRGB, true)

	var diff int
	for i := range a {
		if a[i] != b[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Error("serpentine must produce a different pattern than straight scan")
	}
}
