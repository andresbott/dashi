package einkimage

// diffusionCell describes one neighbor in an error-diffusion kernel.
type diffusionCell struct {
	DX, DY int
	Factor float64
}

var diffusionKernels = map[string][]diffusionCell{
	"floydSteinberg": {
		{1, 0, 7.0 / 16}, {-1, 1, 3.0 / 16}, {0, 1, 5.0 / 16}, {1, 1, 1.0 / 16},
	},
	"falseFloydSteinberg": {
		{1, 0, 3.0 / 8}, {0, 1, 3.0 / 8}, {1, 1, 2.0 / 8},
	},
	"atkinson": {
		{1, 0, 1.0 / 8}, {2, 0, 1.0 / 8},
		{-1, 1, 1.0 / 8}, {0, 1, 1.0 / 8}, {1, 1, 1.0 / 8},
		{0, 2, 1.0 / 8},
	},
	"jarvis": {
		{1, 0, 7.0 / 48}, {2, 0, 5.0 / 48},
		{-2, 1, 3.0 / 48}, {-1, 1, 5.0 / 48}, {0, 1, 7.0 / 48}, {1, 1, 5.0 / 48}, {2, 1, 3.0 / 48},
		{-2, 2, 1.0 / 48}, {-1, 2, 3.0 / 48}, {0, 2, 5.0 / 48}, {1, 2, 3.0 / 48}, {2, 2, 1.0 / 48},
	},
	"stucki": {
		{1, 0, 8.0 / 42}, {2, 0, 4.0 / 42},
		{-2, 1, 2.0 / 42}, {-1, 1, 4.0 / 42}, {0, 1, 8.0 / 42}, {1, 1, 4.0 / 42}, {2, 1, 2.0 / 42},
		{-2, 2, 1.0 / 42}, {-1, 2, 2.0 / 42}, {0, 2, 4.0 / 42}, {1, 2, 2.0 / 42}, {2, 2, 1.0 / 42},
	},
	"burkes": {
		{1, 0, 8.0 / 32}, {2, 0, 4.0 / 32},
		{-2, 1, 2.0 / 32}, {-1, 1, 4.0 / 32}, {0, 1, 8.0 / 32}, {1, 1, 4.0 / 32}, {2, 1, 2.0 / 32},
	},
	"sierra3": {
		{1, 0, 5.0 / 32}, {2, 0, 3.0 / 32},
		{-2, 1, 2.0 / 32}, {-1, 1, 4.0 / 32}, {0, 1, 5.0 / 32}, {1, 1, 4.0 / 32}, {2, 1, 2.0 / 32},
		{-1, 2, 2.0 / 32}, {0, 2, 3.0 / 32}, {1, 2, 2.0 / 32},
	},
	"sierra2": {
		{1, 0, 4.0 / 16}, {2, 0, 3.0 / 16},
		{-2, 1, 1.0 / 16}, {-1, 1, 2.0 / 16}, {0, 1, 3.0 / 16}, {1, 1, 2.0 / 16}, {2, 1, 1.0 / 16},
	},
	"sierra2-4a": {
		{1, 0, 2.0 / 4}, {-2, 1, 1.0 / 4}, {-1, 1, 1.0 / 4},
	},
}

// getDiffusionKernel returns the named kernel. Unknown names fall back to
// floydSteinberg (matching JS behavior).
func getDiffusionKernel(name string) []diffusionCell {
	if k, ok := diffusionKernels[name]; ok {
		return k
	}
	return diffusionKernels["floydSteinberg"]
}

// applyErrorDiffusion runs error diffusion dithering in place on buf. The
// quantization error at each pixel is distributed to neighboring pixels via
// the named kernel. If serpentine is true, every other row is scanned in
// reverse and kernel X-offsets are mirrored.
func applyErrorDiffusion(
	buf []uint8, width, height int,
	palette Palette, matrixName string,
	matching ColorMatchingMode, serpentine bool,
) {
	kernel := getDiffusionKernel(matrixName)

	for y := 0; y < height; y++ {
		reverse := serpentine && y%2 == 1
		xStart, xEnd, xStep := 0, width, 1
		if reverse {
			xStart, xEnd, xStep = width-1, -1, -1
		}
		for x := xStart; x != xEnd; x += xStep {
			idx := (y*width + x) * 4
			oldPixel := [3]uint8{buf[idx], buf[idx+1], buf[idx+2]}
			newPixel := findClosestPaletteColor(oldPixel, palette, matching)
			buf[idx] = newPixel[0]
			buf[idx+1] = newPixel[1]
			buf[idx+2] = newPixel[2]

			errR := float64(oldPixel[0]) - float64(newPixel[0])
			errG := float64(oldPixel[1]) - float64(newPixel[1])
			errB := float64(oldPixel[2]) - float64(newPixel[2])

			for _, cell := range kernel {
				dx := cell.DX
				if reverse {
					dx = -dx
				}
				nx := x + dx
				ny := y + cell.DY
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				ni := (ny*width + nx) * 4
				buf[ni] = clampByte(float64(buf[ni]) + errR*cell.Factor)
				buf[ni+1] = clampByte(float64(buf[ni+1]) + errG*cell.Factor)
				buf[ni+2] = clampByte(float64(buf[ni+2]) + errB*cell.Factor)
			}
		}
	}
}
