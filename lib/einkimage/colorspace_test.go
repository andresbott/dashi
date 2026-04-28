package einkimage

import (
	"math"
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    [3]uint8
		wantErr bool
	}{
		{"6-digit with hash", "#FF8000", [3]uint8{255, 128, 0}, false},
		{"6-digit without hash", "FF8000", [3]uint8{255, 128, 0}, false},
		{"lowercase", "#ff8000", [3]uint8{255, 128, 0}, false},
		{"3-digit shorthand", "#f80", [3]uint8{255, 136, 0}, false},
		{"3-digit without hash", "f80", [3]uint8{255, 136, 0}, false},
		{"black", "#000000", [3]uint8{0, 0, 0}, false},
		{"white", "#FFFFFF", [3]uint8{255, 255, 255}, false},
		{"invalid length", "#FF80", [3]uint8{}, true},
		{"invalid chars", "#GGHHII", [3]uint8{}, true},
		{"empty", "", [3]uint8{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := hexToRGB(tc.hex)
			if (err != nil) != tc.wantErr {
				t.Fatalf("hexToRGB(%q) err=%v, wantErr=%v", tc.hex, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("hexToRGB(%q)=%v, want %v", tc.hex, got, tc.want)
			}
		})
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		in   float64
		want uint8
	}{
		{-10, 0},
		{0, 0},
		{127.4, 127},
		{127.5, 128},
		{254.6, 255},
		{300, 255},
		{math.NaN(), 0},
		{math.Inf(1), 0},
		{math.Inf(-1), 0},
	}
	for _, tc := range tests {
		if got := clampByte(tc.in); got != tc.want {
			t.Errorf("clampByte(%v)=%d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestLuma709(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    float64
	}{
		{0, 0, 0, 0},
		{255, 255, 255, 255},
		{255, 0, 0, 0.2126 * 255},
		{0, 255, 0, 0.7152 * 255},
		{0, 0, 255, 0.0722 * 255},
	}
	for _, tc := range tests {
		got := luma709(tc.r, tc.g, tc.b)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("luma709(%d,%d,%d)=%v, want %v", tc.r, tc.g, tc.b, got, tc.want)
		}
	}
}

func TestRGBToLabRoundTrip(t *testing.T) {
	// JS reference values (computed by feeding the same RGB through the JS
	// rgbToLab function and recording the result).
	tests := []struct {
		r, g, b uint8
		wantLab [3]float64
	}{
		{0, 0, 0, [3]float64{0, 0, 0}},
		{255, 255, 255, [3]float64{100, 0.00526, -0.01040}},
		{255, 0, 0, [3]float64{53.2408, 80.0925, 67.2032}},
		{0, 255, 0, [3]float64{87.7347, -86.1827, 83.1793}},
		{0, 0, 255, [3]float64{32.2970, 79.1875, -107.8602}},
	}
	for _, tc := range tests {
		lab := RGBToLab(tc.r, tc.g, tc.b)
		for i := 0; i < 3; i++ {
			if math.Abs(lab[i]-tc.wantLab[i]) > 0.015 {
				t.Errorf("RGBToLab(%d,%d,%d)[%d]=%v, want %v",
					tc.r, tc.g, tc.b, i, lab[i], tc.wantLab[i])
			}
		}
	}
}

func TestLabToRGBRoundTrip(t *testing.T) {
	colors := [][3]uint8{
		{0, 0, 0}, {255, 255, 255}, {128, 64, 200}, {200, 100, 50},
		{10, 20, 30}, {240, 240, 240}, {255, 128, 0}, {0, 200, 100},
	}
	for _, c := range colors {
		lab := RGBToLab(c[0], c[1], c[2])
		r, g, b := LabToRGB(lab)
		// Round-trip tolerance: 1 unit each channel (gamma rounding).
		if abs8(r, c[0]) > 1 || abs8(g, c[1]) > 1 || abs8(b, c[2]) > 1 {
			t.Errorf("round-trip %v -> lab %v -> %v", c, lab, [3]uint8{r, g, b})
		}
	}
}

func abs8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestDeltaE(t *testing.T) {
	a := [3]float64{50, 10, 20}
	b := [3]float64{50, 10, 20}
	if d := DeltaE(a, b); d != 0 {
		t.Errorf("DeltaE identical = %v, want 0", d)
	}
	c := [3]float64{53, 14, 24}
	want := math.Sqrt(9 + 16 + 16) // sqrt(41)
	if d := DeltaE(a, c); math.Abs(d-want) > 1e-9 {
		t.Errorf("DeltaE=%v, want %v", d, want)
	}
}
