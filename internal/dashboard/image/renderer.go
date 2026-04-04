package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	litehtml "github.com/andresbott/litehtml-go"
)

const defaultWidth = 1024

// Renderer converts HTML to PNG images using litehtml-go.
type Renderer struct {
	customFonts map[string][]byte
}

// NewRenderer creates an image Renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		customFonts: make(map[string][]byte),
	}
}

// RegisterFont registers a custom TTF font that will be available
// to the litehtml container when rendering HTML.
func (r *Renderer) RegisterFont(family string, ttfData []byte) {
	r.customFonts[family] = ttfData
}

// Render converts the given HTML string to a PNG image.
// If width is 0, defaultWidth (1024) is used.
// If height is 0, it is auto-calculated from the rendered content.
func (r *Renderer) Render(html string, width, height int) ([]byte, error) {
	if width <= 0 {
		width = defaultWidth
	}

	initialHeight := 4096
	if height > 0 {
		initialHeight = height
	}

	container := newPNGContainer(width, initialHeight)
	for family, ttfData := range r.customFonts {
		container.registerCustomFont(family, ttfData)
	}

	doc, err := litehtml.NewDocument(html, container, "", "")
	if err != nil {
		return nil, fmt.Errorf("create litehtml document: %w", err)
	}
	defer doc.Close()

	doc.Render(float32(width))

	canvasHeight := height
	if canvasHeight <= 0 {
		canvasHeight = int(math.Ceil(float64(doc.Height())))
		if canvasHeight < 1 {
			canvasHeight = initialHeight
		}
	}

	container.img = image.NewRGBA(image.Rect(0, 0, width, canvasHeight))
	draw.Draw(container.img, container.img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	container.w = float32(width)
	container.h = float32(canvasHeight)

	clip := litehtml.Position{X: 0, Y: 0, Width: float32(width), Height: float32(canvasHeight)}
	doc.Draw(0, 0, 0, &clip)

	var buf bytes.Buffer
	if err := png.Encode(&buf, container.img); err != nil {
		return nil, fmt.Errorf("encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}
