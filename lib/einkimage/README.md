# einkimage

Image processing and dithering for color e-paper displays (Spectra 6, ACeP / Gallery, custom palettes). Go port inspired on [epdoptimize](https://github.com/paperlesspaper/epdoptimize).

## The core idea: two-stage color pipeline

E-paper displays have **native colors** (what you send to the panel — e.g. `#FF0000` pure red) and **calibrated colors** (what the panel physically looks like showing that native color — e.g. `#62201E`, a muted dark red for Spectra 6).

Dithering against native colors produces wrong-looking output. This library dithers against each palette entry's **calibrated** `Color` so the dithered image looks right on the display, then maps calibrated colors to **native** `DeviceColor` just before sending pixels to the hardware.

```go
// 1. Dither against calibrated colors.
dithered, err := einkimage.DitherImage(src, einkimage.DitherOptions{
    Palette: einkimage.AitjcizeSpectra6Palette,
})

// 2. Translate calibrated colors to device colors for export.
out, unmatched := einkimage.ReplaceColors(dithered, einkimage.AitjcizeSpectra6Palette)
```

## Quick start

```go
import (
    "image"

    "github.com/andresbott/dashi/lib/einkimage"
)

// Dither against Spectra 6, Floyd-Steinberg (default).
out, err := einkimage.DitherImage(src, einkimage.DitherOptions{
    Palette: einkimage.AitjcizeSpectra6Palette,
})

// With preset + explicit kernel.
out, err = einkimage.DitherImage(src, einkimage.DitherOptions{
    Palette:              einkimage.AitjcizeSpectra6Palette,
    ProcessingPreset:     "dynamic",
    ErrorDiffusionMatrix: "stucki",
    Serpentine:           true,
})

// Let the auto-recommender pick options.
suggestion := einkimage.SuggestProcessingOptions(src, einkimage.AitjcizeSpectra6Palette,
    einkimage.SuggestInput{Intent: einkimage.IntentNatural})
out, err = einkimage.DitherImage(src, suggestion.DitherOptions)
```

## Public API

### Entry points

```go
func DitherImage(src *image.RGBA, opts DitherOptions) (*image.RGBA, error)
```

Applies tone mapping, dynamic range compression, level compression, then the selected dithering algorithm. The returned image's pixels are the palette's calibrated `Color` values — use `ReplaceColors` to translate them to `DeviceColor` before exporting to hardware.

```go
func ReplaceColors(src *image.RGBA, palette Palette) (*image.RGBA, int)
```

Rewrites every pixel whose RGB matches a palette entry's `Color` to that entry's `DeviceColor`. Pixels that don't match any entry are left unchanged; the `int` return is the count of unmatched pixels. Alpha is preserved.

### Classification

```go
func ClassifyImageStyle(img *image.RGBA, opts ClassifyOptions) ImageStyleClassification
func IsPhotoImage(img *image.RGBA, opts ClassifyOptions) bool
func IsIllustrationImage(img *image.RGBA, opts ClassifyOptions) bool
```

Heuristic classifier. Labels an image as `StylePhoto` / `StyleIllustration` / `StyleUnknown` plus a finer `ImageKind` (`KindPhoto`, `KindLowContrastPhoto`, `KindHighContrastPhoto`, `KindFlatIllustration`, `KindLineArt`, `KindTextOrUI`, `KindPixelArt`, `KindUnknown`). Uses color distribution, edge density, saturation spread, and tile-based signals.

### Auto-processing

```go
func SuggestProcessingOptions(img *image.RGBA, palette Palette, in SuggestInput) ProcessingSuggestion
```

Combines the classifier, palette characteristics, and an optional `Intent` (`IntentNatural`, `IntentVivid`, `IntentReadable`, `IntentFaithful`, `IntentLowNoise`) to recommend a full `DitherOptions` with human-readable reasons.

### Presets

```go
func GetProcessingPreset(name string) (ProcessingPreset, bool)
func GetProcessingPresetNames() []string
```

Supported names (case-insensitive): `"balanced"`, `"dynamic"`, `"vivid"`, `"soft"`, `"grayscale"`.

Each preset bundles defaults for `ToneMapping`, `DRC`, `ColorMatching`, and `ErrorDiffusionMatrix`. When `DitherOptions.ProcessingPreset` is set, preset fields fill in anything the caller left at zero; user-set fields always win.

### Palettes

```go
var (
    DefaultPalette           Palette  // 2 colors: black + white
    AitjcizeSpectra6Palette  Palette  // 6 colors: Spectra 6 (recommended calibration)
    AcepPalette              Palette  // 7 colors: ACeP / Gallery
    GameboyPalette           Palette  // 4 colors: gameboy green
)

func GetPaletteByName(name string) (Palette, bool)
func AlignDeviceColors(src, target Palette) Palette
```

`GetPaletteByName` supports `"default"`, `"aitjcize-spectra6"`, `"acep"`, `"gameboy"`.

`AlignDeviceColors` returns `src` with each entry's `DeviceColor` replaced by the one from `target` that shares the same role name (e.g. match `"red"` across calibrations).

### Color conversions

```go
func RGBToLab(r, g, b uint8) [3]float64           // returns [L, a, b]
func LabToRGB(lab [3]float64) (r, g, b uint8)
func DeltaE(lab1, lab2 [3]float64) float64        // CIE76 Euclidean distance
```

## Types

### `DitherOptions`

Zero value produces Floyd-Steinberg error diffusion against `DefaultPalette` with RGB matching — a reasonable default for black-and-white output.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `Palette` | `Palette` | `DefaultPalette` | Target palette. |
| `ProcessingPreset` | `string` | `""` | `"balanced"` / `"dynamic"` / `"vivid"` / `"soft"` / `"grayscale"` / `""`. |
| `DitheringType` | `DitheringType` | `ErrorDiffusion` | `ErrorDiffusion`, `Ordered`, `Random`, `QuantizationOnly`. |
| `ErrorDiffusionMatrix` | `string` | `"floydSteinberg"` | See [kernels](#error-diffusion-kernels). |
| `Serpentine` | `bool` | `false` | Alternate scan direction every row (reduces directional artifacts). |
| `OrderedDitheringMatrix` | `[2]int` | `[4,4]` | Bayer matrix size `[x, y]` for `Ordered` mode. |
| `RandomDitheringType` | `string` | `"blackAndWhite"` | `"blackAndWhite"` or `"rgb"` for `Random` mode. |
| `ColorMatching` | `ColorMatchingMode` | `MatchRGB` | `MatchRGB` or `MatchLab`. |
| `ToneMapping` | `*ToneMappingOptions` | `nil` | Nil disables tone mapping. |
| `DynamicRangeCompression` | `*DRCOptions` | `nil` | Nil disables DRC. |
| `LevelCompression` | `*LevelCompressionOptions` | `nil` | Nil disables level compression. |

### `ToneMappingOptions`

Applied before palette matching. Numeric zero values are treated as the JS reference defaults.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `Mode` | `ToneMappingMode` | `ToneMapContrast` | `ToneMapContrast` / `ToneMapOff` / `ToneMapSCurve`. |
| `Exposure` | `float64` | `1` | Brightness multiplier. |
| `Saturation` | `float64` | `1` | HSL saturation multiplier (0 = grayscale). |
| `Contrast` | `float64` | `1` | Contrast multiplier (contrast mode only). |
| `Strength` | `float64` | `0.9` | S-curve strength (scurve mode only). |
| `ShadowBoost` | `float64` | `0` | Lifts darks (scurve mode only). |
| `HighlightCompress` | `float64` | `1.5` | Compresses highlights (scurve mode only). |
| `Midpoint` | `float64` | `0.5` | S-curve midpoint (scurve mode only). |

### `DRCOptions`

Remaps image luminance into the palette's L* range, preventing shadows / highlights from crushing on limited-range displays.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `Mode` | `DRCMode` | `DRCDisplay` | `DRCDisplay` / `DRCOff` / `DRCAuto` (percentile clipping). |
| `Black`, `White` | `[3]uint8` | derived from palette | Endpoint overrides; pair with `BlackSet` / `WhiteSet = true` to use. |
| `BlackSet`, `WhiteSet` | `bool` | `false` | Set `true` to use explicit `Black` / `White`; otherwise darkest / lightest palette entry is used. |
| `Strength` | `float64` | `1` | Blend factor `[0, 1]`. |
| `LowPercentile` | `float64` | `0.01` | Auto mode only. |
| `HighPercentile` | `float64` | `0.99` | Auto mode only. |

### `LevelCompressionOptions`

Optional per-channel or luma-preserving remap to fit pixels into the display's effective black/white limits.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `Mode` | `LevelCompressionMode` | `LevelPerChannel` | `LevelPerChannel` / `LevelOff` / `LevelLuma`. |
| `Black`, `White` | `[3]uint8` | `{0,0,0}` / `{255,255,255}` | Target black / white points. |
| `BlackSet`, `WhiteSet` | `bool` | `false` | Set `true` to use explicit values. |
| `Auto` | `bool` | `false` | Only run if `AutoThreshold` fraction of pixels are out of range. |
| `AutoThreshold` | `float64` | `0.01` | Fraction (0–1). |

### `PaletteEntry`, `Palette`

```go
type PaletteEntry struct {
    Name        string    // canonical role: "black", "red", "yellow", ...
    Color       [3]uint8  // calibrated RGB (display appearance)
    DeviceColor [3]uint8  // native RGB (what to send to the display)
}

type Palette []PaletteEntry
```

Canonical role ordering: `black`, `white`, `blue`, `green`, `red`, `orange`, `yellow`, `gameboy0`..`gameboy3`. Unknown roles sort after known ones, in input order.

### Classifier types

```go
type ImageStyle int // StyleUnknown | StylePhoto | StyleIllustration

type ImageKind int
// KindUnknown | KindPhoto | KindLowContrastPhoto | KindHighContrastPhoto
// KindFlatIllustration | KindLineArt | KindTextOrUI | KindPixelArt

type ClassifyOptions struct {
    MaxSampleDimension        int     // 0 => 160
    TransparentAlphaThreshold uint8   // 0 => 16
    PhotoThreshold            float64 // 0 => 0.5
}

type ImageStyleClassification struct {
    Style       ImageStyle
    Kind        ImageKind
    KindScores  map[ImageKind]float64
    Confidence  float64
    PhotoScore  float64
    Metrics     ImageStyleMetrics
}
```

`ImageStyleMetrics` exposes every intermediate signal — color distribution, flat/soft/strong-edge ratios, edge density, tile ratios, saturation + luma statistics. See `classify.go` for the full field list.

### Auto-processing types

```go
type AutoProcessingIntent int
// IntentNatural | IntentVivid | IntentReadable | IntentFaithful | IntentLowNoise

type SuggestInput struct {
    Classify ClassifyOptions
    Intent   AutoProcessingIntent
}

type ProcessingSuggestion struct {
    Classification ImageStyleClassification
    ImageKind      ImageKind
    Intent         AutoProcessingIntent
    DitherOptions  DitherOptions
    Reasons        []string             // human-readable "why this choice"
    Scores         map[string]float64   // per-preset score
}
```

## Error diffusion kernels

Valid values for `ErrorDiffusionMatrix`:

| Name | Notes |
| --- | --- |
| `floydSteinberg` | Default. Classic 4-neighbor kernel. |
| `falseFloydSteinberg` | Simplified 3-neighbor variant. Faster, slightly different texture. |
| `atkinson` | Distinctive high-contrast look. Doesn't sum to 1 (loses 25% of error — produces cleaner whites and darker blacks). |
| `jarvis` | Jarvis, Judice, Ninke. Smooth gradients, more blur. |
| `stucki` | Similar to Jarvis with different weights — balances smoothness and sharpness. Common choice for photos. |
| `burkes` | Simplified Stucki. Fewer neighbors, less computation. |
| `sierra3` | Sierra 3. High quality, less blur than Jarvis. |
| `sierra2` | Reduced Sierra 3. Faster. |
| `sierra2-4a` | Lightweight Sierra variant. Fastest. |

Unknown names fall back to `floydSteinberg`.

## Processing presets

| Preset | Tone mapping | DRC | Matching | Kernel | Use for |
| --- | --- | --- | --- | --- | --- |
| `balanced` | contrast | display, strength 1 | RGB | floydSteinberg | General photos. |
| `dynamic` | scurve (punchy) | off | RGB | floydSteinberg | Low-contrast photos that need impact. |
| `vivid` | scurve (saturated) | off | RGB | floydSteinberg | Illustrations, logos, colorful content. |
| `soft` | contrast 0.9 | display, strength 1 | RGB | stucki | High-contrast photos that would otherwise crush. |
| `grayscale` | scurve, saturation 0 | display | LAB | floydSteinberg | Black-and-white displays. |

## How the pipeline works

```
src *image.RGBA
    │
    ▼ applyToneMapping    (exposure → saturation → contrast OR scurve)
    ▼ applyDRC            (LAB lightness compression into palette range)
    ▼ applyLevelCompression (optional perChannel or luma remap)
    ▼ dispatch by DitheringType:
    │     ErrorDiffusion  → scan with selected kernel + optional serpentine
    │     Ordered         → Bayer threshold + nearest palette match
    │     Random          → per-pixel random threshold (BW or RGB)
    │     QuantizationOnly → nearest palette match per pixel
    ▼
out *image.RGBA (calibrated Color values)
    │
    ▼ ReplaceColors       (maps Color → DeviceColor)
    ▼
device-ready *image.RGBA
```

## Caveats and deviations from the JS source

- **Jarvis kernel**: the JS library's `[0,2]` weight is `4/48` (kernel sum ≈ 0.98). The Go port uses `5/48` (sum 1.0), matching the canonical Jarvis–Judice–Ninke weights used in every other reference. Jarvis output will therefore differ slightly from the JS library; all other kernels match exactly.
- **No `spectra6Palette` / `spectra6legacyPalette`**: the JS README marks these "not recommended"; only `aitjcize-spectra6` is ported.
- **No browser / Canvas API**: this is a pure Go library using `*image.RGBA`.
- **Random dither uses a seeded RNG**: output is deterministic across runs (seed fixed at 1). This matches what JS does at a higher level but makes the Go port reproducible.

## License

See the repo root `LICENSE` file.
