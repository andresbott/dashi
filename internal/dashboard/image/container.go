package image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	stddraw "image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "embed"

	litehtml "github.com/andresbott/litehtml-go"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
)

//go:embed Inter-Regular.ttf
var interRegularTTF []byte

//go:embed Inter-Bold.ttf
var interBoldTTF []byte

//go:embed Inter-Italic.ttf
var interItalicTTF []byte

//go:embed Inter-BoldItalic.ttf
var interBoldItalicTTF []byte

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

	images      map[string]image.Image
	customFonts map[string]*opentype.Font

	regular    *opentype.Font
	bold       *opentype.Font
	italic     *opentype.Font
	boldItalic *opentype.Font
	mono       *opentype.Font
}

func newPNGContainer(w, h int) *pngContainer {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	stddraw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, stddraw.Src)

	reg, _ := opentype.Parse(interRegularTTF)
	bld, _ := opentype.Parse(interBoldTTF)
	ita, _ := opentype.Parse(interItalicTTF)
	bi, _ := opentype.Parse(interBoldItalicTTF)
	mon, _ := opentype.Parse(gomono.TTF)

	return &pngContainer{
		img:         img,
		w:           float32(w),
		h:           float32(h),
		fonts:       make(map[uintptr]*fontEntry),
		images:      make(map[string]image.Image),
		customFonts: make(map[string]*opentype.Font),
		regular:     reg,
		bold:        bld,
		italic:      ita,
		boldItalic:  bi,
		mono:        mon,
	}
}

func (c *pngContainer) registerCustomFont(family string, ttfData []byte) {
	f, err := opentype.Parse(ttfData)
	if err != nil {
		return
	}
	c.customFonts[strings.ToLower(family)] = f
}

func (c *pngContainer) pickFont(family string, weight int, style litehtml.FontStyle) *opentype.Font {
	fam := strings.ToLower(family)

	// Check custom fonts first
	if f, ok := c.customFonts[fam]; ok {
		return f
	}

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
	if !hasRadius(layer.BorderRadius) {
		stddraw.Draw(c.img, r, &image.Uniform{col}, image.Point{}, stddraw.Over)
		return
	}
	mask := image.NewAlpha(r)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			if isInsideRoundedRect(px, py, layer.BorderBox, layer.BorderRadius) {
				mask.SetAlpha(px, py, color.Alpha{A: 255})
			}
		}
	}
	stddraw.DrawMask(c.img, r, &image.Uniform{col}, image.Point{}, mask, r.Min, stddraw.Over)
}

func (c *pngContainer) DrawBorders(hdc uintptr, borders litehtml.Borders, drawPos litehtml.Position, root bool) {
	x0 := int(math.Round(float64(drawPos.X)))
	y0 := int(math.Round(float64(drawPos.Y)))
	x1 := x0 + int(math.Round(float64(drawPos.Width)))
	y1 := y0 + int(math.Round(float64(drawPos.Height)))

	if !hasRadius(borders.Radius) {
		drawBorderSide := func(b litehtml.Border, r image.Rectangle) {
			if b.Width > 0 && b.Style != litehtml.BorderStyleNone {
				col := color.NRGBA{R: b.Color.Red, G: b.Color.Green, B: b.Color.Blue, A: b.Color.Alpha}
				stddraw.Draw(c.img, r, &image.Uniform{col}, image.Point{}, stddraw.Over)
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
		return
	}

	// Rounded border: draw border pixels between outer and inner rounded rects,
	// using diagonal quadrants to determine which side each pixel belongs to.
	r := image.Rect(x0, y0, x1, y1)
	borderImg := image.NewNRGBA(r)

	bwT := float64(borders.Top.Width)
	bwR := float64(borders.Right.Width)
	bwB := float64(borders.Bottom.Width)
	bwL := float64(borders.Left.Width)

	innerPos := litehtml.Position{
		X:      drawPos.X + float32(bwL),
		Y:      drawPos.Y + float32(bwT),
		Width:  drawPos.Width - float32(bwL+bwR),
		Height: drawPos.Height - float32(bwT+bwB),
	}
	innerRadius := litehtml.BorderRadiuses{
		TopLeftX:     maxF32(0, borders.Radius.TopLeftX-float32(bwL)),
		TopLeftY:     maxF32(0, borders.Radius.TopLeftY-float32(bwT)),
		TopRightX:    maxF32(0, borders.Radius.TopRightX-float32(bwR)),
		TopRightY:    maxF32(0, borders.Radius.TopRightY-float32(bwT)),
		BottomRightX: maxF32(0, borders.Radius.BottomRightX-float32(bwR)),
		BottomRightY: maxF32(0, borders.Radius.BottomRightY-float32(bwB)),
		BottomLeftX:  maxF32(0, borders.Radius.BottomLeftX-float32(bwL)),
		BottomLeftY:  maxF32(0, borders.Radius.BottomLeftY-float32(bwB)),
	}

	w := float64(x1 - x0)
	h := float64(y1 - y0)

	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if !isInsideRoundedRect(px, py, drawPos, borders.Radius) {
				continue
			}
			if isInsideRoundedRect(px, py, innerPos, innerRadius) {
				continue
			}
			relX := float64(px-x0) + 0.5
			relY := float64(py-y0) + 0.5
			d1 := relY*w - relX*h
			d2 := relY*w + relX*h - w*h

			var b litehtml.Border
			switch {
			case d1 <= 0 && d2 <= 0:
				b = borders.Top
			case d1 >= 0 && d2 >= 0:
				b = borders.Bottom
			case d1 >= 0 && d2 <= 0:
				b = borders.Left
			default:
				b = borders.Right
			}
			if b.Width > 0 && b.Style != litehtml.BorderStyleNone {
				borderImg.SetNRGBA(px, py, color.NRGBA{R: b.Color.Red, G: b.Color.Green, B: b.Color.Blue, A: b.Color.Alpha})
			}
		}
	}
	stddraw.Draw(c.img, r, borderImg, r.Min, stddraw.Over)
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

func (c *pngContainer) DrawImage(hdc uintptr, layer litehtml.BackgroundLayer, url, baseURL string) {
	srcImg, ok := c.images[url]
	if !ok {
		return
	}
	dstRect := rectFromPos(layer.BorderBox)
	if dstRect.Dx() <= 0 || dstRect.Dy() <= 0 {
		return
	}
	draw.BiLinear.Scale(c.img, dstRect, srcImg, srcImg.Bounds(), draw.Over, nil)
}
func (c *pngContainer) DrawLinearGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.LinearGradient) {
	if len(gradient.ColorPoints) == 0 {
		return
	}
	r := rectFromPos(layer.BorderBox)
	if r.Dx() <= 0 || r.Dy() <= 0 {
		return
	}
	gradImg := image.NewNRGBA(r)
	dx := float64(gradient.End.X - gradient.Start.X)
	dy := float64(gradient.End.Y - gradient.Start.Y)
	lenSq := dx*dx + dy*dy
	sx := float64(gradient.Start.X)
	sy := float64(gradient.Start.Y)
	rad := hasRadius(layer.BorderRadius)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			if rad && !isInsideRoundedRect(px, py, layer.BorderBox, layer.BorderRadius) {
				continue
			}
			var t float32
			if lenSq < 0.001 {
				t = 1.0
			} else {
				vx := float64(px) - sx
				vy := float64(py) - sy
				t = float32((vx*dx + vy*dy) / lenSq)
			}
			gradImg.SetNRGBA(px, py, interpolateGradient(gradient.ColorPoints, t))
		}
	}
	stddraw.Draw(c.img, r, gradImg, r.Min, stddraw.Over)
}

func (c *pngContainer) DrawRadialGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.RadialGradient) {
	if len(gradient.ColorPoints) == 0 {
		return
	}
	r := rectFromPos(layer.BorderBox)
	if r.Dx() <= 0 || r.Dy() <= 0 {
		return
	}
	gradImg := image.NewNRGBA(r)
	cx := float64(gradient.Position.X)
	cy := float64(gradient.Position.Y)
	rx := math.Max(float64(gradient.Radius.X), 0.001)
	ry := math.Max(float64(gradient.Radius.Y), 0.001)
	rad := hasRadius(layer.BorderRadius)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			if rad && !isInsideRoundedRect(px, py, layer.BorderBox, layer.BorderRadius) {
				continue
			}
			ndx := (float64(px) - cx) / rx
			ndy := (float64(py) - cy) / ry
			t := float32(math.Sqrt(ndx*ndx + ndy*ndy))
			gradImg.SetNRGBA(px, py, interpolateGradient(gradient.ColorPoints, t))
		}
	}
	stddraw.Draw(c.img, r, gradImg, r.Min, stddraw.Over)
}

func (c *pngContainer) DrawConicGradient(hdc uintptr, layer litehtml.BackgroundLayer, gradient litehtml.ConicGradient) {
	if len(gradient.ColorPoints) == 0 {
		return
	}
	r := rectFromPos(layer.BorderBox)
	if r.Dx() <= 0 || r.Dy() <= 0 {
		return
	}
	gradImg := image.NewNRGBA(r)
	cx := float64(gradient.Position.X)
	cy := float64(gradient.Position.Y)
	startAngle := float64(gradient.Angle)
	rad := hasRadius(layer.BorderRadius)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			if rad && !isInsideRoundedRect(px, py, layer.BorderBox, layer.BorderRadius) {
				continue
			}
			// atan2(dx, -dy) gives clockwise angle from top (CSS convention)
			a := math.Atan2(float64(px)-cx, -(float64(py) - cy))
			a -= startAngle
			a = math.Mod(a, 2*math.Pi)
			if a < 0 {
				a += 2 * math.Pi
			}
			t := float32(a / (2 * math.Pi))
			gradImg.SetNRGBA(px, py, interpolateGradient(gradient.ColorPoints, t))
		}
	}
	stddraw.Draw(c.img, r, gradImg, r.Min, stddraw.Over)
}

func (c *pngContainer) PtToPx(pt float32) float32   { return pt * 96.0 / 72.0 }
func (c *pngContainer) GetDefaultFontSize() float32 { return 16 }
func (c *pngContainer) GetDefaultFontName() string  { return "serif" }
func (c *pngContainer) LoadImage(src, baseurl string, redrawOnReady bool) {
	if src == "" {
		return
	}

	// Handle data: URIs (e.g. data:image/png;base64,...)
	if strings.HasPrefix(src, "data:") {
		idx := strings.Index(src, ",")
		if idx < 0 {
			return
		}
		encoded := src[idx+1:]
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return
		}
		img, _, err := image.Decode(bytes.NewReader(decoded))
		if err != nil {
			return
		}
		c.images[src] = img
		return
	}

	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		// supported
	default:
		return
	}
	f, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	img, _, err := image.Decode(f)
	if err != nil {
		return
	}
	c.images[src] = img
}
func (c *pngContainer) GetImageSize(src, baseurl string) litehtml.Size {
	img, ok := c.images[src]
	if !ok {
		return litehtml.Size{}
	}
	b := img.Bounds()
	return litehtml.Size{Width: float32(b.Dx()), Height: float32(b.Dy())}
}
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
func (c *pngContainer) ImportCSS(url, baseurl string) (string, string)                   { return "", baseurl }
func (c *pngContainer) SetClip(pos litehtml.Position, bdrRadius litehtml.BorderRadiuses) {}
func (c *pngContainer) DelClip()                                                         {}
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

func hasRadius(r litehtml.BorderRadiuses) bool {
	return r.TopLeftX > 0 || r.TopLeftY > 0 ||
		r.TopRightX > 0 || r.TopRightY > 0 ||
		r.BottomRightX > 0 || r.BottomRightY > 0 ||
		r.BottomLeftX > 0 || r.BottomLeftY > 0
}

func isInsideRoundedRect(px, py int, pos litehtml.Position, r litehtml.BorderRadiuses) bool {
	x := float64(px) - float64(pos.X) + 0.5
	y := float64(py) - float64(pos.Y) + 0.5
	w := float64(pos.Width)
	h := float64(pos.Height)
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	return !isInRoundedCorner(x, y, w, h, r)
}

// isInRoundedCorner checks if a point is outside any of the rounded corners.
func isInRoundedCorner(x, y, w, h float64, r litehtml.BorderRadiuses) bool {
	// Top-left corner
	if isOutsideCornerEllipse(x, y, 0, 0, float64(r.TopLeftX), float64(r.TopLeftY), true, true) {
		return true
	}
	// Top-right corner
	if isOutsideCornerEllipse(x, y, w, 0, float64(r.TopRightX), float64(r.TopRightY), false, true) {
		return true
	}
	// Bottom-right corner
	if isOutsideCornerEllipse(x, y, w, h, float64(r.BottomRightX), float64(r.BottomRightY), false, false) {
		return true
	}
	// Bottom-left corner
	if isOutsideCornerEllipse(x, y, 0, h, float64(r.BottomLeftX), float64(r.BottomLeftY), true, false) {
		return true
	}
	return false
}

// isOutsideCornerEllipse checks if a point is outside a corner's elliptical radius.
func isOutsideCornerEllipse(x, y, cx, cy, rx, ry float64, left, top bool) bool {
	if rx <= 0 || ry <= 0 {
		return false
	}
	var inRegion bool
	if left && top {
		inRegion = x < rx && y < ry
	} else if !left && top {
		inRegion = x > cx-rx && y < ry
	} else if !left && !top {
		inRegion = x > cx-rx && y > cy-ry
	} else {
		inRegion = x < rx && y > cy-ry
	}
	if !inRegion {
		return false
	}

	var dx, dy float64
	if left {
		dx = (rx - x) / rx
	} else {
		dx = (x - (cx - rx)) / rx
	}
	if top {
		dy = (ry - y) / ry
	} else {
		dy = (y - (cy - ry)) / ry
	}
	return dx*dx+dy*dy > 1
}

func interpolateGradient(stops []litehtml.ColorPoint, t float32) color.NRGBA {
	if len(stops) == 0 {
		return color.NRGBA{}
	}
	if t <= stops[0].Offset {
		c := stops[0].Color
		return color.NRGBA{R: c.Red, G: c.Green, B: c.Blue, A: c.Alpha}
	}
	last := stops[len(stops)-1]
	if t >= last.Offset {
		return color.NRGBA{R: last.Color.Red, G: last.Color.Green, B: last.Color.Blue, A: last.Color.Alpha}
	}
	for i := 1; i < len(stops); i++ {
		if t <= stops[i].Offset {
			span := stops[i].Offset - stops[i-1].Offset
			if span <= 0 {
				c := stops[i].Color
				return color.NRGBA{R: c.Red, G: c.Green, B: c.Blue, A: c.Alpha}
			}
			f := (t - stops[i-1].Offset) / span
			c0, c1 := stops[i-1].Color, stops[i].Color
			return color.NRGBA{
				R: uint8(float32(c0.Red) + f*(float32(c1.Red)-float32(c0.Red))),
				G: uint8(float32(c0.Green) + f*(float32(c1.Green)-float32(c0.Green))),
				B: uint8(float32(c0.Blue) + f*(float32(c1.Blue)-float32(c0.Blue))),
				A: uint8(float32(c0.Alpha) + f*(float32(c1.Alpha)-float32(c0.Alpha))),
			}
		}
	}
	return color.NRGBA{R: last.Color.Red, G: last.Color.Green, B: last.Color.Blue, A: last.Color.Alpha}
}

func maxF32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
