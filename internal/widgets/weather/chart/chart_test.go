package chart

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"
	"time"
)

func TestGenerate_ReturnsPNG(t *testing.T) {
	now := time.Now()
	points := make([]HourlyPoint, 24)
	for i := range points {
		points[i] = HourlyPoint{
			Time:        now.Add(time.Duration(i) * time.Hour),
			Temperature: 10.0 + float64(i)*0.5,
			RainPercent: float64(i * 4),
		}
	}

	opts := Options{
		Width:     800,
		Height:    250,
		TempColor: color.RGBA{R: 0xFF, G: 0x8C, B: 0x42, A: 0xFF},
		RainColor: color.RGBA{R: 0x4A, G: 0x90, B: 0xD9, A: 0xFF},
		BgColor:   color.RGBA{A: 0},
	}

	data, err := Generate(points, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's a valid PNG
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("output is not valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 250 {
		t.Errorf("expected 800x250 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerate_EmptyData(t *testing.T) {
	opts := Options{Width: 400, Height: 150}
	data, err := Generate(nil, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PNG bytes for empty data")
	}
}

func TestGenerate_SinglePoint(t *testing.T) {
	points := []HourlyPoint{
		{Time: time.Now(), Temperature: 20.0, RainPercent: 50.0},
	}
	opts := Options{Width: 400, Height: 150}
	data, err := Generate(points, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PNG bytes for single point")
	}
}
