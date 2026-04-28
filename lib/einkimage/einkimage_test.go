package einkimage

import (
	"image"
	"image/color"
	"testing"
)

func fillRGBA(img *image.RGBA, c color.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func TestDitherImageDefaultsToErrorDiffusion(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	fillRGBA(img, color.RGBA{128, 128, 128, 255})
	out, err := DitherImage(img, DitherOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// default palette is BW, so output must be a mix of black and white.
	var blacks, whites int
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			p := out.RGBAAt(x, y)
			if p.R == 0 && p.G == 0 && p.B == 0 {
				blacks++
			}
			if p.R == 255 && p.G == 255 && p.B == 255 {
				whites++
			}
		}
	}
	if blacks+whites != 64 {
		t.Errorf("all pixels must be palette colors, got b=%d w=%d", blacks, whites)
	}
	if blacks == 0 || whites == 0 {
		t.Errorf("gray must dither to a mix, got b=%d w=%d", blacks, whites)
	}
}

func TestDitherImageQuantizationOnly(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	fillRGBA(img, color.RGBA{200, 200, 200, 255})
	out, err := DitherImage(img, DitherOptions{DitheringType: QuantizationOnly})
	if err != nil {
		t.Fatal(err)
	}
	// Every pixel must snap to white (255,255,255) since RGB-nearest.
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			p := out.RGBAAt(x, y)
			if p.R != 255 || p.G != 255 || p.B != 255 {
				t.Errorf("px(%d,%d)=%v, want white", x, y, p)
			}
		}
	}
}

func TestDitherImageAppliesPreset(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	fillRGBA(img, color.RGBA{128, 128, 128, 255})
	// balanced preset sets DRC to Display with strength 1; with BW palette,
	// this should still produce black/white output.
	out, err := DitherImage(img, DitherOptions{ProcessingPreset: "balanced"})
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			p := out.RGBAAt(x, y)
			if p.R != 0 && p.R != 255 {
				t.Errorf("px(%d,%d)=%v not palette", x, y, p)
			}
		}
	}
}

func TestDitherImageNilSrc(t *testing.T) {
	_, err := DitherImage(nil, DitherOptions{})
	if err == nil {
		t.Error("DitherImage(nil) must return an error")
	}
}

func TestReplaceColorsMapsCalibratedToDeviceColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	// Source pixels match DefaultPalette's calibrated colors (0,0,0) and (255,255,255).
	img.Set(0, 0, color.RGBA{0, 0, 0, 255})
	img.Set(1, 0, color.RGBA{255, 255, 255, 255})

	out, unmatched := ReplaceColors(img, DefaultPalette)
	if unmatched != 0 {
		t.Errorf("unmatched=%d, want 0", unmatched)
	}
	// DefaultPalette DeviceColors: {0x21,0x21,0x21}, {0xe6,0xe6,0xe6}.
	if p := out.RGBAAt(0, 0); p.R != 0x21 || p.G != 0x21 || p.B != 0x21 {
		t.Errorf("px0=%v, want 0x21x3", p)
	}
	if p := out.RGBAAt(1, 0); p.R != 0xe6 || p.G != 0xe6 || p.B != 0xe6 {
		t.Errorf("px1=%v, want 0xe6x3", p)
	}
}

func TestReplaceColorsCountsUnmatched(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{0, 0, 0, 255})
	img.Set(1, 0, color.RGBA{123, 45, 67, 255}) // not a palette Color
	_, unmatched := ReplaceColors(img, DefaultPalette)
	if unmatched != 1 {
		t.Errorf("unmatched=%d, want 1", unmatched)
	}
}

func TestReplaceColorsPreservesAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{0, 0, 0, 128})
	out, _ := ReplaceColors(img, DefaultPalette)
	if p := out.RGBAAt(0, 0); p.A != 128 {
		t.Errorf("alpha changed: %d", p.A)
	}
}

func TestDitherThenReplaceColorsSpectra6(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Gradient from black to red.
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 16), 0, 0, 255})
		}
	}
	dithered, err := DitherImage(img, DitherOptions{
		Palette:              AitjcizeSpectra6Palette,
		DitheringType:        ErrorDiffusion,
		ErrorDiffusionMatrix: "floydSteinberg",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Every dithered pixel must be one of the palette calibrated colors.
	allowed := map[[3]uint8]bool{}
	for _, e := range AitjcizeSpectra6Palette {
		allowed[e.Color] = true
	}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			p := dithered.RGBAAt(x, y)
			if !allowed[[3]uint8{p.R, p.G, p.B}] {
				t.Fatalf("dithered px(%d,%d)=%v not in palette", x, y, p)
			}
		}
	}

	out, unmatched := ReplaceColors(dithered, AitjcizeSpectra6Palette)
	if unmatched != 0 {
		t.Errorf("unmatched after replaceColors: %d", unmatched)
	}
	// Output must consist of device colors (pure 0/255 per channel).
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			p := out.RGBAAt(x, y)
			for _, c := range [3]uint8{p.R, p.G, p.B} {
				if c != 0 && c != 255 {
					t.Fatalf("device px(%d,%d) channel=%d, want 0 or 255", x, y, c)
				}
			}
		}
	}
}
