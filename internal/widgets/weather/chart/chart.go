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

	if opts.BgColor != nil {
		dc.SetColor(opts.BgColor)
		dc.Clear()
	}

	if len(points) == 0 {
		return encodePNG(dc)
	}

	minTemp, maxTemp := tempRange(points)
	tempRng := maxTemp - minTemp

	fw := float64(w)
	fh := float64(h)
	n := len(points)

	pointX := func(i int) float64 {
		if n == 1 {
			return fw / 2
		}
		return float64(i) / float64(n-1) * fw
	}

	tempY := func(temp float64) float64 {
		return (1 - (temp-minTemp)/tempRng) * fh
	}

	temps := make([]float64, n)
	for i, p := range points {
		temps[i] = p.Temperature
	}

	drawRainBars(dc, points, pointX, fw, fh, opts.RainColor)
	drawGridLines(dc, fw, fh)

	tempColor := opts.TempColor
	if tempColor == nil {
		tempColor = color.RGBA{R: 0xFF, G: 0x8C, B: 0x42, A: 0xFF}
	}
	drawTempFilledCurve(dc, temps, pointX, tempY, fw, fh, tempColor)

	dc.SetColor(tempColor)
	dc.SetLineWidth(2.5)
	drawCurvePath(dc, temps, pointX, tempY)
	dc.Stroke()

	return encodePNG(dc)
}

func tempRange(points []HourlyPoint) (minTemp, maxTemp float64) {
	minTemp, maxTemp = points[0].Temperature, points[0].Temperature
	for _, p := range points[1:] {
		if p.Temperature < minTemp {
			minTemp = p.Temperature
		}
		if p.Temperature > maxTemp {
			maxTemp = p.Temperature
		}
	}
	rng := maxTemp - minTemp
	if rng < 1 {
		rng = 1
	}
	padding := rng * 0.15
	minTemp -= padding
	maxTemp += padding
	return minTemp, maxTemp
}

func drawRainBars(dc *gg.Context, points []HourlyPoint, pointX func(int) float64, fw, fh float64, rainColor color.Color) {
	if rainColor == nil {
		rainColor = color.RGBA{R: 0x4A, G: 0x90, B: 0xD9, A: 0xFF}
	}
	r, g, b, _ := rainColor.RGBA()
	barAlpha := color.NRGBA{R: safeUint8(r >> 8), G: safeUint8(g >> 8), B: safeUint8(b >> 8), A: 80}

	n := len(points)
	barWidth := fw / float64(n)
	for i, p := range points {
		if p.RainPercent <= 0 {
			continue
		}
		bh := (p.RainPercent / 100.0) * fh
		x := pointX(i) - barWidth/2
		dc.SetColor(barAlpha)
		dc.DrawRectangle(x, fh-bh, barWidth, bh)
		dc.Fill()
	}
}

func drawGridLines(dc *gg.Context, fw, fh float64) {
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

func drawTempFilledCurve(dc *gg.Context, temps []float64, pointX func(int) float64, tempY func(float64) float64, fw, fh float64, tempColor color.Color) {
	n := len(temps)
	if n == 0 {
		return
	}
	dc.NewSubPath()
	drawCurvePath(dc, temps, pointX, tempY)
	dc.LineTo(pointX(n-1), fh)
	dc.LineTo(pointX(0), fh)
	dc.ClosePath()

	tr, tg, tb, _ := tempColor.RGBA()
	r8 := safeUint8(tr >> 8)
	g8 := safeUint8(tg >> 8)
	b8 := safeUint8(tb >> 8)
	grad := gg.NewLinearGradient(0, 0, 0, fh)
	grad.AddColorStop(0, color.NRGBA{R: r8, G: g8, B: b8, A: 120})
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
