package einkimage

// PaletteEntry describes one palette color with both its calibrated display
// appearance (Color) and the native color to send to the display (DeviceColor).
type PaletteEntry struct {
	// Name is the canonical role used to align palettes (e.g. "black", "red").
	Name string
	// Color is the calibrated RGB that the display physically looks like.
	// Dithering matches against these values.
	Color [3]uint8
	// DeviceColor is the native RGB to send to the display.
	// ReplaceColors maps Color -> DeviceColor before export.
	DeviceColor [3]uint8
}

// Palette is an ordered list of PaletteEntry.
type Palette []PaletteEntry

// mustPalette builds a Palette from {name, color hex, deviceColor hex} triples.
// Used only during package initialization; panics on invalid hex.
func mustPalette(triples [][3]string) Palette {
	out := make(Palette, len(triples))
	for i, t := range triples {
		c, err := hexToRGB(t[1])
		if err != nil {
			panic(err)
		}
		d, err := hexToRGB(t[2])
		if err != nil {
			panic(err)
		}
		out[i] = PaletteEntry{Name: t[0], Color: c, DeviceColor: d}
	}
	return sortByCanonicalOrder(out)
}

var (
	// DefaultPalette is a two-color (black + white) palette.
	DefaultPalette = mustPalette([][3]string{
		{"black", "#000", "#212121"},
		{"white", "#fff", "#e6e6e6"},
	})

	// AitjcizeSpectra6Palette is the aitjcize calibration for Spectra 6 displays.
	AitjcizeSpectra6Palette = mustPalette([][3]string{
		{"black", "#020202", "#000000"},
		{"white", "#BEC8C8", "#FFFFFF"},
		{"blue", "#05409E", "#0000FF"},
		{"green", "#27663C", "#00FF00"},
		{"red", "#871300", "#FF0000"},
		{"yellow", "#CDCA00", "#FFFF00"},
	})

	// AcepPalette is the ACeP / Gallery 7-color palette.
	AcepPalette = mustPalette([][3]string{
		{"black", "#191E21", "#000"},
		{"white", "#F1F1F1", "#fff"},
		{"blue", "#31318F", "#0000FF"},
		{"green", "#53A428", "#00FF00"},
		{"red", "#D20E13", "#FF0000"},
		{"orange", "#B85E1C", "#FF8000"},
		{"yellow", "#F3CF11", "#FFFF00"},
	})

	// GameboyPalette is a 4-color gameboy-green palette.
	GameboyPalette = mustPalette([][3]string{
		{"gameboy0", "#0f380f", "#0F0"},
		{"gameboy1", "#306230", "#3F0"},
		{"gameboy2", "#8bac0f", "#7F0"},
		{"gameboy3", "#9bbc0f", "#FF0"},
	})
)

// GetPaletteByName returns one of the built-in palettes by name. Supported
// names: "default", "aitjcize-spectra6", "acep", "gameboy".
func GetPaletteByName(name string) (Palette, bool) {
	switch name {
	case "default":
		return DefaultPalette, true
	case "aitjcize-spectra6":
		return AitjcizeSpectra6Palette, true
	case "acep":
		return AcepPalette, true
	case "gameboy":
		return GameboyPalette, true
	}
	return nil, false
}
