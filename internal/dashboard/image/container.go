package image

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"sync"

	litehtml "github.com/andresbott/litehtml-go"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type fontEntry struct {
	face    font.Face
	ascent  float32
	descent float32
	height  float32
}

type pngContainer struct {
	img *image.RGBA
	w   float32
	h   float32

	mu         sync.Mutex
	fonts      map[uintptr]*fontEntry
	nextFontID uintptr

	regular    *opentype.Font
	bold       *opentype.Font
	italic     *opentype.Font
	boldItalic *opentype.Font
	mono       *opentype.Font
}

func newPNGContainer(w, h int) *pngContainer {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	reg, _ := opentype.Parse(goregular.TTF)
	bld, _ := opentype.Parse(gobold.TTF)
	ita, _ := opentype.Parse(goitalic.TTF)
	bi, _ := opentype.Parse(gobolditalic.TTF)
	mon, _ := opentype.Parse(gomono.TTF)

	return &pngContainer{
		img:        img,
		w:          float32(w),
		h:          float32(h),
		fonts:      make(map[uintptr]*fontEntry),
		regular:    reg,
		bold:       bld,
		italic:     ita,
		boldItalic: bi,
		mono:       mon,
	}
}

func (c *pngContainer) pickFont(family string, weight int, style litehtml.FontStyle) *opentype.Font {
	fam := strings.ToLower(family)
	if strings.Contains(fam, "mono") || strings.Contains(fam, "courier") || strings.Contains(fam, "consolas") {
		return c.mono
	}
	isBold := weight >= 700
	isItalic := style == litehtml.FontStyleItalic
	switch {
	case isBold && isItalic:
		return c.boldItalic
	case isBold:
		return c.bold
	case isItalic:
		return c.italic
	default:
		return c.regular
	}
}

func (c *pngContainer) CreateFont(descr litehtml.FontDescription) (uintptr, litehtml.FontMetrics) {
	otFont := c.pickFont(descr.Family, descr.Weight, descr.Style)
	size := float64(descr.Size)
	if size < 1 {
		size = 16
	}
	face, err := opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return 0, litehtml.FontMetrics{
			FontSize: descr.Size, Height: descr.Size * 1.2,
			Ascent: descr.Size * 0.8, Descent: descr.Size * 0.2,
			XHeight: descr.Size * 0.5, ChWidth: descr.Size * 0.6, DrawSpaces: true,
		}
	}

	met := face.Metrics()
	ascent := fixedToFloat(met.Ascent)
	descent := fixedToFloat(met.Descent)
	height := ascent + descent

	xBounds, _, _ := face.GlyphBounds('x')
	xHeight := fixedToFloat(-xBounds.Min.Y)
	chAdv, _ := face.GlyphAdvance('0')
	chWidth := fixedToFloat(chAdv)

	c.mu.Lock()
	c.nextFontID++
	id := c.nextFontID
	c.fonts[id] = &fontEntry{face: face, ascent: ascent, descent: descent, height: height}
	c.mu.Unlock()

	return id, litehtml.FontMetrics{
		FontSize: descr.Size, Height: height, Ascent: ascent, Descent: descent,
		XHeight: xHeight, ChWidth: chWidth, DrawSpaces: true,
	}
}

func (c *pngContainer) DeleteFont(hFont uintptr) {
	c.mu.Lock()
	if fe, ok := c.fonts[hFont]; ok {
		_ = fe.face.Close()
		delete(c.fonts, hFont)
	}
	c.mu.Unlock()
}

func (c *pngContainer) getFont(hFont uintptr) *fontEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fonts[hFont]
}

func (c *pngContainer) TextWidth(text string, hFont uintptr) float32 {
	fe := c.getFont(hFont)
	if fe == nil {
		return float32(len(text)) * 8
	}
	return fixedToFloat(font.MeasureString(fe.face, text))
}

func (c *pngContainer) DrawText(hdc uintptr, text string, hFont uintptr, clr litehtml.WebColor, pos litehtml.Position) {
	fe := c.getFont(hFont)
	if fe == nil {
		return
	}
	col := color.NRGBA{R: clr.Red, G: clr.Green, B: clr.Blue, A: clr.Alpha}
	d := &font.Drawer{
		Dst:  c.img,
		Src:  &image.Uniform{col},
		Face: fe.face,
		Dot:  fixed.P(int(math.Round(float64(pos.X))), int(math.Round(float64(pos.Y)))+int(math.Round(float64(fe.ascent)))),
	}
	d.DrawString(text)
}

func (c *pngContainer) DrawSolidFill(hdc uintptr, layer litehtml.BackgroundLayer, clr litehtml.WebColor) {
	col := color.NRGBA{R: clr.Red, G: clr.Green, B: clr.Blue, A: clr.Alpha}
	r := rectFromPos(layer.BorderBox)
	draw.Draw(c.img, r, &image.Uniform{col}, image.Point{}, draw.Over)
}

func (c *pngContainer) DrawBorders(hdc uintptr, borders litehtml.Borders, drawPos litehtml.Position, root bool) {
	x0 := int(math.Round(float64(drawPos.X)))
	y0 := int(math.Round(float64(drawPos.Y)))
	x1 := x0 + int(math.Round(float64(drawPos.Width)))
	y1 := y0 + int(math.Round(float64(drawPos.Height)))

	drawBorderSide := func(b litehtml.Border, r image.Rectangle) {
		if b.Width > 0 && b.Style != litehtml.BorderStyleNone {
			col := color.NRGBA{R: b.Color.Red, G: b.Color.Green, B: b.Color.Blue, A: b.Color.Alpha}
			draw.Draw(c.img, r, &image.Uniform{col}, image.Point{}, draw.Over)
		}
	}

	bwT := int(math.Round(float64(borders.Top.Width)))
	bwB := int(math.Round(float64(borders.Bottom.Width)))
	bwL := int(math.Round(float64(borders.Left.Width)))
	bwR := int(math.Round(float64(borders.Right.Width)))

	drawBorderSide(borders.Top, image.Rect(x0, y0, x1, y0+bwT))
	drawBorderSide(borders.Bottom, image.Rect(x0, y1-bwB, x1, y1))
	drawBorderSide(borders.Left, image.Rect(x0, y0, x0+bwL, y1))
	drawBorderSide(borders.Right, image.Rect(x1-bwR, y0, x1, y1))
}

func (c *pngContainer) DrawListMarker(hdc uintptr, marker litehtml.ListMarker) {
	col := color.NRGBA{R: marker.Color.Red, G: marker.Color.Green, B: marker.Color.Blue, A: marker.Color.Alpha}
	cx := int(math.Round(float64(marker.Pos.X + marker.Pos.Width/2)))
	cy := int(math.Round(float64(marker.Pos.Y + marker.Pos.Height/2)))
	radius := 3
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				c.img.Set(cx+dx, cy+dy, col)
			}
		}
	}
}

func (c *pngContainer) DrawImage(hdc uintptr, layer litehtml.BackgroundLayer, url, baseURL string) {}
func (c *pngContainer) DrawLinearGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.LinearGradient) {}
func (c *pngContainer) DrawRadialGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.RadialGradient) {}
func (c *pngContainer) DrawConicGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.ConicGradient) {}

func (c *pngContainer) PtToPx(pt float32) float32      { return pt * 96.0 / 72.0 }
func (c *pngContainer) GetDefaultFontSize() float32    { return 16 }
func (c *pngContainer) GetDefaultFontName() string     { return "serif" }
func (c *pngContainer) LoadImage(src, baseurl string, redrawOnReady bool) {}
func (c *pngContainer) GetImageSize(src, baseurl string) litehtml.Size { return litehtml.Size{} }
func (c *pngContainer) SetCaption(caption string)              {}
func (c *pngContainer) SetBaseURL(baseURL string)              {}
func (c *pngContainer) Link(href, rel, mediaType string)       {}
func (c *pngContainer) OnAnchorClick(url string)               {}
func (c *pngContainer) OnMouseEvent(event litehtml.MouseEvent) {}
func (c *pngContainer) SetCursor(cursor string)                {}
func (c *pngContainer) TransformText(text string, tt litehtml.TextTransform) string {
	switch tt {
	case litehtml.TextTransformUppercase:
		return strings.ToUpper(text)
	case litehtml.TextTransformLowercase:
		return strings.ToLower(text)
	default:
		return text
	}
}
func (c *pngContainer) ImportCSS(url, baseurl string) (string, string) { return "", baseurl }
func (c *pngContainer) SetClip(pos litehtml.Position, bdrRadius litehtml.BorderRadiuses) {}
func (c *pngContainer) DelClip() {}
func (c *pngContainer) GetViewport() litehtml.Position {
	return litehtml.Position{X: 0, Y: 0, Width: c.w, Height: c.h}
}
func (c *pngContainer) CreateElement(tagName string, attributes map[string]string) uintptr { return 0 }
func (c *pngContainer) GetMediaFeatures() litehtml.MediaFeatures {
	return litehtml.MediaFeatures{
		Type: litehtml.MediaTypeScreen, Width: c.w, Height: c.h,
		DeviceWidth: c.w, DeviceHeight: c.h, Color: 8, Resolution: 96,
	}
}
func (c *pngContainer) GetLanguage() (string, string) { return "en", "en-US" }

func fixedToFloat(v fixed.Int26_6) float32 { return float32(v) / 64.0 }

func rectFromPos(p litehtml.Position) image.Rectangle {
	x0 := int(math.Round(float64(p.X)))
	y0 := int(math.Round(float64(p.Y)))
	return image.Rect(x0, y0, x0+int(math.Round(float64(p.Width))), y0+int(math.Round(float64(p.Height))))
}
