package market

import (
	"bytes"
	"image/png"
	"testing"
	"time"

	mkt "github.com/andresbott/dashi/internal/market"
)

func TestGenerateChart(t *testing.T) {
	points := []mkt.PricePoint{
		{Time: time.Now().Add(-3 * 24 * time.Hour), Close: 170.0},
		{Time: time.Now().Add(-2 * 24 * time.Hour), Close: 175.0},
		{Time: time.Now().Add(-1 * 24 * time.Hour), Close: 172.0},
		{Time: time.Now(), Close: 178.0},
	}

	data, err := generateChart(points, chartOptions{Width: 400, Height: 150})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 150 {
		t.Errorf("dimensions = %dx%d, want 400x150", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateChart_Empty(t *testing.T) {
	data, err := generateChart(nil, chartOptions{Width: 100, Height: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG even with no points")
	}
}

func TestGenerateChart_SinglePoint(t *testing.T) {
	points := []mkt.PricePoint{
		{Time: time.Now(), Close: 100.0},
	}
	data, err := generateChart(points, chartOptions{Width: 200, Height: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG")
	}
}
