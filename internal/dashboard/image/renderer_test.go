package image

import (
	"bytes"
	"encoding/base64"
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
	f, err := os.OpenFile(imgPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, redImg); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

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

func TestRenderer_Render_DefaultWidth(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body><p>Test</p></body></html>`

	// Test with width = 0 (should use default 1024)
	data, err := r.Render(html, 0, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	if img.Bounds().Dx() != 1024 {
		t.Errorf("expected default width 1024, got %d", img.Bounds().Dx())
	}
}

func TestRenderer_Render_NegativeWidth(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body><p>Test</p></body></html>`

	// Test with negative width (should use default 1024)
	data, err := r.Render(html, -1, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	if img.Bounds().Dx() != 1024 {
		t.Errorf("expected default width 1024, got %d", img.Bounds().Dx())
	}
}

func TestRenderer_Render_WithBackgroundImage(t *testing.T) {
	r := NewRenderer()

	// Create a 50x50 blue background image
	bgImg := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			bgImg.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}
	var bgBuf bytes.Buffer
	if err := png.Encode(&bgBuf, bgImg); err != nil {
		t.Fatal(err)
	}

	html := `<!DOCTYPE html>
<html><body style="margin:0;"><p style="color:white;margin:10px;">Test</p></body></html>`

	data, err := r.Render(html, 200, 200, bgBuf.Bytes())
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that background has blue pixels
	hasBlue := false
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			r, g, b, _ := result.At(x, y).RGBA()
			if b > 0x8000 && r < 0x4000 && g < 0x4000 {
				hasBlue = true
				break
			}
		}
		if hasBlue {
			break
		}
	}
	if !hasBlue {
		t.Error("expected blue background pixels")
	}
}

func TestRenderer_Render_WithEmptyBackgroundImage(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body><p>Test</p></body></html>`

	// Test with empty background image slice (should not error)
	data, err := r.Render(html, 200, 200, []byte{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithInvalidBackgroundImage(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body><p>Test</p></body></html>`

	// Test with invalid background image data (should not error, just skip background)
	data, err := r.Render(html, 200, 200, []byte("not an image"))
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithRoundedBorders(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.box {
  width: 100px;
  height: 100px;
  border: 5px solid red;
  border-radius: 20px;
  background: blue;
  margin: 10px;
}
</style></head>
<body><div class="box"></div></body></html>`

	data, err := r.Render(html, 200, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that we have red border pixels
	hasRed := false
	for y := 10; y < 120; y++ {
		for x := 10; x < 120; x++ {
			r, g, b, _ := result.At(x, y).RGBA()
			if r > 0x8000 && g < 0x4000 && b < 0x4000 {
				hasRed = true
				break
			}
		}
		if hasRed {
			break
		}
	}
	if !hasRed {
		t.Error("expected red border pixels")
	}
}

func TestRenderer_Render_WithBordersNoRadius(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.box {
  width: 50px;
  height: 50px;
  border-top: 3px solid red;
  border-right: 3px solid green;
  border-bottom: 3px solid blue;
  border-left: 3px solid yellow;
  margin: 10px;
}
</style></head>
<body><div class="box"></div></body></html>`

	data, err := r.Render(html, 150, 150)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that we have colored border pixels
	hasColor := false
	for y := 10; y < 65; y++ {
		for x := 10; x < 65; x++ {
			r, g, b, a := result.At(x, y).RGBA()
			if a > 0x8000 && (r > 0x8000 || g > 0x8000 || b > 0x8000) {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("expected colored border pixels")
	}
}

func TestRenderer_Render_WithListMarkers(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<ul>
  <li>Item 1</li>
  <li>Item 2</li>
</ul>
</body></html>`

	data, err := r.Render(html, 300, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// Verify it decodes properly
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
}

func TestRenderer_Render_WithLinearGradient(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.gradient {
  width: 200px;
  height: 100px;
  background: linear-gradient(to right, red, blue);
}
</style></head>
<body><div class="gradient"></div></body></html>`

	data, err := r.Render(html, 300, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check for gradient: left side should have more red, right side more blue
	leftRed := 0
	rightBlue := 0
	for y := 10; y < 110; y++ {
		r1, _, _, _ := result.At(20, y).RGBA()
		if r1 > 0x8000 {
			leftRed++
		}
		_, _, b2, _ := result.At(190, y).RGBA()
		if b2 > 0x8000 {
			rightBlue++
		}
	}
	if leftRed < 10 || rightBlue < 10 {
		t.Error("expected gradient with red on left and blue on right")
	}
}

func TestRenderer_Render_WithRadialGradient(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.gradient {
  width: 100px;
  height: 100px;
  background: radial-gradient(circle, yellow, red);
}
</style></head>
<body><div class="gradient"></div></body></html>`

	data, err := r.Render(html, 200, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// Verify it decodes properly
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
}

func TestRenderer_Render_WithConicGradient(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.gradient {
  width: 100px;
  height: 100px;
  background: conic-gradient(red, yellow, green, blue, red);
}
</style></head>
<body><div class="gradient"></div></body></html>`

	data, err := r.Render(html, 200, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// Verify it decodes properly
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
}

func TestRenderer_Render_WithDataURIImage(t *testing.T) {
	r := NewRenderer()

	// Create a tiny 2x2 red PNG and encode it as base64
	tinyImg := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			tinyImg.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, tinyImg); err != nil {
		t.Fatal(err)
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body>
<img src="%s" style="width:20px;height:20px;">
</body></html>`, dataURI)

	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check for red pixels where the image should be
	hasRed := false
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			r, g, b, a := result.At(x, y).RGBA()
			if r > 0x8000 && g < 0x4000 && b < 0x4000 && a > 0x8000 {
				hasRed = true
				break
			}
		}
		if hasRed {
			break
		}
	}
	if !hasRed {
		t.Error("expected red pixels from data URI image")
	}
}

func TestRenderer_Render_WithTextTransform(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<p style="text-transform: uppercase;">lowercase text</p>
<p style="text-transform: lowercase;">UPPERCASE TEXT</p>
</body></html>`

	data, err := r.Render(html, 300, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// Verify it decodes properly
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
}

func TestRenderer_Render_WithDifferentFontWeights(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<p style="font-weight: 400;">Normal</p>
<p style="font-weight: 700;">Bold</p>
<p style="font-style: italic;">Italic</p>
<p style="font-weight: 700; font-style: italic;">Bold Italic</p>
<p style="font-family: monospace;">Monospace</p>
</body></html>`

	data, err := r.Render(html, 300, 300)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// Verify it decodes properly
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
}

func TestRenderer_Render_WithInvalidDataURI(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<img src="data:image/png;base64,invalidbase64data!!!">
</body></html>`

	// Should not error, just skip the invalid image
	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithDataURINoComma(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<img src="data:image/png;base64">
</body></html>`

	// Should not error, just skip the invalid data URI
	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithUnsupportedImageExtension(t *testing.T) {
	r := NewRenderer()

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.bmp")
	if err := os.WriteFile(imgPath, []byte("fake image"), 0o600); err != nil {
		t.Fatal(err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body>
<img src="%s">
</body></html>`, imgPath)

	// Should not error, just skip unsupported images
	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithNonexistentImage(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><body>
<img src="/nonexistent/image.png">
</body></html>`

	// Should not error, just skip missing images
	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithCorruptImageFile(t *testing.T) {
	r := NewRenderer()

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "corrupt.png")
	if err := os.WriteFile(imgPath, []byte("not a real png"), 0o600); err != nil {
		t.Fatal(err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body>
<img src="%s">
</body></html>`, imgPath)

	// Should not error, just skip corrupt images
	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithJPEGImage(t *testing.T) {
	r := NewRenderer()

	dir := t.TempDir()
	// Create a minimal valid JPEG (1x1 pixel, white)
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
		0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
		0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
		0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
		0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x14, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x03, 0xFF, 0xC4, 0x00, 0x14, 0x10, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00,
		0x3F, 0xFF, 0xD9,
	}
	imgPath := filepath.Join(dir, "test.jpg")
	if err := os.WriteFile(imgPath, jpegData, 0o600); err != nil {
		t.Fatal(err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body>
<img src="%s">
</body></html>`, imgPath)

	data, err := r.Render(html, 100, 100)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestRenderer_Render_WithComplexRoundedCorners(t *testing.T) {
	r := NewRenderer()

	html := `<!DOCTYPE html>
<html><head><style>
.box {
  width: 80px;
  height: 80px;
  border-top-left-radius: 30px 20px;
  border-top-right-radius: 10px 30px;
  border-bottom-right-radius: 40px 10px;
  border-bottom-left-radius: 15px 25px;
  background: green;
  border: 4px solid orange;
  margin: 10px;
}
</style></head>
<body><div class="box"></div></body></html>`

	data, err := r.Render(html, 200, 200)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}

	// Check that we have green background pixels
	hasGreen := false
	for y := 20; y < 100; y++ {
		for x := 20; x < 100; x++ {
			r, g, b, _ := result.At(x, y).RGBA()
			if g > 0x8000 && r < 0x4000 && b < 0x4000 {
				hasGreen = true
				break
			}
		}
		if hasGreen {
			break
		}
	}
	if !hasGreen {
		t.Error("expected green background pixels in rounded box")
	}
}

