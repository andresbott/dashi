package einkimage

import "image"

// rgbaToBuffer copies the pixels of img into a dense []uint8 of length w*h*4
// (RGBA stride 4), matching the JS Uint8ClampedArray layout used by the
// port. It handles non-zero image origins via img.RGBAAt.
func rgbaToBuffer(img *image.RGBA) []uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]uint8, w*h*4)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			out[i] = p.R
			out[i+1] = p.G
			out[i+2] = p.B
			out[i+3] = p.A
			i += 4
		}
	}
	return out
}

// bufferToRGBA builds an *image.RGBA from a dense working buffer.
func bufferToRGBA(buf []uint8, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// image.NewRGBA uses stride = width*4; copy directly.
	copy(img.Pix, buf)
	return img
}
