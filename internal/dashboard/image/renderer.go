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

// Renderer converts HTML to images using litehtml-go.
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

// RenderToImage converts the given HTML string to an *image.RGBA.
// If width is 0, defaultWidth (1024) is used.
// If height is 0, it is auto-calculated from the rendered content.
// backgroundImage, if non-nil, is drawn scaled to cover the canvas before rendering HTML.
func (r *Renderer) RenderToImage(html string, width, height int, backgroundImage ...[]byte) (*image.RGBA, error) {
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

	if len(backgroundImage) > 0 && len(backgroundImage[0]) > 0 {
		bgImg, _, decErr := image.Decode(bytes.NewReader(backgroundImage[0]))
		if decErr == nil {
			dstRect := container.img.Bounds()
			srcBounds := bgImg.Bounds()
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

	return container.img, nil
}

// RotateImage rotates an RGBA image by the given degrees (0, 90, 180, 270).
// The rotation maps pixel (0,0) to the panel's native top-left.
func RotateImage(src *image.RGBA, degrees int) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	switch degrees {
	case 90:
		dst := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.SetRGBA(h-1-y, x, src.RGBAAt(x, y))
			}
		}
		return dst
	case 180:
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.SetRGBA(w-1-x, h-1-y, src.RGBAAt(x, y))
			}
		}
		return dst
	case 270:
		dst := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.SetRGBA(y, w-1-x, src.RGBAAt(x, y))
			}
		}
		return dst
	default:
		return src
	}
}

// Render converts the given HTML string to PNG bytes.
func (r *Renderer) Render(html string, width, height int, backgroundImage ...[]byte) ([]byte, error) {
	img, err := r.RenderToImage(html, width, height, backgroundImage...)
	if err != nil {
		return nil, err
	}
	return EncodePNG(img)
}

// EncodePNG encodes an RGBA image as PNG bytes.
func EncodePNG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}
