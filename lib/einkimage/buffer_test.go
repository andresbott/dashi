package einkimage

import (
	"image"
	"image/color"
	"testing"
)

func TestRgbaToBuffer(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	img.Set(1, 0, color.RGBA{4, 5, 6, 128})
	buf := rgbaToBuffer(img)
	want := []uint8{1, 2, 3, 255, 4, 5, 6, 128}
	if len(buf) != len(want) {
		t.Fatalf("len=%d, want %d", len(buf), len(want))
	}
	for i := range want {
		if buf[i] != want[i] {
			t.Errorf("buf[%d]=%d, want %d", i, buf[i], want[i])
		}
	}
}

func TestRgbaToBufferRespectsBounds(t *testing.T) {
	// *image.RGBA can have a non-zero origin; the working buffer must be dense.
	full := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			full.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}
	sub := full.SubImage(image.Rect(1, 0, 3, 2)).(*image.RGBA)
	buf := rgbaToBuffer(sub)
	if len(buf) != 2*2*4 {
		t.Fatalf("sub len=%d, want 16", len(buf))
	}
	// Top-left of sub should be the pixel at (1, 0).
	if buf[0] != 1 || buf[1] != 0 {
		t.Errorf("sub[0,0]=%v,%v, want 1,0", buf[0], buf[1])
	}
}

func TestBufferToRgba(t *testing.T) {
	buf := []uint8{10, 20, 30, 255, 40, 50, 60, 200}
	img := bufferToRGBA(buf, 2, 1)
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 1 {
		t.Fatalf("bounds=%v", img.Bounds())
	}
	p := img.RGBAAt(0, 0)
	if p.R != 10 || p.G != 20 || p.B != 30 || p.A != 255 {
		t.Errorf("px0=%+v", p)
	}
	p = img.RGBAAt(1, 0)
	if p.R != 40 || p.A != 200 {
		t.Errorf("px1=%+v", p)
	}
}
