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

func TestGenerateChart_TwoPoints(t *testing.T) {
	points := []mkt.PricePoint{
		{Time: time.Now().Add(-1 * time.Hour), Close: 100.0},
		{Time: time.Now(), Close: 105.0},
	}
	data, err := generateChart(points, chartOptions{Width: 300, Height: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG")
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 300 || bounds.Dy() != 100 {
		t.Errorf("dimensions = %dx%d, want 300x100", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateChart_DefaultDimensions(t *testing.T) {
	points := []mkt.PricePoint{
		{Time: time.Now().Add(-2 * time.Hour), Close: 100.0},
		{Time: time.Now().Add(-1 * time.Hour), Close: 102.0},
		{Time: time.Now(), Close: 101.0},
	}
	// Test with zero dimensions - should default to 400x150
	data, err := generateChart(points, chartOptions{Width: 0, Height: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 150 {
		t.Errorf("dimensions = %dx%d, want 400x150 (defaults)", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateChart_NegativeDimensions(t *testing.T) {
	points := []mkt.PricePoint{
		{Time: time.Now(), Close: 100.0},
		{Time: time.Now().Add(1 * time.Hour), Close: 105.0},
	}
	// Test with negative dimensions - should default to 400x150
	data, err := generateChart(points, chartOptions{Width: -10, Height: -20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 150 {
		t.Errorf("dimensions = %dx%d, want 400x150 (defaults)", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateChart_VerySmallPriceRange(t *testing.T) {
	// Test with very small price range (< 0.01)
	points := []mkt.PricePoint{
		{Time: time.Now().Add(-3 * time.Hour), Close: 100.000},
		{Time: time.Now().Add(-2 * time.Hour), Close: 100.001},
		{Time: time.Now().Add(-1 * time.Hour), Close: 100.002},
		{Time: time.Now(), Close: 100.003},
	}
	data, err := generateChart(points, chartOptions{Width: 400, Height: 150})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG")
	}
}

func TestGenerateChart_DecreasingPrices(t *testing.T) {
	// Test with decreasing prices to trigger red color
	points := []mkt.PricePoint{
		{Time: time.Now().Add(-3 * time.Hour), Close: 200.0},
		{Time: time.Now().Add(-2 * time.Hour), Close: 190.0},
		{Time: time.Now().Add(-1 * time.Hour), Close: 180.0},
		{Time: time.Now(), Close: 170.0},
	}
	data, err := generateChart(points, chartOptions{Width: 400, Height: 150})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG")
	}
	// Verify it's valid PNG
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
}

func TestSafeUint8(t *testing.T) {
	tests := []struct {
		input    uint32
		expected uint8
	}{
		{0, 0},
		{100, 100},
		{255, 255},
		{256, 255},
		{1000, 255},
		{4294967295, 255},
	}

	for _, tt := range tests {
		result := safeUint8(tt.input)
		if result != tt.expected {
			t.Errorf("safeUint8(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestPickLineColor(t *testing.T) {
	// Test increasing prices (green)
	increasingPoints := []mkt.PricePoint{
		{Time: time.Now().Add(-1 * time.Hour), Close: 100.0},
		{Time: time.Now(), Close: 110.0},
	}
	greenColor := pickLineColor(increasingPoints)
	r, g, b, a := greenColor.RGBA()
	// Green should have high G value
	if g < r || g < b {
		t.Errorf("expected green color for increasing prices, got R=%d G=%d B=%d A=%d", r>>8, g>>8, b>>8, a>>8)
	}

	// Test decreasing prices (red)
	decreasingPoints := []mkt.PricePoint{
		{Time: time.Now().Add(-1 * time.Hour), Close: 110.0},
		{Time: time.Now(), Close: 100.0},
	}
	redColor := pickLineColor(decreasingPoints)
	r, g, b, a = redColor.RGBA()
	// Red should have high R value
	if r < g || r < b {
		t.Errorf("expected red color for decreasing prices, got R=%d G=%d B=%d A=%d", r>>8, g>>8, b>>8, a>>8)
	}
}
