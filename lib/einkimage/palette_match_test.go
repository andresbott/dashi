package einkimage

import "testing"

func TestFindClosestPaletteColorRGB(t *testing.T) {
	p := Palette{
		{Name: "black", Color: [3]uint8{0, 0, 0}},
		{Name: "white", Color: [3]uint8{255, 255, 255}},
		{Name: "red", Color: [3]uint8{255, 0, 0}},
	}
	if got := findClosestPaletteColor([3]uint8{10, 5, 5}, p, MatchRGB); got != ([3]uint8{0, 0, 0}) {
		t.Errorf("near-black matched to %v", got)
	}
	if got := findClosestPaletteColor([3]uint8{250, 250, 250}, p, MatchRGB); got != ([3]uint8{255, 255, 255}) {
		t.Errorf("near-white matched to %v", got)
	}
	if got := findClosestPaletteColor([3]uint8{200, 20, 20}, p, MatchRGB); got != ([3]uint8{255, 0, 0}) {
		t.Errorf("near-red matched to %v", got)
	}
}

func TestFindClosestPaletteColorLAB(t *testing.T) {
	p := Palette{
		{Name: "black", Color: [3]uint8{0, 0, 0}},
		{Name: "red", Color: [3]uint8{255, 0, 0}},
		{Name: "green", Color: [3]uint8{0, 255, 0}},
	}
	got := findClosestPaletteColor([3]uint8{220, 40, 40}, p, MatchLab)
	if got != ([3]uint8{255, 0, 0}) {
		t.Errorf("LAB near-red matched to %v", got)
	}
}

// naiveClosestLab is the matcher before palette Labs were precomputed: it
// converts every palette entry per pixel. The matcher must agree with it.
func naiveClosestLab(pixel [3]uint8, palette Palette) [3]uint8 {
	pixLab := RGBToLab(pixel[0], pixel[1], pixel[2])
	best := palette[0].Color
	bestDist := DeltaE(RGBToLab(best[0], best[1], best[2]), pixLab)
	for _, e := range palette[1:] {
		d := DeltaE(RGBToLab(e.Color[0], e.Color[1], e.Color[2]), pixLab)
		if d < bestDist {
			bestDist = d
			best = e.Color
		}
	}
	return best
}

func TestPaletteMatcherLabMatchesNaive(t *testing.T) {
	for _, p := range []Palette{DefaultPalette, AitjcizeSpectra6Palette, AcepPalette, GameboyPalette} {
		m := newPaletteMatcher(p, MatchLab)
		for r := 0; r < 256; r += 7 {
			for g := 0; g < 256; g += 11 {
				for b := 0; b < 256; b += 13 {
					px := [3]uint8{uint8(r), uint8(g), uint8(b)}
					if got, want := m.closest(px), naiveClosestLab(px, p); got != want {
						t.Fatalf("pixel %v: matcher %v, naive %v", px, got, want)
					}
				}
			}
		}
	}
}

func TestFindClosestPaletteColorEmptyPalette(t *testing.T) {
	got := findClosestPaletteColor([3]uint8{1, 2, 3}, nil, MatchRGB)
	if got != ([3]uint8{1, 2, 3}) {
		t.Errorf("empty palette must pass pixel through, got %v", got)
	}
}
