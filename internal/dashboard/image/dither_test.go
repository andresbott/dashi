package image

import (
	"image"
	"image/color"
	"testing"
)

func TestDitherBWPacked_OutputSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	result := DitherBWPacked(img)

	// ceil(10/8) * 5 = 2 * 5 = 10 bytes
	expected := 10
	if len(result) != expected {
		t.Errorf("Expected %d bytes, got %d", expected, len(result))
	}
}

func TestDitherBWPacked_AllWhite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, white)
		}
	}

	result := DitherBWPacked(img)

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

func TestDitherBWPacked_AllBlack(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, black)
		}
	}

	result := DitherBWPacked(img)

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

func TestDitherBWPacked_WidthNotMultipleOf8(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 11, 1))
	white := color.RGBA{255, 255, 255, 255}
	for x := 0; x < 11; x++ {
		img.Set(x, 0, white)
	}

	result := DitherBWPacked(img)

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

func TestDitherBWRGBA_PixelsAreDeviceColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Fill with gradient from black to white
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			val := uint8((x + y*4) * 16) // 0, 16, 32, ... 240
			img.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}

	result := DitherBWRGBA(img)

	// Every pixel should be either {0x21,0x21,0x21} or {0xe6,0xe6,0xe6}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			p := result.RGBAAt(x, y)
			isDarkGray := p.R == 0x21 && p.G == 0x21 && p.B == 0x21
			isLightGray := p.R == 0xe6 && p.G == 0xe6 && p.B == 0xe6
			if !isDarkGray && !isLightGray {
				t.Errorf("Pixel (%d,%d): expected device color, got RGB(%02X,%02X,%02X)",
					x, y, p.R, p.G, p.B)
			}
		}
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

func TestDitherSpectra6Packed_OutputSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	result := DitherSpectra6Packed(img)

	// (10 * 5) / 2 = 25 bytes
	expected := 25
	if len(result) != expected {
		t.Errorf("Expected %d bytes, got %d", expected, len(result))
	}
}

func TestDitherSpectra6Packed_AllBlack(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, black)
		}
	}

	result := DitherSpectra6Packed(img)

	// All nibbles should be 0 (black)
	expected := []byte{0x00, 0x00, 0x00, 0x00}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestDitherSpectra6Packed_AllWhite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, white)
		}
	}

	result := DitherSpectra6Packed(img)

	// All nibbles should be 1 (white)
	expected := []byte{0x11, 0x11, 0x11, 0x11}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestDitherSpectra6Packed_AllRed(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, red)
		}
	}

	result := DitherSpectra6Packed(img)

	// All nibbles should be 4 (red - wire index)
	expected := []byte{0x44, 0x44, 0x44, 0x44}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestDitherSpectra6Packed_IndicesInRange(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Fill with various colors
	colors := []color.RGBA{
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
		{0, 255, 0, 255},     // Green
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

	result := DitherSpectra6Packed(img)

	// Every nibble should be in range [0, 5]
	for i, b := range result {
		highNibble := (b >> 4) & 0x0F
		lowNibble := b & 0x0F
		if highNibble > 5 {
			t.Errorf("Byte %d high nibble: %d out of range", i, highNibble)
		}
		if lowNibble > 5 {
			t.Errorf("Byte %d low nibble: %d out of range", i, lowNibble)
		}
	}
}

func TestDitherSpectra6RGBA_PixelsAreCalibratedColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Fill with various colors
	colors := []color.RGBA{
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 0, 0, 255},     // Red
		{255, 255, 0, 255},   // Yellow
		{128, 128, 128, 255}, // Gray
		{255, 165, 0, 255},   // Orange
	}
	idx := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, colors[idx%len(colors)])
			idx++
		}
	}

	result := DitherSpectra6RGBA(img)

	// Every pixel should be one of the 6 calibrated colors
	validColors := map[[3]uint8]bool{
		{0x02, 0x02, 0x02}: true, // black
		{0xBE, 0xC8, 0xC8}: true, // white
		{0x27, 0x66, 0x3C}: true, // green
		{0x05, 0x40, 0x9E}: true, // blue
		{0x87, 0x13, 0x00}: true, // red
		{0xCD, 0xCA, 0x00}: true, // yellow
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			p := result.RGBAAt(x, y)
			key := [3]uint8{p.R, p.G, p.B}
			if !validColors[key] {
				t.Errorf("Pixel (%d,%d): RGB(%02X,%02X,%02X) is not a calibrated color",
					x, y, p.R, p.G, p.B)
			}
		}
	}
}
