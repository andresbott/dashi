package einkimage

// DitheringType selects the dithering algorithm.
type DitheringType int

const (
	// ErrorDiffusion is the default: diffuse quantization error to neighboring
	// pixels using the kernel selected by DitherOptions.ErrorDiffusionMatrix.
	ErrorDiffusion DitheringType = iota
	// Ordered uses a Bayer threshold matrix followed by nearest-palette match.
	Ordered
	// Random thresholds each pixel against a random value.
	Random
	// QuantizationOnly maps each pixel to its nearest palette color (no dither).
	QuantizationOnly
)

// ColorMatchingMode selects the distance model used by palette matching.
type ColorMatchingMode int

const (
	// MatchRGB uses Euclidean distance in sRGB.
	MatchRGB ColorMatchingMode = iota
	// MatchLab uses Euclidean distance in CIE L*a*b* (deltaE).
	MatchLab
)

// ToneMappingMode selects the tone-shape curve.
type ToneMappingMode int

const (
	// ToneMapContrast is the default when ToneMappingOptions is non-nil.
	ToneMapContrast ToneMappingMode = iota
	// ToneMapOff disables contrast/s-curve shaping (exposure/saturation still apply).
	ToneMapOff
	// ToneMapSCurve applies an S-curve with shadow/highlight controls.
	ToneMapSCurve
)

// DRCMode selects the dynamic range compression behavior.
type DRCMode int

const (
	// DRCDisplay compresses into the palette's luminance range. Default.
	DRCDisplay DRCMode = iota
	// DRCOff disables compression entirely.
	DRCOff
	// DRCAuto uses percentile clipping on the source before compression.
	DRCAuto
)

// LevelCompressionMode selects level-compression behavior.
type LevelCompressionMode int

const (
	// LevelPerChannel remaps each RGB channel independently. Default.
	LevelPerChannel LevelCompressionMode = iota
	// LevelOff disables level compression.
	LevelOff
	// LevelLuma remaps by luminance preserving chroma ratio.
	LevelLuma
)

// ToneMappingOptions configures tone mapping. Zero values for numeric fields
// are treated as the documented defaults (JS library semantics).
type ToneMappingOptions struct {
	Mode              ToneMappingMode
	Exposure          float64 // 0 treated as 1
	Saturation        float64 // 0 treated as 1
	Contrast          float64 // 0 treated as 1 (Contrast mode)
	Strength          float64 // SCurve mode; 0 treated as 0.9
	ShadowBoost       float64 // SCurve mode
	HighlightCompress float64 // SCurve mode; 0 treated as 1.5
	Midpoint          float64 // SCurve mode; 0 treated as 0.5
}

// DRCOptions configures dynamic range compression.
type DRCOptions struct {
	Mode           DRCMode
	Black          [3]uint8 // palette-derived when BlackSet == false
	White          [3]uint8 // palette-derived when WhiteSet == false
	BlackSet       bool
	WhiteSet       bool
	Strength       float64 // 0 treated as 1, clamped to [0, 1]
	LowPercentile  float64 // Auto mode; 0 treated as 0.01
	HighPercentile float64 // Auto mode; 0 treated as 0.99
}

// LevelCompressionOptions configures level compression preprocessing.
type LevelCompressionOptions struct {
	Mode          LevelCompressionMode
	Black         [3]uint8
	White         [3]uint8
	BlackSet      bool
	WhiteSet      bool
	Auto          bool
	AutoThreshold float64 // 0 treated as 0.01
}

// DitherOptions is the full set of options accepted by DitherImage.
// The zero value produces Floyd-Steinberg error diffusion against
// DefaultPalette with RGB matching.
type DitherOptions struct {
	Palette Palette

	// ProcessingPreset fills defaults for ToneMapping, DRC, ColorMatching,
	// and ErrorDiffusionMatrix. User-set fields below override preset values.
	// One of: "balanced", "dynamic", "vivid", "soft", "grayscale", or "".
	ProcessingPreset string

	DitheringType DitheringType

	// ErrorDiffusionMatrix selects the kernel. "" defaults to "floydSteinberg".
	ErrorDiffusionMatrix string

	Serpentine bool

	// OrderedDitheringMatrix is the Bayer matrix size [x, y]. Zero = [4, 4].
	OrderedDitheringMatrix [2]int

	// RandomDitheringType: "blackAndWhite" or "rgb". "" = "blackAndWhite".
	RandomDitheringType string

	ColorMatching ColorMatchingMode

	ToneMapping             *ToneMappingOptions
	DynamicRangeCompression *DRCOptions
	LevelCompression        *LevelCompressionOptions
}
