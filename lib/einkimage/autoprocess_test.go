package einkimage

import (
	"image"
	"image/color"
	"testing"
)

func TestPaletteProfileBasic(t *testing.T) {
	p, ok := computePaletteProfile(DefaultPalette)
	if !ok {
		t.Fatal("must build profile from non-empty palette")
	}
	if p.ColorCount != 2 {
		t.Errorf("ColorCount=%d, want 2", p.ColorCount)
	}
	if p.LumaRange <= 0 {
		t.Errorf("LumaRange=%v must be > 0 for BW palette", p.LumaRange)
	}
}

func TestPaletteProfileEmpty(t *testing.T) {
	if _, ok := computePaletteProfile(nil); ok {
		t.Error("empty palette must return ok=false")
	}
}

func TestBaseRecommendationTextOrUI(t *testing.T) {
	r := baseRecommendation(KindTextOrUI, "balanced")
	if r.ProcessingPreset != "balanced" {
		t.Errorf("preset=%s", r.ProcessingPreset)
	}
	if r.ColorMatching != MatchLab {
		t.Errorf("textOrUI should use LAB matching")
	}
	if r.DitheringType != QuantizationOnly {
		t.Errorf("textOrUI should use QuantizationOnly")
	}
}

func TestBaseRecommendationPhotoFallback(t *testing.T) {
	r := baseRecommendation(KindPhoto, "soft")
	if r.ErrorDiffusionMatrix != "stucki" {
		t.Errorf("soft photo should pick stucki, got %s", r.ErrorDiffusionMatrix)
	}
}

func TestBaseRecommendationPixelArt(t *testing.T) {
	r := baseRecommendation(KindPixelArt, "balanced")
	if r.ProcessingPreset != "vivid" {
		t.Errorf("pixelArt preset=%s", r.ProcessingPreset)
	}
	if r.DitheringType != QuantizationOnly {
		t.Errorf("pixelArt should use QuantizationOnly")
	}
}

func TestPresetScoresPhotoPreferDynamic(t *testing.T) {
	c := ImageStyleClassification{
		Style: StylePhoto,
		KindScores: map[ImageKind]float64{
			KindPhoto: 0.8, KindLowContrastPhoto: 0.6,
			KindHighContrastPhoto: 0.4, KindFlatIllustration: 0.1,
			KindLineArt: 0, KindTextOrUI: 0, KindPixelArt: 0,
		},
		Metrics: ImageStyleMetrics{LumaStdDev: 40, SaturationMean: 0.3, GrayRatio: 0.2},
	}
	s := presetScores(c, nil, IntentNatural)
	if best := bestPreset(s); best != "dynamic" && best != "balanced" {
		t.Errorf("photo path should favor dynamic or balanced, got %q", best)
	}
}

func TestApplyIntentVivid(t *testing.T) {
	r := recommendation{ProcessingPreset: "balanced"}
	applyIntent(&r, IntentVivid, &[]string{})
	if r.ProcessingPreset != "vivid" {
		t.Errorf("Vivid intent must set preset to vivid, got %s", r.ProcessingPreset)
	}
	if r.ToneMapping == nil || r.ToneMapping.Saturation < 1.45 {
		t.Errorf("Vivid intent should raise saturation to >= 1.45")
	}
}

func TestApplyPaletteTuningMonochrome(t *testing.T) {
	r := recommendation{ProcessingPreset: "balanced"}
	profile := paletteProfile{ColorCount: 2, LumaRange: 100}
	applyPaletteTuning(&r, &profile, &[]string{})
	if r.ProcessingPreset != "grayscale" {
		t.Errorf("Monochrome palette should pick grayscale preset, got %s", r.ProcessingPreset)
	}
	if r.ColorMatching != MatchLab {
		t.Errorf("Monochrome palette should set MatchLab")
	}
}

func TestSuggestProcessingOptionsReturnsReasons(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fillRGBA(img, color.RGBA{200, 50, 50, 255})
	got := SuggestProcessingOptions(img, DefaultPalette, SuggestInput{})
	if len(got.Reasons) == 0 {
		t.Error("reasons must not be empty")
	}
	// BW palette is monochrome so palette tuning must force grayscale.
	if got.DitherOptions.ProcessingPreset != "grayscale" {
		t.Errorf("BW palette should force grayscale preset, got %q",
			got.DitherOptions.ProcessingPreset)
	}
}

func TestSuggestProcessingOptionsNilImage(t *testing.T) {
	got := SuggestProcessingOptions(nil, DefaultPalette, SuggestInput{})
	if got.Classification.Style != StyleUnknown {
		t.Errorf("nil image should classify as unknown, got style=%v", got.Classification.Style)
	}
}
