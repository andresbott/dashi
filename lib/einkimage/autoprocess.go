package einkimage

import (
	"image"
	"math"
)

// paletteProfile summarizes a palette for the auto-processing scorer.
type paletteProfile struct {
	ColorCount        int
	LumaRange         float64
	SaturationRange   float64
	AverageSaturation float64
}

func computePaletteProfile(p Palette) (paletteProfile, bool) {
	if len(p) == 0 {
		return paletteProfile{}, false
	}
	lumas := make([]float64, 0, len(p))
	sats := make([]float64, 0, len(p))
	for _, e := range p {
		lumas = append(lumas, luma709(e.Color[0], e.Color[1], e.Color[2]))
		sats = append(sats, channelSaturation(e.Color[0], e.Color[1], e.Color[2]))
	}
	lumaMin, lumaMax := math.Inf(1), math.Inf(-1)
	satMin, satMax, satSum := math.Inf(1), math.Inf(-1), 0.0
	for i := range lumas {
		lumaMin = math.Min(lumaMin, lumas[i])
		lumaMax = math.Max(lumaMax, lumas[i])
		satMin = math.Min(satMin, sats[i])
		satMax = math.Max(satMax, sats[i])
		satSum += sats[i]
	}
	return paletteProfile{
		ColorCount:        len(p),
		LumaRange:         lumaMax - lumaMin,
		SaturationRange:   satMax - satMin,
		AverageSaturation: satSum / float64(len(p)),
	}, true
}

// recommendation holds a partial DitherOptions plus preset name.
type recommendation struct {
	ProcessingPreset     string
	ColorMatching        ColorMatchingMode
	ErrorDiffusionMatrix string
	DitheringType        DitheringType
	ToneMapping          *ToneMappingOptions
	DRC                  *DRCOptions
}

func baseRecommendation(kind ImageKind, fallback string) recommendation {
	switch kind {
	case KindTextOrUI:
		return recommendation{
			ProcessingPreset:     "balanced",
			ColorMatching:        MatchLab,
			ErrorDiffusionMatrix: "floydSteinberg",
			DitheringType:        QuantizationOnly,
			ToneMapping: &ToneMappingOptions{
				Mode: ToneMapContrast, Exposure: 1.05, Saturation: 1, Contrast: 1.18,
			},
			DRC: &DRCOptions{Mode: DRCDisplay, Strength: 0.75},
		}
	case KindLineArt:
		return recommendation{
			ProcessingPreset:     "balanced",
			ColorMatching:        MatchLab,
			ErrorDiffusionMatrix: "floydSteinberg",
			DitheringType:        QuantizationOnly,
			ToneMapping: &ToneMappingOptions{
				Mode: ToneMapContrast, Exposure: 1, Saturation: 0.8, Contrast: 1.25,
			},
			DRC: &DRCOptions{Mode: DRCDisplay, Strength: 0.65},
		}
	case KindPixelArt:
		return recommendation{
			ProcessingPreset:     "vivid",
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: "floydSteinberg",
			DitheringType:        QuantizationOnly,
			ToneMapping:          &ToneMappingOptions{Mode: ToneMapOff, Exposure: 1, Saturation: 1},
			DRC:                  &DRCOptions{Mode: DRCOff},
		}
	case KindFlatIllustration:
		return recommendation{
			ProcessingPreset:     "vivid",
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: "floydSteinberg",
			DitheringType:        ErrorDiffusion,
		}
	case KindLowContrastPhoto:
		return recommendation{
			ProcessingPreset:     "dynamic",
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: "stucki",
			DitheringType:        ErrorDiffusion,
			ToneMapping: &ToneMappingOptions{
				Mode: ToneMapSCurve, Exposure: 1.08, Saturation: 1.25,
				Strength: 0.82, ShadowBoost: 0.06, HighlightCompress: 1.35, Midpoint: 0.48,
			},
			DRC: &DRCOptions{Mode: DRCDisplay, Strength: 0.85},
		}
	case KindHighContrastPhoto:
		return recommendation{
			ProcessingPreset:     "soft",
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: "stucki",
			DitheringType:        ErrorDiffusion,
			DRC:                  &DRCOptions{Mode: DRCDisplay, Strength: 0.9},
		}
	case KindPhoto:
		edm := "floydSteinberg"
		if fallback == "soft" {
			edm = "stucki"
		}
		return recommendation{
			ProcessingPreset:     fallback,
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: edm,
			DitheringType:        ErrorDiffusion,
		}
	default:
		return recommendation{
			ProcessingPreset:     "balanced",
			ColorMatching:        MatchRGB,
			ErrorDiffusionMatrix: "floydSteinberg",
			DitheringType:        ErrorDiffusion,
		}
	}
}

// AutoProcessingIntent adjusts the auto-suggestion toward a user preference.
type AutoProcessingIntent int

const (
	IntentNatural AutoProcessingIntent = iota
	IntentVivid
	IntentReadable
	IntentFaithful
	IntentLowNoise
)

func presetScores(c ImageStyleClassification, profile *paletteProfile, intent AutoProcessingIntent) map[string]float64 {
	s := map[string]float64{
		"balanced":  0.52,
		"dynamic":   0.48,
		"vivid":     0.45,
		"soft":      0.44,
		"grayscale": 0.28,
	}
	switch c.Style {
	case StylePhoto:
		s["dynamic"] += 0.18
		s["balanced"] += 0.12
		if c.Metrics.LumaStdDev >= 68 {
			s["soft"] += 0.2
		} else {
			s["soft"] += 0.06
		}
	case StyleIllustration:
		s["vivid"] += 0.28
		s["balanced"] += 0.08
	}

	ks := c.KindScores
	s["dynamic"] += ks[KindLowContrastPhoto] * 0.24
	s["soft"] += ks[KindHighContrastPhoto] * 0.26
	s["vivid"] += ks[KindFlatIllustration] * 0.24
	s["vivid"] += ks[KindPixelArt] * 0.18
	s["balanced"] += (ks[KindTextOrUI] + ks[KindLineArt]) * 0.18
	grayBoost := 0.08
	if c.Metrics.GrayRatio >= 0.7 {
		grayBoost = 0.24
	}
	s["grayscale"] += (ks[KindTextOrUI] + ks[KindLineArt]) * grayBoost

	if c.Metrics.SaturationMean <= 0.1 && c.Metrics.GrayRatio >= 0.82 {
		s["grayscale"] += 0.22
	}
	if profile != nil && profile.ColorCount <= 2 {
		s["grayscale"] += 0.3
		s["vivid"] -= 0.1
	}

	switch intent {
	case IntentVivid:
		s["vivid"] += 0.18
	case IntentFaithful:
		s["balanced"] += 0.16
	case IntentLowNoise:
		s["soft"] += 0.16
	case IntentReadable:
		s["balanced"] += 0.14
		s["grayscale"] += 0.1
	}
	return s
}

func bestPreset(scores map[string]float64) string {
	best := "balanced"
	bestScore := math.Inf(-1)
	for _, name := range []string{"balanced", "dynamic", "vivid", "soft", "grayscale"} {
		if v, ok := scores[name]; ok && v > bestScore {
			bestScore = v
			best = name
		}
	}
	return best
}

func applyIntent(r *recommendation, intent AutoProcessingIntent, reasons *[]string) {
	switch intent {
	case IntentVivid:
		r.ProcessingPreset = "vivid"
		r.ColorMatching = MatchRGB
		tm := &ToneMappingOptions{Mode: ToneMapSCurve}
		if r.ToneMapping != nil {
			*tm = *r.ToneMapping
			tm.Mode = ToneMapSCurve
		}
		if tm.Saturation < 1.45 {
			tm.Saturation = 1.45
		}
		if tm.Strength == 0 {
			tm.Strength = 0.72
		}
		if tm.ShadowBoost == 0 {
			tm.ShadowBoost = 0.08
		}
		if tm.HighlightCompress == 0 {
			tm.HighlightCompress = 1.3
		}
		if tm.Midpoint == 0 {
			tm.Midpoint = 0.5
		}
		r.ToneMapping = tm
		*reasons = append(*reasons, "Vivid intent boosts saturation and color-priority matching.")
	case IntentReadable:
		r.ColorMatching = MatchLab
		r.DitheringType = QuantizationOnly
		*reasons = append(*reasons, "Readable intent favors clear edges over dithering texture.")
	case IntentLowNoise:
		r.ErrorDiffusionMatrix = "stucki"
		r.ProcessingPreset = "soft"
		*reasons = append(*reasons, "Low-noise intent chooses smoother tone handling.")
	case IntentFaithful:
		r.ProcessingPreset = "balanced"
		*reasons = append(*reasons, "Faithful intent keeps transformations restrained.")
	}
}

func applyPaletteTuning(r *recommendation, profile *paletteProfile, reasons *[]string) {
	if profile == nil {
		return
	}
	if profile.ColorCount <= 2 {
		r.ColorMatching = MatchLab
		r.ProcessingPreset = "grayscale"
		r.ToneMapping = &ToneMappingOptions{
			Mode: ToneMapSCurve, Exposure: 1, Saturation: 0,
			Strength: 0.8, ShadowBoost: 0.1, HighlightCompress: 1.4, Midpoint: 0.5,
		}
		*reasons = append(*reasons, "Monochrome palette switches to grayscale-oriented settings.")
		return
	}
	if profile.LumaRange <= 150 {
		prev := 0.0
		if r.DRC != nil {
			prev = r.DRC.Strength
		}
		if prev < 0.8 {
			prev = 0.8
		}
		r.DRC = &DRCOptions{Mode: DRCDisplay, Strength: prev}
	}
}

// SuggestInput tweaks the auto-suggest heuristic.
type SuggestInput struct {
	Classify ClassifyOptions
	Intent   AutoProcessingIntent
}

// ProcessingSuggestion is the result of SuggestProcessingOptions.
type ProcessingSuggestion struct {
	Classification ImageStyleClassification
	ImageKind      ImageKind
	Intent         AutoProcessingIntent
	DitherOptions  DitherOptions
	Reasons        []string
	Scores         map[string]float64
}

// SuggestProcessingOptions classifies img, scores presets against a palette
// profile + intent, and returns a recommended DitherOptions with reasons.
func SuggestProcessingOptions(img *image.RGBA, palette Palette, in SuggestInput) ProcessingSuggestion {
	classification := ClassifyImageStyle(img, in.Classify)
	profile, _ := computePaletteProfile(palette)
	var profilePtr *paletteProfile
	if len(palette) > 0 {
		profilePtr = &profile
	}

	scores := presetScores(classification, profilePtr, in.Intent)
	best := bestPreset(scores)
	rec := baseRecommendation(classification.Kind, best)

	reasons := []string{}
	addClassificationReasons(&reasons, classification)
	addPaletteReasons(&reasons, profilePtr)
	applyIntent(&rec, in.Intent, &reasons)
	applyPaletteTuning(&rec, profilePtr, &reasons)

	out := DitherOptions{
		ProcessingPreset:        rec.ProcessingPreset,
		ColorMatching:           rec.ColorMatching,
		ErrorDiffusionMatrix:    rec.ErrorDiffusionMatrix,
		DitheringType:           rec.DitheringType,
		ToneMapping:             rec.ToneMapping,
		DynamicRangeCompression: rec.DRC,
	}

	return ProcessingSuggestion{
		Classification: classification,
		ImageKind:      classification.Kind,
		Intent:         in.Intent,
		DitherOptions:  out,
		Reasons:        reasons,
		Scores:         scores,
	}
}

func addClassificationReasons(reasons *[]string, c ImageStyleClassification) {
	kindLabels := map[ImageKind]string{
		KindPhoto: "photo", KindLowContrastPhoto: "lowContrastPhoto",
		KindHighContrastPhoto: "highContrastPhoto",
		KindFlatIllustration: "flatIllustration", KindLineArt: "lineArt",
		KindTextOrUI: "textOrUi", KindPixelArt: "pixelArt",
		KindUnknown: "unknown",
	}
	*reasons = append(*reasons, "Detected "+kindLabels[c.Kind]+".")
	m := c.Metrics
	if m.FlatRatio >= 0.65 {
		*reasons = append(*reasons, "Large flat regions suggest graphic-style preservation.")
	}
	if m.SoftChangeRatio >= 0.38 {
		*reasons = append(*reasons, "Soft tonal transitions suggest photo-oriented processing.")
	}
	if m.LumaStdDev <= 28 {
		*reasons = append(*reasons, "Low luminance spread benefits from stronger tone shaping.")
	}
	if m.LumaStdDev >= 72 {
		*reasons = append(*reasons, "High luminance spread benefits from softer compression.")
	}
	if m.StrongEdgeRatio >= 0.22 {
		*reasons = append(*reasons, "Strong edges favor edge-preserving quantization.")
	}
	if m.TopColorCoverage >= 0.55 {
		*reasons = append(*reasons, "Dominant repeated colors suggest palette-preserving settings.")
	}
	if m.TextTileRatio >= 0.12 {
		*reasons = append(*reasons, "Text-like tiles favor readable edge handling.")
	}
	if m.PhotoTileRatio >= 0.4 {
		*reasons = append(*reasons, "Photo-like tiles favor smoother tonal processing.")
	}
	if m.EdgeDensity >= 0.14 {
		*reasons = append(*reasons, "High edge density affects dithering and matching choice.")
	}
}

func addPaletteReasons(reasons *[]string, profile *paletteProfile) {
	if profile == nil {
		return
	}
	if profile.ColorCount <= 2 {
		*reasons = append(*reasons,
			"Two-color palette favors LAB matching and grayscale-safe output.")
	} else if profile.AverageSaturation >= 0.55 {
		*reasons = append(*reasons, "Colorful target palette can support vivid color mapping.")
	}
	if profile.LumaRange <= 150 {
		*reasons = append(*reasons,
			"Limited palette luminance range benefits from range compression.")
	}
}
