package einkimage

// findClosestPaletteColor returns the palette entry's Color nearest to pixel
// under the selected distance model. If palette is empty, pixel is returned
// unchanged.
func findClosestPaletteColor(pixel [3]uint8, palette Palette, mode ColorMatchingMode) [3]uint8 {
	if len(palette) == 0 {
		return pixel
	}

	if mode == MatchLab {
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

	best := palette[0].Color
	bestDist := rgbDistSquared(best, pixel)
	for _, e := range palette[1:] {
		d := rgbDistSquared(e.Color, pixel)
		if d < bestDist {
			bestDist = d
			best = e.Color
		}
	}
	return best
}

func rgbDistSquared(a, b [3]uint8) int {
	dr := int(a[0]) - int(b[0])
	dg := int(a[1]) - int(b[1])
	db := int(a[2]) - int(b[2])
	return dr*dr + dg*dg + db*db
}
