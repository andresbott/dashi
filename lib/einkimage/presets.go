package einkimage

import "strings"

// ProcessingPreset bundles tone-mapping, DRC, color-matching, and kernel
// defaults under a named preset.
type ProcessingPreset struct {
	Name                 string
	Title                string
	Description          string
	ToneMapping          *ToneMappingOptions
	DRC                  *DRCOptions
	ColorMatching        ColorMatchingMode
	ErrorDiffusionMatrix string
}

var processingPresets = map[string]ProcessingPreset{
	"balanced": {
		Name:  "balanced",
		Title: "Balanced",
		Description: "Compresses display luminance range for general photo conversion.",
		ToneMapping: &ToneMappingOptions{
			Mode: ToneMapContrast, Exposure: 1, Saturation: 1, Contrast: 1,
		},
		DRC:                  &DRCOptions{Mode: DRCDisplay, Strength: 1},
		ColorMatching:        MatchRGB,
		ErrorDiffusionMatrix: "floydSteinberg",
	},
	"dynamic": {
		Name:  "dynamic",
		Title: "Dynamic",
		Description: "Uses S-curve tone mapping for brighter, punchier photographic output.",
		ToneMapping: &ToneMappingOptions{
			Mode: ToneMapSCurve, Exposure: 1, Saturation: 1.3,
			Strength: 0.9, ShadowBoost: 0, HighlightCompress: 1.5, Midpoint: 0.5,
		},
		DRC:                  &DRCOptions{Mode: DRCOff},
		ColorMatching:        MatchRGB,
		ErrorDiffusionMatrix: "floydSteinberg",
	},
	"vivid": {
		Name:  "vivid",
		Title: "Vivid",
		Description: "Boosts color and applies a gentler S-curve for illustrations.",
		ToneMapping: &ToneMappingOptions{
			Mode: ToneMapSCurve, Exposure: 1.1, Saturation: 1.6,
			Strength: 0.7, ShadowBoost: 0.1, HighlightCompress: 1.3, Midpoint: 0.5,
		},
		DRC:                  &DRCOptions{Mode: DRCOff},
		ColorMatching:        MatchRGB,
		ErrorDiffusionMatrix: "floydSteinberg",
	},
	"soft": {
		Name:  "soft",
		Title: "Soft",
		Description: "Reduces contrast and uses Stucki diffusion for smoother tones.",
		ToneMapping: &ToneMappingOptions{
			Mode: ToneMapContrast, Exposure: 1, Saturation: 1.1, Contrast: 0.9,
		},
		DRC:                  &DRCOptions{Mode: DRCDisplay, Strength: 1},
		ColorMatching:        MatchRGB,
		ErrorDiffusionMatrix: "stucki",
	},
	"grayscale": {
		Name:  "grayscale",
		Title: "Grayscale",
		Description: "Removes saturation and uses LAB matching for monochrome work.",
		ToneMapping: &ToneMappingOptions{
			Mode: ToneMapSCurve, Exposure: 1, Saturation: 0,
			Strength: 0.8, ShadowBoost: 0.1, HighlightCompress: 1.4, Midpoint: 0.5,
		},
		DRC:                  &DRCOptions{Mode: DRCDisplay, Strength: 1},
		ColorMatching:        MatchLab,
		ErrorDiffusionMatrix: "floydSteinberg",
	},
}

// GetProcessingPreset looks up a preset by name (case-insensitive). Returns a
// deep copy of the preset so callers can't mutate shared state.
func GetProcessingPreset(name string) (ProcessingPreset, bool) {
	p, ok := processingPresets[strings.ToLower(name)]
	if !ok {
		return ProcessingPreset{}, false
	}
	// Deep copy of inner pointers.
	out := p
	if p.ToneMapping != nil {
		tm := *p.ToneMapping
		out.ToneMapping = &tm
	}
	if p.DRC != nil {
		d := *p.DRC
		out.DRC = &d
	}
	return out, true
}

// GetProcessingPresetNames returns all preset names in a deterministic order.
func GetProcessingPresetNames() []string {
	return []string{"balanced", "dynamic", "vivid", "soft", "grayscale"}
}
