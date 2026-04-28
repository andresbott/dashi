package einkimage

import (
	"errors"
	"image"
	"math/rand"
)

// DitherImage applies tone mapping, DRC, level compression, then the selected
// dithering algorithm. The returned image's pixels are the palette's
// calibrated Color values — use ReplaceColors to translate them to
// DeviceColor before exporting to a device.
func DitherImage(src *image.RGBA, opts DitherOptions) (*image.RGBA, error) {
	if src == nil {
		return nil, errors.New("einkimage: nil src")
	}
	b := src.Bounds()
	width, height := b.Dx(), b.Dy()
	if width == 0 || height == 0 {
		return nil, errors.New("einkimage: empty image")
	}

	resolved := resolveDitherOptions(opts)
	palette := resolved.Palette
	if len(palette) == 0 {
		palette = DefaultPalette
	}

	buf := rgbaToBuffer(src)

	applyToneMapping(buf, resolved.ToneMapping)
	applyDRC(buf, resolved.DynamicRangeCompression, palette)
	applyLevelCompression(buf, resolved.LevelCompression)

	switch resolved.DitheringType {
	case QuantizationOnly:
		for i := 0; i < len(buf); i += 4 {
			newPixel := findClosestPaletteColor(
				[3]uint8{buf[i], buf[i+1], buf[i+2]},
				palette, resolved.ColorMatching,
			)
			buf[i], buf[i+1], buf[i+2] = newPixel[0], newPixel[1], newPixel[2]
		}
	case Random:
		mode := resolved.RandomDitheringType
		if mode == "" {
			mode = "blackAndWhite"
		}
		randomDither(buf, mode, rand.New(rand.NewSource(1))) //nolint:gosec // deterministic RNG for dithering, not security-sensitive
	case Ordered:
		size := resolved.OrderedDitheringMatrix
		if size == [2]int{0, 0} {
			size = [2]int{4, 4}
		}
		orderedDither(buf, width, height, palette, size, resolved.ColorMatching)
	default: // ErrorDiffusion
		matrix := resolved.ErrorDiffusionMatrix
		if matrix == "" {
			matrix = "floydSteinberg"
		}
		applyErrorDiffusion(buf, width, height, palette, matrix, resolved.ColorMatching, resolved.Serpentine)
	}

	return bufferToRGBA(buf, width, height), nil
}

// resolveDitherOptions merges the preset (if any) with user-set fields. User
// fields win over preset fields; preset fields win over zero defaults.
func resolveDitherOptions(opts DitherOptions) DitherOptions {
	out := opts
	if opts.ProcessingPreset != "" {
		if preset, ok := GetProcessingPreset(opts.ProcessingPreset); ok {
			out = mergePresetDefaults(out, preset)
		}
	}
	return out
}

// mergePresetDefaults overlays preset defaults onto out for fields not set by the user.
func mergePresetDefaults(out DitherOptions, preset ProcessingPreset) DitherOptions {
	if out.ToneMapping == nil {
		out.ToneMapping = preset.ToneMapping
	}
	if out.DynamicRangeCompression == nil {
		out.DynamicRangeCompression = preset.DRC
	}
	if out.ColorMatching == MatchRGB && preset.ColorMatching != MatchRGB {
		out.ColorMatching = preset.ColorMatching
	}
	if out.ErrorDiffusionMatrix == "" {
		out.ErrorDiffusionMatrix = preset.ErrorDiffusionMatrix
	}
	return out
}

// ReplaceColors rewrites every pixel in src whose RGB equals a palette entry's
// Color with that entry's DeviceColor. Pixels that do not match any entry are
// left unchanged; the second return value is the count of unmatched pixels.
// Alpha is preserved.
func ReplaceColors(src *image.RGBA, palette Palette) (*image.RGBA, int) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()

	lookup := make(map[[3]uint8][3]uint8, len(palette))
	for _, e := range palette {
		lookup[e.Color] = e.DeviceColor
	}

	out := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := rgbaToBuffer(src)
	unmatched := 0
	for i := 0; i < len(buf); i += 4 {
		key := [3]uint8{buf[i], buf[i+1], buf[i+2]}
		if dc, ok := lookup[key]; ok {
			buf[i], buf[i+1], buf[i+2] = dc[0], dc[1], dc[2]
		} else {
			unmatched++
		}
	}
	copy(out.Pix, buf)
	return out, unmatched
}
