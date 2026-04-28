package einkimage

import "testing"

func TestApplyExposureIdentity(t *testing.T) {
	buf := []uint8{10, 20, 30, 255, 40, 50, 60, 255}
	applyExposure(buf, 1)
	if buf[0] != 10 || buf[5] != 50 {
		t.Errorf("exposure=1 must not change pixels, got %v", buf)
	}
}

func TestApplyExposureDouble(t *testing.T) {
	buf := []uint8{10, 100, 200, 255}
	applyExposure(buf, 2)
	want := []uint8{20, 200, 255, 255} // 200*2=400 clamps to 255
	for i, w := range want {
		if buf[i] != w {
			t.Errorf("buf[%d]=%d, want %d", i, buf[i], w)
		}
	}
}

func TestApplyExposureDoesNotTouchAlpha(t *testing.T) {
	buf := []uint8{10, 20, 30, 128}
	applyExposure(buf, 5)
	if buf[3] != 128 {
		t.Errorf("alpha changed: %d", buf[3])
	}
}

func TestApplyContrastIdentity(t *testing.T) {
	buf := []uint8{10, 128, 200, 255}
	applyContrast(buf, 1)
	if buf[0] != 10 || buf[1] != 128 || buf[2] != 200 {
		t.Errorf("contrast=1 must not change pixels, got %v", buf)
	}
}

func TestApplyContrastDouble(t *testing.T) {
	// (v - 128) * 2 + 128
	buf := []uint8{0, 128, 255, 255}
	applyContrast(buf, 2)
	// 0 -> (-128)*2+128 = -128 clamp 0; 128 -> 128; 255 -> 382 clamp 255
	want := []uint8{0, 128, 255, 255}
	for i, w := range want {
		if buf[i] != w {
			t.Errorf("buf[%d]=%d, want %d", i, buf[i], w)
		}
	}
}

func TestApplySaturationIdentity(t *testing.T) {
	buf := []uint8{200, 100, 50, 255}
	applySaturation(buf, 1)
	if buf[0] != 200 || buf[1] != 100 || buf[2] != 50 {
		t.Errorf("sat=1 must not change pixels, got %v", buf)
	}
}

func TestApplySaturationZeroProducesGray(t *testing.T) {
	buf := []uint8{200, 100, 50, 255}
	applySaturation(buf, 0)
	// Fully desaturated: R == G == B.
	if buf[0] != buf[1] || buf[1] != buf[2] {
		t.Errorf("fully desaturated expected gray, got %v", buf[:3])
	}
}

func TestApplySaturationSkipsGrayPixels(t *testing.T) {
	// max == min means pixel is already gray; it must pass unchanged.
	buf := []uint8{128, 128, 128, 255}
	applySaturation(buf, 2)
	if buf[0] != 128 || buf[1] != 128 || buf[2] != 128 {
		t.Errorf("gray pixel changed: %v", buf[:3])
	}
}

func TestApplySCurveStrengthZeroIsIdentity(t *testing.T) {
	buf := []uint8{10, 128, 200, 255}
	applySCurve(buf, 0, 0, 1.5, 0.5)
	if buf[0] != 10 || buf[1] != 128 || buf[2] != 200 {
		t.Errorf("strength=0 changed pixels: %v", buf)
	}
}

func TestApplySCurveMidpointIsFixedPoint(t *testing.T) {
	// At the midpoint, input == output regardless of strength.
	buf := []uint8{128, 128, 128, 255}
	applySCurve(buf, 0.9, 0, 1.5, 0.5)
	// 128/255 ≈ 0.502, very close to midpoint 0.5 — diff should be small.
	if diff := int(buf[0]) - 128; diff < -1 || diff > 1 {
		t.Errorf("midpoint shift too large: %d", diff)
	}
}

func TestApplySCurvePushesShadowsAndHighlights(t *testing.T) {
	// With shadowBoost=0 and highlightCompress=1.5, the S-curve pushes
	// highlights down and keeps shadows relatively stable or pushed down.
	dark := []uint8{40, 40, 40, 255}
	light := []uint8{220, 220, 220, 255}
	applySCurve(dark, 0.9, 0, 1.5, 0.5)
	applySCurve(light, 0.9, 0, 1.5, 0.5)
	// Dark should not brighten significantly; light should compress.
	if dark[0] > 40 {
		t.Errorf("dark should not brighten, got %d", dark[0])
	}
	if light[0] >= 220 {
		t.Errorf("light should compress, got %d (expected < 220)", light[0])
	}
}

func TestApplyToneMappingNil(t *testing.T) {
	buf := []uint8{10, 20, 30, 255}
	applyToneMapping(buf, nil) // must not panic, must not change buf
	if buf[0] != 10 {
		t.Errorf("nil opts changed pixels: %v", buf)
	}
}

func TestApplyToneMappingAppliesExposureAndContrast(t *testing.T) {
	// Exposure 2 + identity contrast: every channel should double (clamped).
	buf := []uint8{10, 100, 200, 255}
	applyToneMapping(buf, &ToneMappingOptions{
		Mode: ToneMapContrast, Exposure: 2, Saturation: 1, Contrast: 1,
	})
	want := []uint8{20, 200, 255, 255}
	for i, w := range want {
		if buf[i] != w {
			t.Errorf("buf[%d]=%d, want %d", i, buf[i], w)
		}
	}
}

func TestApplyToneMappingZeroMeansDefault(t *testing.T) {
	// Zero for Exposure/Saturation/Contrast must be treated as 1 (identity).
	buf := []uint8{10, 100, 200, 255}
	applyToneMapping(buf, &ToneMappingOptions{Mode: ToneMapContrast})
	if buf[0] != 10 || buf[1] != 100 || buf[2] != 200 {
		t.Errorf("zero defaults must be identity, got %v", buf[:3])
	}
}

func TestApplyDRCNil(t *testing.T) {
	buf := []uint8{10, 20, 30, 255}
	applyDRC(buf, nil, nil) // no-op
	if buf[0] != 10 {
		t.Errorf("nil opts changed pixels: %v", buf)
	}
}

func TestApplyDRCOff(t *testing.T) {
	buf := []uint8{10, 20, 30, 255}
	applyDRC(buf, &DRCOptions{Mode: DRCOff}, nil)
	if buf[0] != 10 {
		t.Errorf("DRCOff changed pixels: %v", buf)
	}
}

func TestApplyDRCDisplayWithPaletteCompressesRange(t *testing.T) {
	// Palette range [black=30, white=200] — a pure black pixel (0) must
	// come out closer to 30, a pure white pixel (255) closer to 200.
	palette := Palette{
		{Name: "black", Color: [3]uint8{30, 30, 30}},
		{Name: "white", Color: [3]uint8{200, 200, 200}},
	}
	buf := []uint8{
		0, 0, 0, 255,
		255, 255, 255, 255,
	}
	applyDRC(buf, &DRCOptions{Mode: DRCDisplay, Strength: 1}, palette)
	if buf[0] < 20 || buf[0] > 50 {
		t.Errorf("black pixel should land near palette black, got %d", buf[0])
	}
	if buf[4] < 180 || buf[4] > 220 {
		t.Errorf("white pixel should land near palette white, got %d", buf[4])
	}
}

func TestApplyLevelCompressionNil(t *testing.T) {
	buf := []uint8{10, 20, 30, 255}
	applyLevelCompression(buf, nil)
	if buf[0] != 10 {
		t.Errorf("nil opts changed pixels: %v", buf)
	}
}

func TestApplyLevelCompressionOff(t *testing.T) {
	buf := []uint8{10, 20, 30, 255}
	applyLevelCompression(buf, &LevelCompressionOptions{Mode: LevelOff})
	if buf[0] != 10 {
		t.Errorf("LevelOff changed pixels: %v", buf)
	}
}

func TestApplyLevelCompressionPerChannel(t *testing.T) {
	// Range [20, 230] — 0 maps to 20, 255 maps to 230 approximately.
	buf := []uint8{0, 0, 0, 255, 255, 255, 255, 255}
	applyLevelCompression(buf, &LevelCompressionOptions{
		Mode:     LevelPerChannel,
		Black:    [3]uint8{20, 20, 20}, BlackSet: true,
		White:    [3]uint8{230, 230, 230}, WhiteSet: true,
	})
	if buf[0] != 20 {
		t.Errorf("0 mapped to %d, want 20", buf[0])
	}
	if buf[4] != 230 {
		t.Errorf("255 mapped to %d, want 230", buf[4])
	}
}

func TestApplyLevelCompressionLumaScalesProportionally(t *testing.T) {
	// Luma mode: scale each channel by the same ratio based on luma.
	buf := []uint8{100, 0, 0, 255}
	applyLevelCompression(buf, &LevelCompressionOptions{
		Mode:     LevelLuma,
		Black:    [3]uint8{10, 10, 10}, BlackSet: true,
		White:    [3]uint8{240, 240, 240}, WhiteSet: true,
	})
	// Original ratio R:G:B = 100:0:0 must stay 100:0:0 scaled.
	if buf[1] != 0 || buf[2] != 0 {
		t.Errorf("Luma mode should preserve G=B=0, got %v", buf[:3])
	}
}
