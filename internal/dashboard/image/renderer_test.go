package image

import (
	"bytes"
	"image/png"
	"testing"
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
