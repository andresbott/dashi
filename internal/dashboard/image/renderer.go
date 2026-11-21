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
	"strings"
	"sync"

	litehtml "github.com/andresbott/litehtml-go"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font/opentype"
)

const defaultWidth = 1024

// litehtmlMu serialises every litehtml document's lifetime across the
// process. litehtml-go is not safe for concurrent documents: its
// goGetDefaultFontName callback frees and reallocates one package-global C
// string (goDefaultFontNameBuf), so two overlapping renders double-free it
// and abort with heap corruption ("malloc(): unaligned tcache chunk"). The
// lock is package-level, not per Renderer, because that state is global.
// Drop it once litehtml-go keeps that buffer per container.
var litehtmlMu sync.Mutex

// Renderer converts HTML to images using litehtml-go.
//
// It is safe for concurrent use: fonts may be registered or replaced while
// renders are in flight. The font set is copy-on-write — a render takes the
// current map once and keeps using it, and every change swaps in a new map —
// so an in-flight render never observes a half-applied change.
type Renderer struct {
	mu sync.RWMutex
	// customFonts maps a lower-cased CSS family to its parsed font. Never
	// mutated after it is published; replaced wholesale under mu.
	customFonts map[string]*opentype.Font
}

// NewRenderer creates an image Renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		customFonts: make(map[string]*opentype.Font),
	}
}

// RegisterFont registers a custom TTF font that will be available
// to the litehtml container when rendering HTML. Data that does not parse
// as a font is ignored, as is an empty family.
func (r *Renderer) RegisterFont(family string, ttfData []byte) {
	key, f, ok := parseFont(family, ttfData)
	if !ok {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	next := make(map[string]*opentype.Font, len(r.customFonts)+1)
	for k, v := range r.customFonts {
		next[k] = v
	}
	next[key] = f
	r.customFonts = next
}

// SetFonts replaces the whole set of custom fonts with fonts (family → TTF
// bytes). Families absent from fonts are unregistered, so callers can
// re-sync after fonts were added, replaced or removed. Entries that do not
// parse as a font are skipped. Renders already in flight finish with the
// previous set; the next render sees the new one.
func (r *Renderer) SetFonts(fonts map[string][]byte) {
	next := make(map[string]*opentype.Font, len(fonts))
	for family, data := range fonts {
		if key, f, ok := parseFont(family, data); ok {
			next[key] = f
		}
	}
	r.mu.Lock()
	r.customFonts = next
	r.mu.Unlock()
}

// fontSet returns the current custom font map. The map must be treated as
// read-only.
func (r *Renderer) fontSet() map[string]*opentype.Font {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.customFonts
}

// parseFont parses ttfData once, at registration, so renders share the
// parsed font instead of re-parsing every registered font per render.
// opentype.Font is safe for concurrent use (each Face gets its own buffer).
func parseFont(family string, ttfData []byte) (string, *opentype.Font, bool) {
	if family == "" {
		return "", nil, false
	}
	f, err := opentype.Parse(ttfData)
	if err != nil {
		return "", nil, false
	}
	return strings.ToLower(family), f, true
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
	container.customFonts = r.fontSet()

	// Held until doc.Close has run (defers unwind in reverse order).
	litehtmlMu.Lock()
	defer litehtmlMu.Unlock()

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
