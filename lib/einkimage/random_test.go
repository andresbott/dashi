package einkimage

import (
	"math/rand"
	"testing"
)

func TestRandomDitherBWProducesMix(t *testing.T) {
	w, h := 32, 32
	buf := make([]uint8, w*h*4)
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = 128, 128, 128, 255
	}
	rng := rand.New(rand.NewSource(42))
	randomDither(buf, "blackAndWhite", rng)
	var blacks, whites int
	for i := 0; i < len(buf); i += 4 {
		switch buf[i] {
		case 0:
			blacks++
		case 255:
			whites++
		}
	}
	if blacks+whites != w*h {
		t.Errorf("all pixels must be 0 or 255; got b=%d w=%d of %d", blacks, whites, w*h)
	}
	if blacks == 0 || whites == 0 {
		t.Errorf("gray should produce a mix; b=%d w=%d", blacks, whites)
	}
}

func TestRandomDitherRGBProducesMix(t *testing.T) {
	w, h := 32, 32
	buf := make([]uint8, w*h*4)
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = 128, 64, 200, 255
	}
	rng := rand.New(rand.NewSource(42))
	randomDither(buf, "rgb", rng)
	// Every channel must be 0 or 255.
	for i := 0; i < len(buf); i += 4 {
		for c := 0; c < 3; c++ {
			if buf[i+c] != 0 && buf[i+c] != 255 {
				t.Fatalf("pixel %d ch %d = %d, want 0 or 255", i/4, c, buf[i+c])
			}
		}
	}
}
