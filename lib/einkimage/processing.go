package einkimage

import "math"

// applyExposure multiplies each RGB channel by the given factor. Alpha untouched.
func applyExposure(buf []uint8, exposure float64) {
	if exposure == 1 {
		return
	}
	for i := 0; i < len(buf); i += 4 {
		buf[i] = clampByte(float64(buf[i]) * exposure)
		buf[i+1] = clampByte(float64(buf[i+1]) * exposure)
		buf[i+2] = clampByte(float64(buf[i+2]) * exposure)
	}
}

// applyContrast stretches each channel around 128 by the given factor.
// (v - 128) * contrast + 128. Alpha untouched.
func applyContrast(buf []uint8, contrast float64) {
	if contrast == 1 {
		return
	}
	for i := 0; i < len(buf); i += 4 {
		buf[i] = clampByte((float64(buf[i])-128)*contrast + 128)
		buf[i+1] = clampByte((float64(buf[i+1])-128)*contrast + 128)
		buf[i+2] = clampByte((float64(buf[i+2])-128)*contrast + 128)
	}
}

// clamp01 clamps a float to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// clampF clamps a float to [lo, hi].
func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// applySaturation adjusts HSL saturation by the given factor. Alpha untouched.
// A factor of 0 produces grayscale; 1 is identity; >1 intensifies colors.
func applySaturation(buf []uint8, saturation float64) {
	if saturation == 1 {
		return
	}
	for i := 0; i < len(buf); i += 4 {
		r := float64(buf[i]) / 255
		g := float64(buf[i+1]) / 255
		b := float64(buf[i+2]) / 255
		max := math.Max(r, math.Max(g, b))
		min := math.Min(r, math.Min(g, b))
		lightness := (max + min) / 2

		if max == min {
			continue
		}

		delta := max - min
		var sat float64
		if lightness > 0.5 {
			sat = delta / (2 - max - min)
		} else {
			sat = delta / math.Max(max+min, 0.000001)
		}

		var hue float64
		switch max {
		case r:
			h := (g - b) / delta
			if g < b {
				h += 6
			}
			hue = h / 6
		case g:
			hue = ((b-r)/delta + 2) / 6
		default:
			hue = ((r-g)/delta + 4) / 6
		}

		newSat := clamp01(sat * saturation)
		c := (1 - math.Abs(2*lightness-1)) * newSat
		x := c * (1 - math.Abs(math.Mod(hue*6, 2)-1))
		m := lightness - c/2

		sector := int(math.Floor(hue * 6))
		var rp, gp, bp float64
		switch sector {
		case 0:
			rp, gp, bp = c, x, 0
		case 1:
			rp, gp, bp = x, c, 0
		case 2:
			rp, gp, bp = 0, c, x
		case 3:
			rp, gp, bp = 0, x, c
		case 4:
			rp, gp, bp = x, 0, c
		default:
			rp, gp, bp = c, 0, x
		}

		buf[i] = clampByte((rp + m) * 255)
		buf[i+1] = clampByte((gp + m) * 255)
		buf[i+2] = clampByte((bp + m) * 255)
	}
}

// applySCurve applies an S-curve tone remap in-place per RGB channel.
// Pixels below the midpoint use a shadow-boost power curve; pixels above
// use a highlight-compress power curve.
func applySCurve(buf []uint8, strength, shadowBoost, highlightCompress, midpoint float64) {
	if strength == 0 {
		return
	}
	mid := clampF(midpoint, 0.01, 0.99)

	for i := 0; i < len(buf); i += 4 {
		for c := 0; c < 3; c++ {
			normalized := float64(buf[i+c]) / 255
			var result float64
			if normalized <= mid {
				result = math.Pow(normalized/mid, 1-strength*shadowBoost) * mid
			} else {
				highlightVal := (normalized - mid) / (1 - mid)
				result = mid + math.Pow(highlightVal, 1+strength*highlightCompress)*(1-mid)
			}
			buf[i+c] = clampByte(result * 255)
		}
	}
}

// applyToneMapping orchestrates exposure, saturation, and contrast/s-curve
// tone shaping. Nil opts is a no-op. Zero values for Exposure, Saturation,
// and Contrast are treated as 1 (identity). Zero Strength, HighlightCompress,
// and Midpoint use documented defaults.
func applyToneMapping(buf []uint8, opts *ToneMappingOptions) {
	if opts == nil {
		return
	}
	exposure := opts.Exposure
	if exposure == 0 {
		exposure = 1
	}
	saturation := opts.Saturation
	if saturation == 0 {
		saturation = 1
	}

	applyExposure(buf, exposure)
	applySaturation(buf, saturation)

	switch opts.Mode {
	case ToneMapOff:
		// no curve
	case ToneMapSCurve:
		strength := opts.Strength
		if strength == 0 {
			strength = 0.9
		}
		highlight := opts.HighlightCompress
		if highlight == 0 {
			highlight = 1.5
		}
		midpoint := opts.Midpoint
		if midpoint == 0 {
			midpoint = 0.5
		}
		applySCurve(buf, strength, opts.ShadowBoost, highlight, midpoint)
	default: // ToneMapContrast (zero value)
		contrast := opts.Contrast
		if contrast == 0 {
			contrast = 1
		}
		applyContrast(buf, contrast)
	}
}

// applyDRC compresses image luminance into the palette's luminance range.
// Palette may be nil; Black/White fall back to [0,0,0] / [255,255,255].
func applyDRC(buf []uint8, opts *DRCOptions, palette Palette) {
	if opts == nil || opts.Mode == DRCOff {
		return
	}

	strength := opts.Strength
	if strength == 0 {
		strength = 1
	}
	strength = clamp01(strength)
	if strength == 0 {
		return
	}

	blackRGB, whiteRGB := paletteEndpoints(palette, opts.Black, opts.White, opts.BlackSet, opts.WhiteSet)
	blackLab := RGBToLab(blackRGB[0], blackRGB[1], blackRGB[2])
	whiteLab := RGBToLab(whiteRGB[0], whiteRGB[1], whiteRGB[2])
	targetRange := whiteLab[0] - blackLab[0]
	if targetRange <= 0 {
		return
	}

	sourceBlackL := 0.0
	sourceWhiteL := 100.0
	if opts.Mode == DRCAuto {
		lowP := opts.LowPercentile
		if lowP == 0 {
			lowP = 0.01
		}
		highP := opts.HighPercentile
		if highP == 0 {
			highP = 0.99
		}
		lightnesses := make([]float64, 0, len(buf)/4)
		for i := 0; i < len(buf); i += 4 {
			lab := RGBToLab(buf[i], buf[i+1], buf[i+2])
			lightnesses = append(lightnesses, lab[0])
		}
		sourceBlackL = percentile(lightnesses, lowP)
		sourceWhiteL = percentile(lightnesses, highP)
	}

	sourceRange := sourceWhiteL - sourceBlackL
	if sourceRange <= 0.0001 {
		return
	}

	for i := 0; i < len(buf); i += 4 {
		lab := RGBToLab(buf[i], buf[i+1], buf[i+2])
		normL := clamp01((lab[0] - sourceBlackL) / sourceRange)
		compressedL := blackLab[0] + normL*targetRange
		blendedL := lab[0] + (compressedL-lab[0])*strength
		r, g, b := LabToRGB([3]float64{blendedL, lab[1], lab[2]})
		buf[i] = r
		buf[i+1] = g
		buf[i+2] = b
	}
}

// paletteEndpoints picks the {black, white} RGB pair used by DRC. If user
// provided both explicit overrides, those win. Otherwise derive from the
// darkest and lightest palette entry (by luma).
func paletteEndpoints(
	palette Palette,
	black, white [3]uint8,
	blackSet, whiteSet bool,
) ([3]uint8, [3]uint8) {
	if blackSet && whiteSet {
		return black, white
	}
	if len(palette) == 0 {
		b := [3]uint8{0, 0, 0}
		w := [3]uint8{255, 255, 255}
		if blackSet {
			b = black
		}
		if whiteSet {
			w = white
		}
		return b, w
	}
	darkest := palette[0].Color
	lightest := palette[0].Color
	for _, e := range palette {
		if luma709(e.Color[0], e.Color[1], e.Color[2]) <
			luma709(darkest[0], darkest[1], darkest[2]) {
			darkest = e.Color
		}
		if luma709(e.Color[0], e.Color[1], e.Color[2]) >
			luma709(lightest[0], lightest[1], lightest[2]) {
			lightest = e.Color
		}
	}
	b := darkest
	w := lightest
	if blackSet {
		b = black
	}
	if whiteSet {
		w = white
	}
	return b, w
}

// percentile returns the pth percentile of values. p in [0, 1].
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sortFloats(sorted)
	idx := int(math.Round(float64(len(sorted)-1) * p))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func sortFloats(v []float64) {
	// insertion sort is fine: called once per DRC invocation, values small.
	for i := 1; i < len(v); i++ {
		x := v[i]
		j := i - 1
		for j >= 0 && v[j] > x {
			v[j+1] = v[j]
			j--
		}
		v[j+1] = x
	}
}

// applyLevelCompression remaps [0..255] pixel values into [black..white].
// PerChannel remaps each RGB channel independently; Luma preserves chroma
// ratio and scales by luminance.
func applyLevelCompression(buf []uint8, opts *LevelCompressionOptions) {
	if opts == nil || opts.Mode == LevelOff {
		return
	}

	autoThreshold := opts.AutoThreshold
	if autoThreshold == 0 {
		autoThreshold = 0.01
	}
	if opts.Auto {
		if !shouldEnableLevelCompression(buf, opts, autoThreshold) {
			return
		}
	}

	if opts.Mode == LevelLuma {
		blackL := toLumaScalar(opts.Black, opts.BlackSet, 0)
		whiteL := toLumaScalar(opts.White, opts.WhiteSet, 255)
		dL := whiteL - blackL
		if dL <= 0 {
			return
		}
		for i := 0; i < len(buf); i += 4 {
			r := float64(buf[i])
			g := float64(buf[i+1])
			b := float64(buf[i+2])
			y := luma709(buf[i], buf[i+1], buf[i+2])
			yNew := blackL + (y*dL)/255
			ratio := 0.0
			if y > 0 {
				ratio = yNew / y
			}
			maxChan := math.Max(r, math.Max(g, b))
			if maxChan > 0 {
				ratio = math.Min(ratio, 255/maxChan)
			}
			buf[i] = clampByte(r * ratio)
			buf[i+1] = clampByte(g * ratio)
			buf[i+2] = clampByte(b * ratio)
		}
		return
	}

	// PerChannel
	black := toRGBWithFallback(opts.Black, opts.BlackSet, 0)
	white := toRGBWithFallback(opts.White, opts.WhiteSet, 255)
	dR := float64(white[0]) - float64(black[0])
	dG := float64(white[1]) - float64(black[1])
	dB := float64(white[2]) - float64(black[2])
	if dR <= 0 || dG <= 0 || dB <= 0 {
		return
	}
	for i := 0; i < len(buf); i += 4 {
		buf[i] = clampByte(float64(black[0]) + float64(buf[i])*dR/255)
		buf[i+1] = clampByte(float64(black[1]) + float64(buf[i+1])*dG/255)
		buf[i+2] = clampByte(float64(black[2]) + float64(buf[i+2])*dB/255)
	}
}

func shouldEnableLevelCompression(buf []uint8, opts *LevelCompressionOptions, threshold float64) bool {
	pixelCount := len(buf) / 4
	if pixelCount == 0 {
		return false
	}
	outOfRange := 0
	if opts.Mode == LevelPerChannel {
		b := toRGBWithFallback(opts.Black, opts.BlackSet, 0)
		w := toRGBWithFallback(opts.White, opts.WhiteSet, 255)
		for i := 0; i < len(buf); i += 4 {
			r, g, bch := buf[i], buf[i+1], buf[i+2]
			if r < b[0] || r > w[0] || g < b[1] || g > w[1] || bch < b[2] || bch > w[2] {
				outOfRange++
			}
		}
	} else {
		bl := toLumaScalar(opts.Black, opts.BlackSet, 0)
		wh := toLumaScalar(opts.White, opts.WhiteSet, 255)
		for i := 0; i < len(buf); i += 4 {
			y := luma709(buf[i], buf[i+1], buf[i+2])
			if y < bl || y > wh {
				outOfRange++
			}
		}
	}
	return float64(outOfRange)/float64(pixelCount) >= threshold
}

func toRGBWithFallback(v [3]uint8, set bool, fallback uint8) [3]uint8 {
	if set {
		return v
	}
	return [3]uint8{fallback, fallback, fallback}
}

func toLumaScalar(v [3]uint8, set bool, fallback uint8) float64 {
	if set {
		return luma709(v[0], v[1], v[2])
	}
	return float64(fallback)
}
