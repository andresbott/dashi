package image

import (
	"image"
	"math"
)

// DitherBW applies Floyd-Steinberg dithering to produce 1-bit BW packed data.
// Output format: 1bpp, MSB first, row-major, ceil(width/8)*height bytes.
// Pixel values: 1 = white, 0 = black.
func DitherBW(img *image.RGBA) []byte {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to float64 luminance (0.0-1.0)
	luminance := make([][]float64, height)
	for y := 0; y < height; y++ {
		luminance[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// Convert from uint32 (0-65535) to float64 (0.0-1.0) using BT.601
			l := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			luminance[y][x] = l
		}
	}

	// Calculate output size
	bytesPerRow := (width + 7) / 8
	output := make([]byte, bytesPerRow*height)

	// Floyd-Steinberg dithering
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			oldVal := luminance[y][x]
			var newVal float64
			var bit byte

			// Threshold at 0.5
			if oldVal > 0.5 {
				newVal = 1.0
				bit = 1
			} else {
				newVal = 0.0
				bit = 0
			}

			// Set bit in output (MSB first)
			byteIdx := y*bytesPerRow + x/8
			bitPos := 7 - (x % 8)
			output[byteIdx] |= bit << bitPos

			// Calculate and diffuse error
			err := oldVal - newVal

			// Distribute error using Floyd-Steinberg coefficients
			if x+1 < width {
				luminance[y][x+1] += err * 7.0 / 16.0
			}
			if y+1 < height {
				if x > 0 {
					luminance[y+1][x-1] += err * 3.0 / 16.0
				}
				luminance[y+1][x] += err * 5.0 / 16.0
				if x+1 < width {
					luminance[y+1][x+1] += err * 1.0 / 16.0
				}
			}
		}
	}

	return output
}

// PackBW packs a grayscale image into 1bpp format.
// Pixels with Y > 127 are white (bit=1), else black (bit=0).
// MSB first, row-major, ceil(width/8)*height bytes.
func PackBW(img *image.Gray, width, height int) []byte {
	bytesPerRow := (width + 7) / 8
	output := make([]byte, bytesPerRow*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := img.GrayAt(x, y).Y
			var bit byte
			if gray > 127 {
				bit = 1
			}

			byteIdx := y*bytesPerRow + x/8
			bitPos := 7 - (x % 8)
			output[byteIdx] |= bit << bitPos
		}
	}

	return output
}

// rgbToLab converts an RGB color (0-255) to CIELAB.
func rgbToLab(r, g, b float64) (float64, float64, float64) {
	// sRGB to linear
	rLin := srgbToLinear(r / 255.0)
	gLin := srgbToLinear(g / 255.0)
	bLin := srgbToLinear(b / 255.0)

	// Linear RGB to XYZ (D65)
	x := 0.4124564*rLin + 0.3575761*gLin + 0.1804375*bLin
	y := 0.2126729*rLin + 0.7151522*gLin + 0.0721750*bLin
	z := 0.0193339*rLin + 0.1191920*gLin + 0.9503041*bLin

	// XYZ to Lab (D65 reference white)
	x /= 0.95047
	z /= 1.08883

	x = labF(x)
	y = labF(y)
	z = labF(z)

	L := 116.0*y - 16.0
	a := 500.0 * (x - y)
	bVal := 200.0 * (y - z)
	return L, a, bVal
}

func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func labF(t float64) float64 {
	if t > 0.008856 {
		return math.Cbrt(t)
	}
	return 7.787*t + 16.0/116.0
}

// nearestPaletteColor finds the perceptually closest palette entry using L*a*b*.
func nearestPaletteColor(r, g, b float64, paletteLab [][3]float64) (uint8, float64) {
	L, a, bVal := rgbToLab(r, g, b)
	minDist := math.MaxFloat64
	var idx uint8
	for i, pLab := range paletteLab {
		dL := L - pLab[0]
		dA := a - pLab[1]
		dB := bVal - pLab[2]
		dist := dL*dL + dA*dA + dB*dB
		if dist < minDist {
			minDist = dist
			idx = uint8(i) //nolint:gosec // palette has 7 entries
		}
	}
	return idx, minDist
}

// isAchromatic returns true if a pixel is a neutral gray (low saturation).
func isAchromatic(r, g, b float64) bool {
	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	return (maxC - minC) < 10
}

// DitherSpectra6 quantizes an image to the Spectra 6 e-ink palette (6 colors).
// Achromatic and near-palette pixels are pre-snapped to exact palette RGB values
// so they produce zero quantization error. Then standard Floyd-Steinberg dithering
// runs over the entire image, giving smooth gradients while keeping solid areas clean.
// Returns a 2D array of palette indices (0-5).
func DitherSpectra6(img *image.RGBA) [][]uint8 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	paletteRGB := [][3]float64{
		{0, 0, 0},       // 0: Black
		{255, 255, 255}, // 1: White
		{0, 128, 0},     // 2: Green
		{0, 0, 255},     // 3: Blue
		{255, 0, 0},     // 4: Red
		{255, 255, 0},   // 5: Yellow
	}

	paletteLab := make([][3]float64, len(paletteRGB))
	for i, c := range paletteRGB {
		L, a, b := rgbToLab(c[0], c[1], c[2])
		paletteLab[i] = [3]float64{L, a, b}
	}

	rgb := make([][][3]float64, height)
	for y := 0; y < height; y++ {
		rgb[y] = make([][3]float64, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			rgb[y][x] = [3]float64{
				float64(r) / 257.0,
				float64(g) / 257.0,
				float64(b) / 257.0,
			}
		}
	}

	// Pre-snap: replace achromatic (gray) pixels with exact black/white.
	// This keeps text and UI backgrounds clean during dithering.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := rgb[y][x]
			if isAchromatic(c[0], c[1], c[2]) {
				lum := 0.299*c[0] + 0.587*c[1] + 0.114*c[2]
				if lum > 128 {
					rgb[y][x] = paletteRGB[1] // white
				} else {
					rgb[y][x] = paletteRGB[0] // black
				}
			}
		}
	}

	indices := make([][]uint8, height)
	for y := 0; y < height; y++ {
		indices[y] = make([]uint8, width)
	}

	// Pre-compute rotated coordinate coefficients per channel (halftone-style angles).
	// Each channel samples noise at a different angle so dot patterns don't align.
	type screenAngle struct {
		cos, sin float64
	}
	angles := [3]screenAngle{
		{math.Cos(15.0 * math.Pi / 180.0), math.Sin(15.0 * math.Pi / 180.0)},   // R
		{math.Cos(75.0 * math.Pi / 180.0), math.Sin(75.0 * math.Pi / 180.0)},   // G
		{math.Cos(45.0 * math.Pi / 180.0), math.Sin(45.0 * math.Pi / 180.0)},   // B
	}

	const noiseAmp = 14.0

	// Deterministic triangular-PDF noise from a hash.
	tpdfNoise := func(ix, iy int) float64 {
		h := uint32(ix*7919+iy*104729+42) * 2654435761
		n1 := float64(h&0xFFFF)/65535.0*2.0 - 1.0
		h = h*2654435761 + 1
		n2 := float64(h&0xFFFF)/65535.0*2.0 - 1.0
		return (n1 + n2) * 0.5
	}

	// Floyd-Steinberg with serpentine scanning over the full image.
	for y := 0; y < height; y++ {
		leftToRight := (y%2 == 0)

		xStart, xEnd, xStep := 0, width, 1
		if !leftToRight {
			xStart, xEnd, xStep = width-1, -1, -1
		}

		fx, fy := float64(0), float64(y)
		for x := xStart; x != xEnd; x += xStep {
			oldColor := rgb[y][x]
			fx = float64(x)

			// Per-channel noise sampled at rotated coordinates.
			for ch := 0; ch < 3; ch++ {
				a := angles[ch]
				rx := int(fx*a.cos + fy*a.sin)
				ry := int(-fx*a.sin + fy*a.cos)
				oldColor[ch] = math.Max(0, math.Min(255, oldColor[ch]+tpdfNoise(rx, ry)*noiseAmp))
			}

			nearIdx, _ := nearestPaletteColor(oldColor[0], oldColor[1], oldColor[2], paletteLab)

			indices[y][x] = nearIdx
			newColor := paletteRGB[nearIdx]

			errR := oldColor[0] - newColor[0]
			errG := oldColor[1] - newColor[1]
			errB := oldColor[2] - newColor[2]

			if nx := x + xStep; nx >= 0 && nx < width {
				rgb[y][nx][0] += errR * 7.0 / 16.0
				rgb[y][nx][1] += errG * 7.0 / 16.0
				rgb[y][nx][2] += errB * 7.0 / 16.0
			}
			if y+1 < height {
				if nx := x - xStep; nx >= 0 && nx < width {
					rgb[y+1][nx][0] += errR * 3.0 / 16.0
					rgb[y+1][nx][1] += errG * 3.0 / 16.0
					rgb[y+1][nx][2] += errB * 3.0 / 16.0
				}
				rgb[y+1][x][0] += errR * 5.0 / 16.0
				rgb[y+1][x][1] += errG * 5.0 / 16.0
				rgb[y+1][x][2] += errB * 5.0 / 16.0
				if nx := x + xStep; nx >= 0 && nx < width {
					rgb[y+1][nx][0] += errR * 1.0 / 16.0
					rgb[y+1][nx][1] += errG * 1.0 / 16.0
					rgb[y+1][nx][2] += errB * 1.0 / 16.0
				}
			}
		}
	}

	return indices
}

// PackSpectra6 packs palette indices into 4bpp nibble format.
// High nibble = left pixel, low nibble = right pixel.
// Output size: (width * height) / 2 bytes.
func PackSpectra6(indices [][]uint8, width, height int) []byte {
	outputSize := (width * height) / 2
	output := make([]byte, outputSize)

	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 2 {
			highNibble := indices[y][x]
			lowNibble := uint8(0)
			if x+1 < width {
				lowNibble = indices[y][x+1]
			}
			output[idx] = (highNibble << 4) | lowNibble
			idx++
		}
	}

	return output
}
