package einkimage

// paletteMatcher finds the nearest palette colour for many pixels. It
// converts the palette to L*a*b* once instead of once per pixel, which is
// most of the cost of MatchLab on a large image; the results are identical
// to findClosestPaletteColor.
type paletteMatcher struct {
	palette Palette
	mode    ColorMatchingMode
	labs    [][3]float64 // palette Colors in L*a*b*; only set for MatchLab
}

func newPaletteMatcher(palette Palette, mode ColorMatchingMode) *paletteMatcher {
	m := &paletteMatcher{palette: palette, mode: mode}
	if mode == MatchLab {
		m.labs = make([][3]float64, len(palette))
		for i, e := range palette {
			m.labs[i] = RGBToLab(e.Color[0], e.Color[1], e.Color[2])
		}
	}
	return m
}

// closest returns the palette entry's Color nearest to pixel under the
// matcher's distance model. If the palette is empty, pixel is returned
// unchanged.
func (m *paletteMatcher) closest(pixel [3]uint8) [3]uint8 {
	if len(m.palette) == 0 {
		return pixel
	}

	if m.mode == MatchLab {
		pixLab := RGBToLab(pixel[0], pixel[1], pixel[2])
		best := m.palette[0].Color
		bestDist := DeltaE(m.labs[0], pixLab)
		for i, e := range m.palette[1:] {
			d := DeltaE(m.labs[i+1], pixLab)
			if d < bestDist {
				bestDist = d
				best = e.Color
			}
		}
		return best
	}

	best := m.palette[0].Color
	bestDist := rgbDistSquared(best, pixel)
	for _, e := range m.palette[1:] {
		d := rgbDistSquared(e.Color, pixel)
		if d < bestDist {
			bestDist = d
			best = e.Color
		}
	}
	return best
}

// findClosestPaletteColor returns the palette entry's Color nearest to pixel
// under the selected distance model. If palette is empty, pixel is returned
// unchanged. For many pixels against one palette, use a paletteMatcher.
func findClosestPaletteColor(pixel [3]uint8, palette Palette, mode ColorMatchingMode) [3]uint8 {
	return newPaletteMatcher(palette, mode).closest(pixel)
}

func rgbDistSquared(a, b [3]uint8) int {
	dr := int(a[0]) - int(b[0])
	dg := int(a[1]) - int(b[1])
	db := int(a[2]) - int(b[2])
	return dr*dr + dg*dg + db*db
}
