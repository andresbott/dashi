package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"

	litehtml "github.com/andresbott/litehtml-go"
	xdraw "golang.org/x/image/draw"
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
// backgroundImage, if non-nil, is drawn scaled to cover the canvas before rendering HTML.
func (r *Renderer) Render(html string, width, height int, backgroundImage ...[]byte) ([]byte, error) {
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

	// Draw background image scaled to cover the canvas
	if len(backgroundImage) > 0 && len(backgroundImage[0]) > 0 {
		bgImg, _, err := image.Decode(bytes.NewReader(backgroundImage[0]))
		if err == nil {
			dstRect := container.img.Bounds()
			srcBounds := bgImg.Bounds()
			// Cover: scale to fill, preserving aspect ratio
			srcW := float64(srcBounds.Dx())
			srcH := float64(srcBounds.Dy())
			dstW := float64(dstRect.Dx())
			dstH := float64(dstRect.Dy())
			scale := math.Max(dstW/srcW, dstH/srcH)
			scaledW := int(srcW * scale)
			scaledH := int(srcH * scale)
			offsetX := (int(dstW) - scaledW) / 2
			offsetY := (int(dstH) - scaledH) / 2
			scaledRect := image.Rect(offsetX, offsetY, offsetX+scaledW, offsetY+scaledH)
			xdraw.BiLinear.Scale(container.img, scaledRect, bgImg, srcBounds, xdraw.Over, nil)
		}
	}

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
