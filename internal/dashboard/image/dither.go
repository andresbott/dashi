package image

import (
	"image"

	"github.com/andresbott/dashi/lib/einkimage"
)

// DitherBWRGBA returns an RGBA image dithered to black/white with muted panel-simulation colors.
// Output pixels are either {0x21,0x21,0x21} (black) or {0xe6,0xe6,0xe6} (white).
func DitherBWRGBA(img *image.RGBA) *image.RGBA {
	dithered, err := einkimage.DitherImage(img, einkimage.DitherOptions{
		Palette: einkimage.DefaultPalette,
	})
	if err != nil {
		b := img.Bounds()
		return image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	}
	out, _ := einkimage.ReplaceColors(dithered, einkimage.DefaultPalette)
	return out
}

// DitherBWPacked applies Floyd-Steinberg dithering to produce 1-bit BW packed data.
// Output format: 1bpp, MSB first, row-major, ceil(width/8)*height bytes.
// Pixel values: 1 = white, 0 = black.
func DitherBWPacked(img *image.RGBA) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dithered, err := einkimage.DitherImage(img, einkimage.DitherOptions{
		Palette: einkimage.DefaultPalette,
	})
	if err != nil {
		return make([]byte, (w+7)/8*h)
	}
	// dithered pixels are (0,0,0) black or (255,255,255) white.
	bytesPerRow := (w + 7) / 8
	out := make([]byte, bytesPerRow*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := dithered.RGBAAt(x, y)
			var bit byte
			if p.R > 127 { // white
				bit = 1
			}
			byteIdx := y*bytesPerRow + x/8
			bitPos := 7 - (x % 8)
			out[byteIdx] |= bit << bitPos
		}
	}
	return out
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

// spectra6WirePalette: palette in ESP32 wire-index order.
// einkimage.DitherImage dithers against calibrated Color.
// We read each dithered pixel's RGB (= Color) and convert to index by position.
var spectra6WirePalette = einkimage.Palette{
	{Name: "black", Color: [3]uint8{0x02, 0x02, 0x02}, DeviceColor: [3]uint8{0, 0, 0}},           // idx 0
	{Name: "white", Color: [3]uint8{0xBE, 0xC8, 0xC8}, DeviceColor: [3]uint8{255, 255, 255}},     // idx 1
	{Name: "green", Color: [3]uint8{0x27, 0x66, 0x3C}, DeviceColor: [3]uint8{0, 255, 0}},         // idx 2
	{Name: "blue", Color: [3]uint8{0x05, 0x40, 0x9E}, DeviceColor: [3]uint8{0, 0, 255}},          // idx 3
	{Name: "red", Color: [3]uint8{0x87, 0x13, 0x00}, DeviceColor: [3]uint8{255, 0, 0}},           // idx 4
	{Name: "yellow", Color: [3]uint8{0xCD, 0xCA, 0x00}, DeviceColor: [3]uint8{255, 255, 0}},      // idx 5
}

// spectra6ColorToIndex: pre-built map so we don't scan the slice per pixel.
var spectra6ColorToIndex = func() map[[3]uint8]uint8 {
	m := make(map[[3]uint8]uint8, len(spectra6WirePalette))
	for i, e := range spectra6WirePalette {
		m[e.Color] = uint8(i)
	}
	return m
}()

// DitherSpectra6RGBA returns an RGBA image dithered to the Spectra 6 palette with calibrated colors.
// Output pixels are the muted aitjcize calibrated colors.
func DitherSpectra6RGBA(img *image.RGBA) *image.RGBA {
	dithered, err := einkimage.DitherImage(img, einkimage.DitherOptions{
		Palette: spectra6WirePalette,
	})
	if err != nil {
		b := img.Bounds()
		return image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	}
	return dithered
}

// DitherSpectra6Packed applies Floyd-Steinberg dithering and packs palette indices into 4bpp format.
// Output format: 4bpp nibble-packed, high nibble=left pixel. (w*h)/2 bytes.
// Wire indices: 0=Black, 1=White, 2=Green, 3=Blue, 4=Red, 5=Yellow.
func DitherSpectra6Packed(img *image.RGBA) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dithered, err := einkimage.DitherImage(img, einkimage.DitherOptions{
		Palette: spectra6WirePalette,
	})
	if err != nil {
		return make([]byte, (w*h)/2)
	}
	// 4bpp nibble packing: high nibble = left pixel.
	outputSize := (w * h) / 2
	out := make([]byte, outputSize)
	idx := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x += 2 {
			pLeft := dithered.RGBAAt(x, y)
			highNibble := spectra6ColorToIndex[[3]uint8{pLeft.R, pLeft.G, pLeft.B}]
			var lowNibble uint8
			if x+1 < w {
				pRight := dithered.RGBAAt(x+1, y)
				lowNibble = spectra6ColorToIndex[[3]uint8{pRight.R, pRight.G, pRight.B}]
			}
			out[idx] = (highNibble << 4) | lowNibble
			idx++
		}
	}
	return out
}
