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

func TestFindClosestPaletteColorEmptyPalette(t *testing.T) {
	got := findClosestPaletteColor([3]uint8{1, 2, 3}, nil, MatchRGB)
	if got != ([3]uint8{1, 2, 3}) {
		t.Errorf("empty palette must pass pixel through, got %v", got)
	}
}
