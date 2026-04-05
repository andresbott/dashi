package chart

import (
	"bytes"
	"image/color"
	"image/png"
	"time"

	"github.com/fogleman/gg"
)

// HourlyPoint is a single data point for the chart.
type HourlyPoint struct {
	Time        time.Time
	Temperature float64
	RainPercent float64 // 0-100
}

// Options controls the chart appearance.
type Options struct {
	Width     int
	Height    int
	TempColor color.Color
	RainColor color.Color
	BgColor   color.Color
}

// Generate renders a chart PNG with temperature curve and rain bars.
// The image is edge-to-edge with no margins or text.
func Generate(points []HourlyPoint, opts Options) ([]byte, error) {
	w := opts.Width
	h := opts.Height
	if w <= 0 {
		w = 800
	}
	if h <= 0 {
		h = 250
	}

	dc := gg.NewContext(w, h)

	// Background
	if opts.BgColor != nil {
		dc.SetColor(opts.BgColor)
		dc.Clear()
	}

	if len(points) == 0 {
		return encodePNG(dc)
	}

	// Compute temperature range
	minTemp, maxTemp := points[0].Temperature, points[0].Temperature
	for _, p := range points[1:] {
		if p.Temperature < minTemp {
			minTemp = p.Temperature
		}
		if p.Temperature > maxTemp {
			maxTemp = p.Temperature
		}
	}
	// Add padding to temp range
	tempRange := maxTemp - minTemp
	if tempRange < 1 {
		tempRange = 1
	}
	padding := tempRange * 0.15
	minTemp -= padding
	maxTemp += padding
	tempRange = maxTemp - minTemp

	fw := float64(w)
	fh := float64(h)
	n := len(points)

	// Helper to map point index to x coordinate
	pointX := func(i int) float64 {
		if n == 1 {
			return fw / 2
		}
		return float64(i) / float64(n-1) * fw
	}

	// Helper to map temperature to y coordinate (top = maxTemp, bottom = minTemp)
	tempY := func(temp float64) float64 {
		return (1 - (temp-minTemp)/tempRange) * fh
	}

	// Helper to map rain percent to bar height
	rainH := func(pct float64) float64 {
		return (pct / 100.0) * fh
	}

	// Draw rain bars
	rainColor := opts.RainColor
	if rainColor == nil {
		rainColor = color.RGBA{R: 0x4A, G: 0x90, B: 0xD9, A: 0xFF}
	}
	r, g, b, _ := rainColor.RGBA()
	barAlpha := color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 80}

	barWidth := fw / float64(n)
	for i, p := range points {
		if p.RainPercent <= 0 {
			continue
		}
		bh := rainH(p.RainPercent)
		x := pointX(i) - barWidth/2
		dc.SetColor(barAlpha)
		dc.DrawRectangle(x, fh-bh, barWidth, bh)
		dc.Fill()
	}

	// Draw subtle horizontal grid lines
	gridColor := color.NRGBA{R: 128, G: 128, B: 128, A: 40}
	dc.SetColor(gridColor)
	dc.SetLineWidth(0.5)
	for i := 1; i < 4; i++ {
		y := fh * float64(i) / 4.0
		dc.DrawLine(0, y, fw, y)
		dc.Stroke()
	}

	// Draw temperature filled area
	tempColor := opts.TempColor
	if tempColor == nil {
		tempColor = color.RGBA{R: 0xFF, G: 0x8C, B: 0x42, A: 0xFF}
	}

	// Build the curve path for fill
	dc.NewSubPath()
	dc.MoveTo(pointX(0), tempY(points[0].Temperature))
	if n > 2 {
		for i := 0; i < n-1; i++ {
			x0 := pointX(i)
			y0 := tempY(points[i].Temperature)
			x1 := pointX(i + 1)
			y1 := tempY(points[i+1].Temperature)
			cx := (x1 - x0) / 3.0
			dc.CubicTo(x0+cx, y0, x1-cx, y1, x1, y1)
		}
	} else if n == 2 {
		dc.LineTo(pointX(1), tempY(points[1].Temperature))
	}

	// Close for fill: line to bottom-right, bottom-left, close
	dc.LineTo(pointX(n-1), fh)
	dc.LineTo(pointX(0), fh)
	dc.ClosePath()

	// Gradient fill: temp color at top fading to transparent at bottom
	tr, tg, tb, _ := tempColor.RGBA()
	r8, g8, b8 := uint8(tr>>8), uint8(tg>>8), uint8(tb>>8)
	grad := gg.NewLinearGradient(0, 0, 0, fh)
	grad.AddColorStop(0, color.NRGBA{R: r8, G: g8, B: b8, A: 120})
	grad.AddColorStop(1, color.NRGBA{R: r8, G: g8, B: b8, A: 0})
	dc.SetFillStyle(grad)
	dc.Fill()

	// Draw the line on top
	dc.SetColor(tempColor)
	dc.SetLineWidth(2.5)
	dc.MoveTo(pointX(0), tempY(points[0].Temperature))
	if n > 2 {
		for i := 0; i < n-1; i++ {
			x0 := pointX(i)
			y0 := tempY(points[i].Temperature)
			x1 := pointX(i + 1)
			y1 := tempY(points[i+1].Temperature)
			cx := (x1 - x0) / 3.0
			dc.CubicTo(x0+cx, y0, x1-cx, y1, x1, y1)
		}
	} else if n == 2 {
		dc.LineTo(pointX(1), tempY(points[1].Temperature))
	}
	dc.Stroke()

	return encodePNG(dc)
}

func encodePNG(dc *gg.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
