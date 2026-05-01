package einkimage

import (
	"image"
	"math"
)

// ImageStyle is the coarse classification.
type ImageStyle int

const (
	StyleUnknown ImageStyle = iota
	StylePhoto
	StyleIllustration
)

// ImageKind is the finer classification.
type ImageKind int

const (
	KindUnknown ImageKind = iota
	KindPhoto
	KindLowContrastPhoto
	KindHighContrastPhoto
	KindFlatIllustration
	KindLineArt
	KindTextOrUI
	KindPixelArt
)

// ImageStyleMetrics holds every heuristic metric computed from the sampled image.
type ImageStyleMetrics struct {
	SampleCount          int
	UniqueColorRatio     float64
	TopColorCoverage     float64
	PaletteEntropy       float64
	FlatRatio            float64
	SoftChangeRatio      float64
	StrongEdgeRatio      float64
	EdgeDensity          float64
	HorizontalEdgeRatio  float64
	VerticalEdgeRatio    float64
	LumaStdDev           float64
	SaturationMean       float64
	SaturationStdDev     float64
	DarkRatio            float64
	LightRatio           float64
	GrayRatio            float64
	HighSaturationRatio  float64
	PhotoTileRatio       float64
	FlatTileRatio        float64
	TextTileRatio        float64
	GradientTileRatio    float64
	TransparentRatio     float64
}

// ImageStyleClassification is the output of ClassifyImageStyle.
type ImageStyleClassification struct {
	Style       ImageStyle
	Kind        ImageKind
	KindScores  map[ImageKind]float64
	Confidence  float64
	PhotoScore  float64
	Metrics     ImageStyleMetrics
}

// ClassifyOptions customizes classifier behavior.
type ClassifyOptions struct {
	MaxSampleDimension        int
	TransparentAlphaThreshold uint8
	PhotoThreshold            float64
}

type classSample struct {
	Visible    bool
	R, G, B    uint8
	Luma       float64
	Saturation float64
}

// ClassifyImageStyle inspects img and returns coarse + fine labels with metrics.
func ClassifyImageStyle(img *image.RGBA, opts ClassifyOptions) ImageStyleClassification {
	if img == nil {
		return ImageStyleClassification{Style: StyleUnknown, Kind: KindUnknown,
			KindScores: emptyKindScores()}
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return ImageStyleClassification{Style: StyleUnknown, Kind: KindUnknown,
			KindScores: emptyKindScores()}
	}

	maxSample := opts.MaxSampleDimension
	if maxSample <= 0 {
		maxSample = 160
	}
	alphaThreshold := opts.TransparentAlphaThreshold
	if alphaThreshold == 0 {
		alphaThreshold = 16
	}
	photoThreshold := opts.PhotoThreshold
	if photoThreshold == 0 {
		photoThreshold = 0.5
	}

	scale := math.Min(1, float64(maxSample)/float64(classMax(w, h)))
	sw := classMax(1, int(math.Round(float64(w)*scale)))
	sh := classMax(1, int(math.Round(float64(h)*scale)))

	samples := make([]classSample, sw*sh)
	colorCounts := make(map[int]int)

	var visibleCount, transparentCount int
	var lumaSum, lumaSq, satSum, satSq float64
	var darkCount, lightCount, grayCount, highSatCount int

	for y := 0; y < sh; y++ {
		srcY := classMin(h-1, int(float64(y)/float64(sh)*float64(h)))
		for x := 0; x < sw; x++ {
			srcX := classMin(w-1, int(float64(x)/float64(sw)*float64(w)))
			p := img.RGBAAt(b.Min.X+srcX, b.Min.Y+srcY)
			if p.A <= alphaThreshold {
				transparentCount++
				samples[y*sw+x] = classSample{}
				continue
			}
			l := luma709(p.R, p.G, p.B)
			s := channelSaturation(p.R, p.G, p.B)
			visibleCount++
			lumaSum += l
			lumaSq += l * l
			satSum += s
			satSq += s * s
			if l <= 36 {
				darkCount++
			}
			if l >= 220 {
				lightCount++
			}
			if s <= 0.08 {
				grayCount++
			}
			if s >= 0.72 {
				highSatCount++
			}
			key := (int(p.R>>3) << 10) | (int(p.G>>3) << 5) | int(p.B>>3)
			colorCounts[key]++
			samples[y*sw+x] = classSample{Visible: true, R: p.R, G: p.G, B: p.B, Luma: l, Saturation: s}
		}
	}

	if visibleCount == 0 {
		return ImageStyleClassification{
			Style: StyleUnknown, Kind: KindUnknown,
			KindScores: emptyKindScores(),
			Metrics: ImageStyleMetrics{
				TransparentRatio: float64(transparentCount) / float64(len(samples)),
			},
		}
	}

	neighbor := computeNeighborMetrics(samples, sw, sh)
	colorDist := computeColorDistributionMetrics(colorCounts, visibleCount)
	edge := computeEdgeMetrics(samples, sw, sh)
	tile := computeTileMetrics(samples, sw, sh)
	lumaMean := lumaSum / float64(visibleCount)
	satMean := satSum / float64(visibleCount)

	m := ImageStyleMetrics{
		SampleCount:         visibleCount,
		UniqueColorRatio:    float64(len(colorCounts)) / float64(visibleCount),
		TopColorCoverage:    colorDist.TopColorCoverage,
		PaletteEntropy:      colorDist.Entropy,
		FlatRatio:           neighbor.FlatRatio,
		SoftChangeRatio:     neighbor.SoftChangeRatio,
		StrongEdgeRatio:     neighbor.StrongEdgeRatio,
		EdgeDensity:         edge.EdgeDensity,
		HorizontalEdgeRatio: edge.HorizontalRatio,
		VerticalEdgeRatio:   edge.VerticalRatio,
		LumaStdDev:          math.Sqrt(math.Max(0, lumaSq/float64(visibleCount)-lumaMean*lumaMean)),
		SaturationMean:      satMean,
		SaturationStdDev:    math.Sqrt(math.Max(0, satSq/float64(visibleCount)-satMean*satMean)),
		DarkRatio:           float64(darkCount) / float64(visibleCount),
		LightRatio:          float64(lightCount) / float64(visibleCount),
		GrayRatio:           float64(grayCount) / float64(visibleCount),
		HighSaturationRatio: float64(highSatCount) / float64(visibleCount),
		PhotoTileRatio:      tile.Photo,
		FlatTileRatio:       tile.Flat,
		TextTileRatio:       tile.Text,
		GradientTileRatio:   tile.Gradient,
		TransparentRatio:    float64(transparentCount) / float64(len(samples)),
	}

	photoScore := computePhotoScore(m)
	confidence := clamp01(math.Abs(photoScore-photoThreshold) * 2)
	style := StyleIllustration
	if photoScore >= photoThreshold {
		style = StylePhoto
	}
	kindScores := computeKindScores(m, photoScore)
	kind := bestKind(kindScores)

	return ImageStyleClassification{
		Style: style, Kind: kind,
		KindScores: kindScores,
		Confidence: confidence, PhotoScore: photoScore,
		Metrics: m,
	}
}

func channelSaturation(r, g, b uint8) float64 {
	nr := float64(r) / 255
	ng := float64(g) / 255
	nb := float64(b) / 255
	maxC := math.Max(nr, math.Max(ng, nb))
	minC := math.Min(nr, math.Min(ng, nb))
	if maxC == 0 {
		return 0
	}
	return (maxC - minC) / maxC
}

func emptyKindScores() map[ImageKind]float64 {
	return map[ImageKind]float64{
		KindPhoto:              0,
		KindLowContrastPhoto:   0,
		KindHighContrastPhoto:  0,
		KindFlatIllustration:   0,
		KindLineArt:            0,
		KindTextOrUI:           0,
		KindPixelArt:           0,
		KindUnknown:            1,
	}
}

func classMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func classMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type neighborMetricsResult struct{ FlatRatio, SoftChangeRatio, StrongEdgeRatio float64 }
type colorDistMetricsResult struct{ TopColorCoverage, Entropy float64 }
type edgeMetricsResult struct{ EdgeDensity, HorizontalRatio, VerticalRatio float64 }
type tileMetricsResult struct{ Photo, Flat, Text, Gradient float64 }

func computeNeighborMetrics(samples []classSample, w, h int) neighborMetricsResult {
	var neighborCount, flat, soft, strong int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := samples[y*w+x]
			if !s.Visible {
				continue
			}
			if x+1 < w {
				n := samples[y*w+x+1]
				if n.Visible {
					neighborCount++
					d := colorDiff(s, n)
					switch {
					case d <= 4:
						flat++
					case d <= 28:
						soft++
					default:
						strong++
					}
				}
			}
			if y+1 < h {
				n := samples[(y+1)*w+x]
				if n.Visible {
					neighborCount++
					d := colorDiff(s, n)
					switch {
					case d <= 4:
						flat++
					case d <= 28:
						soft++
					default:
						strong++
					}
				}
			}
		}
	}
	if neighborCount == 0 {
		return neighborMetricsResult{FlatRatio: 1}
	}
	return neighborMetricsResult{
		FlatRatio:       float64(flat) / float64(neighborCount),
		SoftChangeRatio: float64(soft) / float64(neighborCount),
		StrongEdgeRatio: float64(strong) / float64(neighborCount),
	}
}

func colorDiff(a, b classSample) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func computeColorDistributionMetrics(colorCounts map[int]int, sampleCount int) colorDistMetricsResult {
	if sampleCount == 0 {
		return colorDistMetricsResult{}
	}
	counts := make([]int, 0, len(colorCounts))
	for _, c := range colorCounts {
		counts = append(counts, c)
	}
	// Sort descending.
	for i := 1; i < len(counts); i++ {
		x := counts[i]
		j := i - 1
		for j >= 0 && counts[j] < x {
			counts[j+1] = counts[j]
			j--
		}
		counts[j+1] = x
	}
	top := 0
	lim := 8
	if lim > len(counts) {
		lim = len(counts)
	}
	for i := 0; i < lim; i++ {
		top += counts[i]
	}
	entropy := 0.0
	for _, c := range counts {
		p := float64(c) / float64(sampleCount)
		entropy -= p * math.Log2(p)
	}
	maxEntropy := math.Log2(math.Max(2, float64(len(colorCounts))))
	norm := 0.0
	if maxEntropy > 0 {
		norm = entropy / maxEntropy
	}
	return colorDistMetricsResult{
		TopColorCoverage: float64(top) / float64(sampleCount),
		Entropy:          norm,
	}
}

func computeEdgeMetrics(samples []classSample, w, h int) edgeMetricsResult {
	var checked, edge, horiz, vert int
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			c := samples[y*w+x]
			l := samples[y*w+x-1]
			r := samples[y*w+x+1]
			u := samples[(y-1)*w+x]
			d := samples[(y+1)*w+x]
			if !c.Visible || !l.Visible || !r.Visible || !u.Visible || !d.Visible {
				continue
			}
			checked++
			dx := math.Abs(r.Luma - l.Luma)
			dy := math.Abs(d.Luma - u.Luma)
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag >= 42 {
				edge++
				if dy > dx*1.2 {
					horiz++
				} else if dx > dy*1.2 {
					vert++
				}
			}
		}
	}
	if checked == 0 || edge == 0 {
		return edgeMetricsResult{}
	}
	return edgeMetricsResult{
		EdgeDensity:     float64(edge) / float64(checked),
		HorizontalRatio: float64(horiz) / float64(edge),
		VerticalRatio:   float64(vert) / float64(edge),
	}
}

func computeTileMetrics(samples []classSample, w, h int) tileMetricsResult {
	tileSize := classMax(8, classMin(w, h)/10)
	var tileCount, photo, flat, text, grad int
	for ty := 0; ty < h; ty += tileSize {
		for tx := 0; tx < w; tx += tileSize {
			stats := tileStats(samples, w, h, tx, ty, tileSize)
			if stats == nil {
				continue
			}
			tileCount++
			if stats.EdgeDensity >= 0.16 && stats.GrayRatio >= 0.55 && stats.LumaStdDev >= 38 {
				text++
			}
			if stats.UniqueColorRatio <= 0.12 && stats.FlatRatio >= 0.62 {
				flat++
			}
			if stats.UniqueColorRatio >= 0.18 && stats.LumaStdDev >= 18 && stats.FlatRatio <= 0.68 {
				photo++
			}
			if stats.SoftChangeRatio >= 0.38 && stats.StrongEdgeRatio <= 0.16 && stats.LumaStdDev >= 12 {
				grad++
			}
		}
	}
	if tileCount == 0 {
		return tileMetricsResult{}
	}
	return tileMetricsResult{
		Photo:    float64(photo) / float64(tileCount),
		Flat:     float64(flat) / float64(tileCount),
		Text:     float64(text) / float64(tileCount),
		Gradient: float64(grad) / float64(tileCount),
	}
}

type tileStatsResult struct {
	UniqueColorRatio, GrayRatio, FlatRatio float64
	SoftChangeRatio, StrongEdgeRatio       float64
	EdgeDensity, LumaStdDev                float64
}

func tileStats(samples []classSample, w, h, tx, ty, size int) *tileStatsResult {
	colors := make(map[int]struct{})
	var visible, gray int
	var lumaSum, lumaSq float64
	var neighborCount, flat, soft, strong int
	maxY := classMin(h, ty+size)
	maxX := classMin(w, tx+size)
	for y := ty; y < maxY; y++ {
		for x := tx; x < maxX; x++ {
			s := samples[y*w+x]
			if !s.Visible {
				continue
			}
			visible++
			lumaSum += s.Luma
			lumaSq += s.Luma * s.Luma
			if s.Saturation <= 0.08 {
				gray++
			}
			key := (int(s.R>>3) << 10) | (int(s.G>>3) << 5) | int(s.B>>3)
			colors[key] = struct{}{}
			if x+1 < maxX {
				n := samples[y*w+x+1]
				if n.Visible {
					neighborCount++
					d := colorDiff(s, n)
					switch {
					case d <= 4:
						flat++
					case d <= 28:
						soft++
					default:
						strong++
					}
				}
			}
			if y+1 < maxY {
				n := samples[(y+1)*w+x]
				if n.Visible {
					neighborCount++
					d := colorDiff(s, n)
					switch {
					case d <= 4:
						flat++
					case d <= 28:
						soft++
					default:
						strong++
					}
				}
			}
		}
	}
	thresh := classMax(12, (size*size)/4)
	if visible < thresh {
		return nil
	}
	lumaMean := lumaSum / float64(visible)
	strongEdgeRatio := 0.0
	flatRatio := 1.0
	softRatio := 0.0
	if neighborCount > 0 {
		strongEdgeRatio = float64(strong) / float64(neighborCount)
		flatRatio = float64(flat) / float64(neighborCount)
		softRatio = float64(soft) / float64(neighborCount)
	}
	return &tileStatsResult{
		UniqueColorRatio: float64(len(colors)) / float64(visible),
		GrayRatio:        float64(gray) / float64(visible),
		FlatRatio:        flatRatio,
		SoftChangeRatio:  softRatio,
		StrongEdgeRatio:  strongEdgeRatio,
		EdgeDensity:      strongEdgeRatio,
		LumaStdDev:       math.Sqrt(math.Max(0, lumaSq/float64(visible)-lumaMean*lumaMean)),
	}
}

func computePhotoScore(m ImageStyleMetrics) float64 {
	unique := normalize01(m.UniqueColorRatio, 0.08, 0.35)
	soft := normalize01(m.SoftChangeRatio, 0.18, 0.48)
	texture := normalize01(1-m.FlatRatio, 0.2, 0.65)
	lumaS := normalize01(m.LumaStdDev, 24, 72)
	satSpread := normalize01(m.SaturationStdDev, 0.08, 0.26)
	entropy := normalize01(m.PaletteEntropy, 0.55, 0.9)
	photoTile := normalize01(m.PhotoTileRatio, 0.18, 0.62)
	desatPhoto := math.Min(normalize01(m.GrayRatio, 0.45, 0.75),
		math.Min(normalize01(m.PhotoTileRatio, 0.34, 0.58),
			math.Min(normalize01(m.LumaStdDev, 48, 76),
				math.Min(normalize01(1-m.FlatRatio, 0.34, 0.58),
					normalize01(m.PaletteEntropy, 0.62, 0.88)))))
	flatPen := normalize01(m.FlatRatio, 0.55, 0.88)
	topPen := normalize01(m.TopColorCoverage, 0.45, 0.86)
	edgePen := 0.0
	if m.FlatRatio > 0.35 {
		edgePen = normalize01(m.StrongEdgeRatio, 0.28, 0.58)
	}
	return clamp01(unique*0.34 + soft*0.22 + texture*0.16 + lumaS*0.1 +
		satSpread*0.06 + entropy*0.07 + photoTile*0.13 + desatPhoto*0.24 -
		flatPen*0.12 - topPen*0.1 - edgePen*0.08)
}

func computeKindScores(m ImageStyleMetrics, photoScore float64) map[ImageKind]float64 {
	endpoint := m.DarkRatio + m.LightRatio
	photoTilePen := normalize01(m.PhotoTileRatio, 0.32, 0.58)
	return map[ImageKind]float64{
		KindPhoto: clamp01(photoScore*0.55 +
			normalize01(m.PhotoTileRatio, 0.12, 0.62)*0.25 +
			normalize01(m.PaletteEntropy, 0.55, 0.9)*0.12 +
			normalize01(m.SoftChangeRatio, 0.22, 0.48)*0.08),
		KindLowContrastPhoto: clamp01(photoScore*0.38 +
			normalize01(34-m.LumaStdDev, 0, 22)*0.34 +
			normalize01(m.GradientTileRatio, 0.16, 0.55)*0.18 +
			normalize01(m.SoftChangeRatio, 0.24, 0.5)*0.1),
		KindHighContrastPhoto: clamp01(photoScore*0.42 +
			normalize01(m.LumaStdDev, 58, 92)*0.3 +
			normalize01(endpoint, 0.18, 0.42)*0.18 +
			normalize01(m.PhotoTileRatio, 0.18, 0.58)*0.1),
		KindFlatIllustration: clamp01((1-photoScore)*0.32 +
			normalize01(m.FlatRatio, 0.52, 0.9)*0.2 +
			normalize01(m.TopColorCoverage, 0.38, 0.85)*0.22 +
			normalize01(m.FlatTileRatio, 0.18, 0.72)*0.18 +
			normalize01(m.HighSaturationRatio, 0.08, 0.38)*0.08),
		KindLineArt: clamp01(normalize01(m.GrayRatio, 0.48, 0.9)*0.28 +
			normalize01(m.EdgeDensity, 0.05, 0.22)*0.24 +
			normalize01(m.FlatRatio, 0.5, 0.86)*0.18 +
			normalize01(m.TopColorCoverage, 0.45, 0.9)*0.18 +
			normalize01(0.16-m.HighSaturationRatio, 0, 0.16)*0.12 -
			photoTilePen*0.14),
		KindTextOrUI: clamp01(normalize01(m.TextTileRatio, 0.05, 0.35)*0.32 +
			normalize01(m.EdgeDensity, 0.06, 0.24)*0.22 +
			normalize01(m.GrayRatio, 0.42, 0.86)*0.16 +
			normalize01(m.TopColorCoverage, 0.42, 0.86)*0.16 +
			normalize01(m.FlatTileRatio, 0.12, 0.58)*0.14),
		KindPixelArt: clamp01(normalize01(m.FlatRatio, 0.62, 0.94)*0.25 +
			normalize01(m.TopColorCoverage, 0.5, 0.92)*0.24 +
			normalize01(m.FlatTileRatio, 0.25, 0.82)*0.2 +
			normalize01(m.HighSaturationRatio, 0.08, 0.45)*0.16 +
			normalize01(0.22-m.SoftChangeRatio, 0, 0.22)*0.15),
		KindUnknown: 0,
	}
}

func normalize01(v, lo, hi float64) float64 {
	if hi <= lo {
		if v >= hi {
			return 1
		}
		return 0
	}
	return clamp01((v - lo) / (hi - lo))
}

func bestKind(m map[ImageKind]float64) ImageKind {
	best := KindUnknown
	bestScore := math.Inf(-1)
	for k, v := range m {
		if v > bestScore {
			bestScore = v
			best = k
		}
	}
	return best
}

// IsPhotoImage returns true iff the classifier places img in StylePhoto.
func IsPhotoImage(img *image.RGBA, opts ClassifyOptions) bool {
	return ClassifyImageStyle(img, opts).Style == StylePhoto
}

// IsIllustrationImage returns true iff the classifier places img in
// StyleIllustration.
func IsIllustrationImage(img *image.RGBA, opts ClassifyOptions) bool {
	return ClassifyImageStyle(img, opts).Style == StyleIllustration
}
