package market

import (
	"bytes"
	"image/color"
	"image/png"

	"github.com/fogleman/gg"
	mkt "github.com/andresbott/dashi/internal/market"
)

type chartOptions struct {
	Width     int
	Height    int
	LineColor color.Color
}

func generateChart(points []mkt.PricePoint, opts chartOptions) ([]byte, error) {
	w := opts.Width
	h := opts.Height
	if w <= 0 {
		w = 400
	}
	if h <= 0 {
		h = 150
	}

	dc := gg.NewContext(w, h)

	if len(points) == 0 {
		return encodePNG(dc)
	}

	minPrice, maxPrice := points[0].Close, points[0].Close
	for _, p := range points[1:] {
		if p.Close < minPrice {
			minPrice = p.Close
		}
		if p.Close > maxPrice {
			maxPrice = p.Close
		}
	}
	priceRange := maxPrice - minPrice
	if priceRange < 0.01 {
		priceRange = 1
	}
	padding := priceRange * 0.15
	minPrice -= padding
	maxPrice += padding
	priceRange = maxPrice - minPrice

	fw := float64(w)
	fh := float64(h)
	n := len(points)

	pointX := func(i int) float64 {
		if n == 1 {
			return fw / 2
		}
		return float64(i) / float64(n-1) * fw
	}

	priceY := func(price float64) float64 {
		return (1 - (price-minPrice)/priceRange) * fh
	}

	values := make([]float64, n)
	for i, p := range points {
		values[i] = p.Close
	}

	lineColor := opts.LineColor
	if lineColor == nil {
		lineColor = pickLineColor(points)
	}

	drawGrid(dc, fw, fh)
	drawFilledCurve(dc, values, pointX, priceY, fw, fh, lineColor)

	dc.SetColor(lineColor)
	dc.SetLineWidth(2)
	drawCurvePath(dc, values, pointX, priceY)
	dc.Stroke()

	return encodePNG(dc)
}

func pickLineColor(points []mkt.PricePoint) color.Color {
	if points[len(points)-1].Close >= points[0].Close {
		return color.RGBA{R: 0x22, G: 0xC5, B: 0x5E, A: 0xFF}
	}
	return color.RGBA{R: 0xEF, G: 0x44, B: 0x44, A: 0xFF}
}

func drawGrid(dc *gg.Context, fw, fh float64) {
	gridColor := color.NRGBA{R: 128, G: 128, B: 128, A: 40}
	dc.SetColor(gridColor)
	dc.SetLineWidth(0.5)
	for i := 1; i < 4; i++ {
		y := fh * float64(i) / 4.0
		dc.DrawLine(0, y, fw, y)
		dc.Stroke()
	}
}

// drawCurvePath traces a smooth curve through the given values without stroking or filling.
func drawCurvePath(dc *gg.Context, values []float64, pointX func(int) float64, valY func(float64) float64) {
	n := len(values)
	if n == 0 {
		return
	}
	dc.MoveTo(pointX(0), valY(values[0]))
	if n > 2 {
		for i := 0; i < n-1; i++ {
			x0 := pointX(i)
			y0 := valY(values[i])
			x1 := pointX(i + 1)
			y1 := valY(values[i+1])
			cx := (x1 - x0) / 3.0
			dc.CubicTo(x0+cx, y0, x1-cx, y1, x1, y1)
		}
	} else if n == 2 {
		dc.LineTo(pointX(1), valY(values[1]))
	}
}

// drawFilledCurve draws a filled area under the curve with a gradient.
func drawFilledCurve(dc *gg.Context, values []float64, pointX func(int) float64, valY func(float64) float64, fw, fh float64, lineColor color.Color) {
	n := len(values)
	if n == 0 {
		return
	}
	dc.NewSubPath()
	drawCurvePath(dc, values, pointX, valY)
	dc.LineTo(pointX(n-1), fh)
	dc.LineTo(pointX(0), fh)
	dc.ClosePath()

	r, g, b, _ := lineColor.RGBA()
	r8 := safeUint8(r >> 8)
	g8 := safeUint8(g >> 8)
	b8 := safeUint8(b >> 8)
	grad := gg.NewLinearGradient(0, 0, 0, fh)
	grad.AddColorStop(0, color.NRGBA{R: r8, G: g8, B: b8, A: 80})
	grad.AddColorStop(1, color.NRGBA{R: r8, G: g8, B: b8, A: 0})
	dc.SetFillStyle(grad)
	dc.Fill()
}

// safeUint8 converts a uint32 value to uint8 with overflow protection.
func safeUint8(val uint32) uint8 {
	if val > 255 {
		return 255
	}
	return uint8(val)
}

func encodePNG(dc *gg.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
