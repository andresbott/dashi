package einkimage

import "testing"

func TestBayerMatrix8x8(t *testing.T) {
	m := bayerMatrix(8, 8)
	if len(m) != 8 || len(m[0]) != 8 {
		t.Fatalf("bayerMatrix 8x8 shape=%dx%d", len(m), len(m[0]))
	}
	// JS reference: top-left should be 0; bottom-left start 42.
	if m[0][0] != 0 {
		t.Errorf("m[0][0]=%d, want 0", m[0][0])
	}
	if m[7][0] != 42 {
		t.Errorf("m[7][0]=%d, want 42", m[7][0])
	}
}

func TestBayerMatrix4x4(t *testing.T) {
	m := bayerMatrix(4, 4)
	if len(m) != 4 || len(m[0]) != 4 {
		t.Fatalf("shape=%dx%d", len(m), len(m[0]))
	}
	// Values must be a permutation of 0..15.
	seen := map[int]bool{}
	for _, row := range m {
		for _, v := range row {
			seen[v] = true
		}
	}
	if len(seen) != 16 {
		t.Errorf("4x4 must contain 16 unique values, got %d", len(seen))
	}
}

func TestOrderedDitherFillsSomePixels(t *testing.T) {
	w, h := 8, 8
	buf := make([]uint8, w*h*4)
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = 128, 128, 128, 255
	}
	orderedDither(buf, w, h, DefaultPalette, [2]int{4, 4}, MatchRGB)
	// Should be a mix of blacks and whites.
	var blacks, whites int
	for i := 0; i < len(buf); i += 4 {
		if buf[i] == 0 {
			blacks++
		}
		if buf[i] == 255 {
			whites++
		}
	}
	if blacks == 0 || whites == 0 {
		t.Errorf("gray should dither to a mix, got b=%d w=%d", blacks, whites)
	}
}
