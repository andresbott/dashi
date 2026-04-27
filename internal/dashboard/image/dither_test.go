package image

import (
	"image"
	"image/color"
	"testing"
)

func TestDitherBW_OutputSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	result := DitherBW(img)

	// ceil(10/8) * 5 = 2 * 5 = 10 bytes
	expected := 10
	if len(result) != expected {
		t.Errorf("Expected %d bytes, got %d", expected, len(result))
	}
}

func TestDitherBW_AllWhite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, white)
		}
	}

	result := DitherBW(img)

	// All pixels should be white (bit = 1)
	expected := []byte{0xFF, 0xFF}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestDitherBW_AllBlack(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, black)
		}
	}

	result := DitherBW(img)

	// All pixels should be black (bit = 0)
	expected := []byte{0x00, 0x00}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestDitherBW_WidthNotMultipleOf8(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 11, 1))
	white := color.RGBA{255, 255, 255, 255}
	for x := 0; x < 11; x++ {
		img.Set(x, 0, white)
	}

	result := DitherBW(img)

	// ceil(11/8) = 2 bytes
	// First byte: 8 white pixels = 0xFF
	// Second byte: 3 white pixels in MSB positions = 0xE0 (11100000)
	if len(result) != 2 {
		t.Fatalf("Expected 2 bytes, got %d", len(result))
	}
	if result[0] != 0xFF {
		t.Errorf("Byte 0: expected 0xFF, got 0x%02X", result[0])
	}
	if result[1] != 0xE0 {
		t.Errorf("Byte 1: expected 0xE0, got 0x%02X", result[1])
	}
}

func TestPackBW_KnownPattern(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 8, 1))
	// Alternating white/black: W B W B W B W B
	for x := 0; x < 8; x++ {
		if x%2 == 0 {
			img.SetGray(x, 0, color.Gray{Y: 255}) // White
		} else {
			img.SetGray(x, 0, color.Gray{Y: 0}) // Black
		}
	}

	result := PackBW(img, 8, 1)

	// Pattern: 1 0 1 0 1 0 1 0 = 0xAA
	expected := byte(0xAA)
	if len(result) != 1 {
		t.Fatalf("Expected 1 byte, got %d", len(result))
	}
	if result[0] != expected {
		t.Errorf("Expected 0x%02X, got 0x%02X", expected, result[0])
	}
}

func TestDitherSpectra6_OutputSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	result := DitherSpectra6(img)

	// Should be 5 rows of 10 columns
	if len(result) != 5 {
		t.Fatalf("Expected 5 rows, got %d", len(result))
	}
	for y, row := range result {
		if len(row) != 10 {
			t.Errorf("Row %d: expected 10 columns, got %d", y, len(row))
		}
	}
}

func TestDitherSpectra6_AllWhite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, white)
		}
	}

	result := DitherSpectra6(img)

	// All pixels should map to palette index 1 (white)
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			if result[y][x] != 1 {
				t.Errorf("Pixel (%d,%d): expected index 1, got %d", x, y, result[y][x])
			}
		}
	}
}

func TestDitherSpectra6_AllBlack(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, black)
		}
	}

	result := DitherSpectra6(img)

	// All pixels should map to palette index 0 (black)
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			if result[y][x] != 0 {
				t.Errorf("Pixel (%d,%d): expected index 0, got %d", x, y, result[y][x])
			}
		}
	}
}

func TestDitherSpectra6_IndicesInRange(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Fill with various colors
	colors := []color.RGBA{
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
		{0, 128, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 0, 0, 255},     // Red
		{255, 255, 0, 255},   // Yellow
		{128, 128, 128, 255}, // Gray
		{255, 165, 0, 255},   // Orange (should dither to red/yellow)
	}
	idx := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, colors[idx%len(colors)])
			idx++
		}
	}

	result := DitherSpectra6(img)

	// All indices should be in range [0, 5]
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if result[y][x] > 5 {
				t.Errorf("Pixel (%d,%d): index %d out of range", x, y, result[y][x])
			}
		}
	}
}

func TestPackSpectra6_KnownPattern(t *testing.T) {
	// Create indices [0, 1, 2, 3] in a 4x1 image
	indices := [][]uint8{
		{0, 1, 2, 3},
	}

	result := PackSpectra6(indices, 4, 1)

	// High nibble = left pixel, low nibble = right pixel
	// Pair 0,1 -> 0x01
	// Pair 2,3 -> 0x23
	expected := []byte{0x01, 0x23}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestPackSpectra6_OutputSize(t *testing.T) {
	// Create 10x5 image
	indices := make([][]uint8, 5)
	for y := 0; y < 5; y++ {
		indices[y] = make([]uint8, 10)
	}

	result := PackSpectra6(indices, 10, 5)

	// (10 * 5) / 2 = 25 bytes
	expected := 25
	if len(result) != expected {
		t.Errorf("Expected %d bytes, got %d", expected, len(result))
	}
}
