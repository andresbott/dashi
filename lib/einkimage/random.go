package einkimage

import "math/rand"

// randomDither thresholds each pixel against a random integer in [0, 255].
// Mode is "blackAndWhite" (average RGB → single threshold → black/white) or
// "rgb" (per-channel threshold → 0/255). Other values default to blackAndWhite.
// rng may be nil; a seeded default is used.
func randomDither(buf []uint8, mode string, rng *rand.Rand) {
	if rng == nil {
		rng = rand.New(rand.NewSource(1)) //nolint:gosec // deterministic RNG for dithering, not security-sensitive
	}
	if mode == "rgb" {
		for i := 0; i < len(buf); i += 4 {
			buf[i] = randomThresholdBW(buf[i], rng)
			buf[i+1] = randomThresholdBW(buf[i+1], rng)
			buf[i+2] = randomThresholdBW(buf[i+2], rng)
		}
		return
	}
	// default: blackAndWhite
	for i := 0; i < len(buf); i += 4 {
		avg := (int(buf[i]) + int(buf[i+1]) + int(buf[i+2])) / 3
		if avg < rng.Intn(256) {
			buf[i], buf[i+1], buf[i+2] = 0, 0, 0
		} else {
			buf[i], buf[i+1], buf[i+2] = 255, 255, 255
		}
	}
}

// randomThresholdBW thresholds a single channel value against a random [0, 255] integer.
func randomThresholdBW(v uint8, rng *rand.Rand) uint8 {
	if int(v) < rng.Intn(256) {
		return 0
	}
	return 255
}
