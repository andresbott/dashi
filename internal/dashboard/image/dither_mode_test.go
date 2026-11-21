package image

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/andresbott/dashi/lib/einkimage"
)

// photoLike returns a smooth two-axis colour gradient: enough tonal range for
// tone mapping and range compression to change what a preset dithers to.
func photoLike(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: uint8((x + y) * 127 / (w + h)),
				A: 255,
			})
		}
	}
	return img
}

var bwDeviceColors = map[[3]uint8]bool{{0x21, 0x21, 0x21}: true, {0xe6, 0xe6, 0xe6}: true}

// allModes is every value a dashboard may store, plus an unknown one.
func allModes() []string {
	return append([]string{DitherPlain, DitherAuto, "bogus"}, einkimage.GetProcessingPresetNames()...)
}

// Plain must be exactly what the request path produced before dither modes:
// DitherImage with only the palette set.
func TestDitherPlainIsUnchanged(t *testing.T) {
	img := photoLike(64, 48)

	wantBW, err := einkimage.DitherImage(img, einkimage.DitherOptions{Palette: einkimage.DefaultPalette})
	if err != nil {
		t.Fatal(err)
	}
	wantBW, _ = einkimage.ReplaceColors(wantBW, einkimage.DefaultPalette)
	if got := DitherBWRGBA(img, DitherPlain); !bytes.Equal(got.Pix, wantBW.Pix) {
		t.Error("plain png-bw differs from palette-only DitherImage")
	}

	wantS6, err := einkimage.DitherImage(img, einkimage.DitherOptions{Palette: spectra6WirePalette})
	if err != nil {
		t.Fatal(err)
	}
	if got := DitherSpectra6RGBA(img, DitherPlain); !bytes.Equal(got.Pix, wantS6.Pix) {
		t.Error("plain png-spectra6 differs from palette-only DitherImage")
	}
}

// The packed formats pack exactly the pixels the PNG variants show, for every
// mode — so pinning the PNG output pins the wire bytes too.
func TestDitherPackedMatchesRGBA(t *testing.T) {
	img := photoLike(30, 20) // even width: spectra6 packs two pixels per byte
	for _, mode := range allModes() {
		t.Run("mode="+mode, func(t *testing.T) {
			rgba := DitherBWRGBA(img, mode)
			packed := DitherBWPacked(img, mode)
			bytesPerRow := (30 + 7) / 8
			for y := 0; y < 20; y++ {
				for x := 0; x < 30; x++ {
					bit := packed[y*bytesPerRow+x/8] >> (7 - x%8) & 1
					white := rgba.RGBAAt(x, y).R == 0xe6
					if (bit == 1) != white {
						t.Fatalf("bw pixel (%d,%d): packed bit %d, png white=%v", x, y, bit, white)
					}
				}
			}

			s6 := DitherSpectra6RGBA(img, mode)
			s6Packed := DitherSpectra6Packed(img, mode)
			for y := 0; y < 20; y++ {
				for x := 0; x < 30; x++ {
					b := s6Packed[(y*30+x)/2]
					nibble := b >> 4
					if x%2 == 1 {
						nibble = b & 0x0F
					}
					p := s6.RGBAAt(x, y)
					if want := spectra6ColorToIndex[[3]uint8{p.R, p.G, p.B}]; nibble != want {
						t.Fatalf("spectra6 pixel (%d,%d): packed %d, png index %d", x, y, nibble, want)
					}
				}
			}
		})
	}
}

// Whatever the mode, every output pixel is a palette colour, so ReplaceColors
// and the wire-index table keep working.
func TestDitherModesOutputPaletteColors(t *testing.T) {
	img := photoLike(40, 30)
	for _, mode := range allModes() {
		t.Run("mode="+mode, func(t *testing.T) {
			bw := DitherBWRGBA(img, mode)
			s6 := DitherSpectra6RGBA(img, mode)
			for y := 0; y < 30; y++ {
				for x := 0; x < 40; x++ {
					p := bw.RGBAAt(x, y)
					if !bwDeviceColors[[3]uint8{p.R, p.G, p.B}] {
						t.Fatalf("bw pixel (%d,%d) = %v is not a device colour", x, y, p)
					}
					q := s6.RGBAAt(x, y)
					if _, ok := spectra6ColorToIndex[[3]uint8{q.R, q.G, q.B}]; !ok {
						t.Fatalf("spectra6 pixel (%d,%d) = %v is not a calibrated colour", x, y, q)
					}
				}
			}
		})
	}
}

func TestDitherPresetChangesOutput(t *testing.T) {
	img := photoLike(64, 48)
	plain := DitherSpectra6RGBA(img, DitherPlain)
	for _, preset := range einkimage.GetProcessingPresetNames() {
		if got := DitherSpectra6RGBA(img, preset); bytes.Equal(got.Pix, plain.Pix) {
			t.Errorf("preset %q dithered identically to plain", preset)
		}
	}
}

func TestDitherUnknownModeIsPlain(t *testing.T) {
	img := photoLike(32, 32)
	if !bytes.Equal(DitherBWPacked(img, "bogus"), DitherBWPacked(img, DitherPlain)) {
		t.Error("unknown bw mode should dither like plain")
	}
	if !bytes.Equal(DitherSpectra6Packed(img, "bogus"), DitherSpectra6Packed(img, DitherPlain)) {
		t.Error("unknown spectra6 mode should dither like plain")
	}
}

func TestDitherOptions(t *testing.T) {
	img := photoLike(32, 32)

	t.Run("plain sets only the palette", func(t *testing.T) {
		opts := ditherOptions(img, spectra6WirePalette, DitherPlain)
		if opts.ProcessingPreset != "" || opts.ToneMapping != nil || opts.DynamicRangeCompression != nil {
			t.Errorf("plain options carry processing: %+v", opts)
		}
	})

	t.Run("preset names the preset", func(t *testing.T) {
		opts := ditherOptions(img, spectra6WirePalette, "vivid")
		if opts.ProcessingPreset != "vivid" {
			t.Errorf("preset = %q, want vivid", opts.ProcessingPreset)
		}
	})

	t.Run("auto keeps the target palette", func(t *testing.T) {
		for _, pal := range []einkimage.Palette{einkimage.DefaultPalette, spectra6WirePalette} {
			opts := ditherOptions(img, pal, DitherAuto)
			if len(opts.Palette) != len(pal) || opts.Palette[0] != pal[0] {
				t.Errorf("auto replaced the palette: %v", opts.Palette)
			}
			if opts.ProcessingPreset == "" {
				t.Error("auto produced no processing preset")
			}
		}
	})

	t.Run("auto picks grayscale for a two-colour panel", func(t *testing.T) {
		if got := ditherOptions(img, einkimage.DefaultPalette, DitherAuto).ProcessingPreset; got != "grayscale" {
			t.Errorf("auto bw preset = %q, want grayscale", got)
		}
	})
}

// dashboardImage renders a small dashboard-like page at the common 7.5"
// panel size, so the benchmarks classify the kind of image the request path
// actually dithers (text and flat UI, not a photo).
func dashboardImage(b *testing.B) *image.RGBA {
	b.Helper()
	const html = `<html><body style="margin:0;background:#ffffff;font-family:Inter">
<div style="padding:16px">
<h1 style="font-size:48px;color:#222">12:45</h1>
<p style="font-size:20px;color:#555">Saturday, 21 November</p>
<div style="background:#0a84ff;color:#fff;padding:12px;width:300px">Weather 14°C cloudy</div>
<div style="background:#e0e0e5;height:120px;margin-top:12px"></div>
<p style="font-size:16px;color:#dc2626">Train S3 delayed +4'</p>
</div></body></html>`
	img, err := NewRenderer().RenderToImage(html, 800, 480)
	if err != nil {
		b.Fatal(err)
	}
	return img
}

// BenchmarkDither800x480 compares the per-render cost of each mode on an
// 800×480 dashboard. Auto adds one SuggestProcessingOptions classification.
func BenchmarkDither800x480(b *testing.B) {
	img := dashboardImage(b)
	for _, mode := range []string{DitherPlain, DitherAuto, "grayscale", "vivid"} {
		name := mode
		if name == "" {
			name = "plain"
		}
		b.Run("bw/"+name, func(b *testing.B) {
			for b.Loop() {
				DitherBWPacked(img, mode)
			}
		})
		b.Run("spectra6/"+name, func(b *testing.B) {
			for b.Loop() {
				DitherSpectra6Packed(img, mode)
			}
		})
	}
}

// BenchmarkSuggestProcessingOptions800x480 isolates the classification that
// the auto mode adds to every render.
func BenchmarkSuggestProcessingOptions800x480(b *testing.B) {
	img := dashboardImage(b)
	for b.Loop() {
		einkimage.SuggestProcessingOptions(img, spectra6WirePalette, einkimage.SuggestInput{})
	}
}
