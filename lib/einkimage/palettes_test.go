package einkimage

import "testing"

func TestPaletteEntryFields(t *testing.T) {
	e := PaletteEntry{
		Name:        "red",
		Color:       [3]uint8{98, 32, 30},
		DeviceColor: [3]uint8{255, 0, 0},
	}
	if e.Name != "red" || e.Color[0] != 98 || e.DeviceColor[0] != 255 {
		t.Errorf("PaletteEntry field access failed: %+v", e)
	}
}

func TestPaletteIsSlice(t *testing.T) {
	p := Palette{
		{Name: "black", Color: [3]uint8{0, 0, 0}, DeviceColor: [3]uint8{0, 0, 0}},
		{Name: "white", Color: [3]uint8{255, 255, 255}, DeviceColor: [3]uint8{255, 255, 255}},
	}
	if len(p) != 2 {
		t.Errorf("len(p)=%d, want 2", len(p))
	}
}

func TestDefaultPalette(t *testing.T) {
	if len(DefaultPalette) != 2 {
		t.Fatalf("DefaultPalette len=%d, want 2", len(DefaultPalette))
	}
	if DefaultPalette[0].Name != "black" {
		t.Errorf("DefaultPalette[0].Name=%q, want black", DefaultPalette[0].Name)
	}
	if DefaultPalette[0].Color != ([3]uint8{0, 0, 0}) {
		t.Errorf("DefaultPalette[0].Color=%v", DefaultPalette[0].Color)
	}
	if DefaultPalette[0].DeviceColor != ([3]uint8{0x21, 0x21, 0x21}) {
		t.Errorf("DefaultPalette[0].DeviceColor=%v", DefaultPalette[0].DeviceColor)
	}
}

func TestAitjcizeSpectra6Palette(t *testing.T) {
	if len(AitjcizeSpectra6Palette) != 6 {
		t.Fatalf("AitjcizeSpectra6Palette len=%d, want 6", len(AitjcizeSpectra6Palette))
	}
	// red entry
	var red *PaletteEntry
	for i := range AitjcizeSpectra6Palette {
		if AitjcizeSpectra6Palette[i].Name == "red" {
			red = &AitjcizeSpectra6Palette[i]
			break
		}
	}
	if red == nil {
		t.Fatal("no red entry")
	}
	if red.Color != ([3]uint8{0x87, 0x13, 0x00}) {
		t.Errorf("red.Color=%v", red.Color)
	}
	if red.DeviceColor != ([3]uint8{255, 0, 0}) {
		t.Errorf("red.DeviceColor=%v", red.DeviceColor)
	}
}

func TestAcepPaletteHasOrange(t *testing.T) {
	var hasOrange bool
	for _, e := range AcepPalette {
		if e.Name == "orange" {
			hasOrange = true
			if e.Color != ([3]uint8{0xB8, 0x5E, 0x1C}) {
				t.Errorf("orange.Color=%v", e.Color)
			}
		}
	}
	if !hasOrange {
		t.Error("AcepPalette missing orange entry")
	}
}

func TestGameboyPaletteHasFourEntries(t *testing.T) {
	if len(GameboyPalette) != 4 {
		t.Fatalf("GameboyPalette len=%d, want 4", len(GameboyPalette))
	}
}

func TestGetPaletteByName(t *testing.T) {
	tests := []struct {
		name     string
		wantLen  int
		wantOk   bool
	}{
		{"default", 2, true},
		{"aitjcize-spectra6", 6, true},
		{"acep", 7, true},
		{"gameboy", 4, true},
		{"unknown", 0, false},
	}
	for _, tc := range tests {
		got, ok := GetPaletteByName(tc.name)
		if ok != tc.wantOk {
			t.Errorf("GetPaletteByName(%q) ok=%v, want %v", tc.name, ok, tc.wantOk)
		}
		if ok && len(got) != tc.wantLen {
			t.Errorf("GetPaletteByName(%q) len=%d, want %d", tc.name, len(got), tc.wantLen)
		}
	}
}

func TestAllBuiltInPalettesAreCanonicallyOrdered(t *testing.T) {
	for name, p := range map[string]Palette{
		"default":           DefaultPalette,
		"aitjcize-spectra6": AitjcizeSpectra6Palette,
		"acep":              AcepPalette,
		"gameboy":           GameboyPalette,
	} {
		sorted := sortByCanonicalOrder(p)
		for i := range p {
			if p[i].Name != sorted[i].Name {
				t.Errorf("%s not in canonical order: got %v, want %v",
					name, namesOf(p), namesOf(sorted))
				break
			}
		}
	}
}
