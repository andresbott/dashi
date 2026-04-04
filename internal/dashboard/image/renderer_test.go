package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestRenderer_Render_BasicHTML(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
body { font-family: serif; margin: 10px; }
</style></head>
<body><h1>Hello</h1><p>World</p></body></html>`

	data, err := r.Render(html, 800, 0)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 {
		t.Errorf("expected width 800, got %d", bounds.Dx())
	}
	if bounds.Dy() < 1 {
		t.Error("expected positive height")
	}
}

func TestRenderer_Render_FixedHeight(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head></head><body><p>Short</p></body></html>`

	data, err := r.Render(html, 400, 300)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 {
		t.Errorf("expected width 400, got %d", bounds.Dx())
	}
	if bounds.Dy() != 300 {
		t.Errorf("expected height 300, got %d", bounds.Dy())
	}
}

func TestRenderer_Render_WithImage(t *testing.T) {
	// Create a 20x20 solid red PNG in a temp file
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "icon.png")

	redImg := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			redImg.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, redImg); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	r := NewRenderer()

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><style>
body { margin: 0; padding: 0; background: white; }
img { display: block; }
</style></head>
<body><img src="%s" style="width:20px;height:20px;"></body></html>`, imgPath)

	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that some pixels in the top-left region are red (not all white)
	hasColor := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			r, g, b, a := result.At(x, y).RGBA()
			if r > 0x8000 && g < 0x4000 && b < 0x4000 && a > 0x8000 {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("expected red pixels in the top-left region where the image should be rendered")
	}
}

func TestRenderer_Render_WithCustomFont(t *testing.T) {
	r := NewRenderer()

	// Use the Go regular font as a stand-in for a custom font.
	r.RegisterFont("test-font", goregular.TTF)

	html := `<!DOCTYPE html>
<html><head><style>
body { margin: 0; padding: 0; background: white; }
</style></head>
<body><span style="font-family: test-font; font-size: 40px; color: black;">A</span></body></html>`

	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that some pixels are dark (text was rendered with the custom font)
	hasDark := false
	for y := 0; y < 60; y++ {
		for x := 0; x < 60; x++ {
			r, g, b, _ := result.At(x, y).RGBA()
			if r < 0x4000 && g < 0x4000 && b < 0x4000 {
				hasDark = true
				break
			}
		}
		if hasDark {
			break
		}
	}
	if !hasDark {
		t.Error("expected dark pixels where custom font text should be rendered")
	}
}
