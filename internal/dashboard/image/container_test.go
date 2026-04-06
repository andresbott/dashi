package image

import (
	"image"
	"image/color"
	"testing"

	litehtml "github.com/andresbott/litehtml-go"
)

func TestPNGContainer_CreateFont(t *testing.T) {
	c := newPNGContainer(100, 100)

	descr := litehtml.FontDescription{
		Family: "serif",
		Size:   16,
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	id, metrics := c.CreateFont(descr)
	if id == 0 {
		t.Error("expected non-zero font ID")
	}
	if metrics.FontSize != 16 {
		t.Errorf("expected font size 16, got %f", metrics.FontSize)
	}
	if metrics.Height <= 0 {
		t.Error("expected positive height")
	}

	c.DeleteFont(id)
}

func TestPNGContainer_CreateFont_WithCustomFont(t *testing.T) {
	c := newPNGContainer(100, 100)

	// Register a custom font
	c.registerCustomFont("mycustomfont", interRegularTTF)

	descr := litehtml.FontDescription{
		Family: "mycustomfont",
		Size:   20,
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	id, metrics := c.CreateFont(descr)
	if id == 0 {
		t.Error("expected non-zero font ID")
	}
	if metrics.FontSize != 20 {
		t.Errorf("expected font size 20, got %f", metrics.FontSize)
	}

	c.DeleteFont(id)
}

func TestPNGContainer_CreateFont_InvalidData(t *testing.T) {
	c := newPNGContainer(100, 100)

	// Register invalid font data
	c.registerCustomFont("badfont", []byte("not a font"))

	descr := litehtml.FontDescription{
		Family: "badfont",
		Size:   16,
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	// Should fall back to regular font
	id, metrics := c.CreateFont(descr)
	if id == 0 {
		t.Error("expected non-zero font ID even with bad custom font")
	}
	if metrics.FontSize != 16 {
		t.Errorf("expected font size 16, got %f", metrics.FontSize)
	}

	c.DeleteFont(id)
}

func TestPNGContainer_CreateFont_ZeroSize(t *testing.T) {
	c := newPNGContainer(100, 100)

	descr := litehtml.FontDescription{
		Family: "serif",
		Size:   0, // Should default to 16
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	id, metrics := c.CreateFont(descr)
	if id == 0 {
		t.Error("expected non-zero font ID")
	}
	if metrics.Height <= 0 {
		t.Error("expected positive height even with size 0")
	}

	c.DeleteFont(id)
}

func TestPNGContainer_CreateFont_Monospace(t *testing.T) {
	c := newPNGContainer(100, 100)

	tests := []struct {
		name   string
		family string
	}{
		{"monospace", "monospace"},
		{"courier", "courier"},
		{"consolas", "consolas"},
		{"Courier New", "courier new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descr := litehtml.FontDescription{
				Family: tt.family,
				Size:   14,
				Weight: 400,
				Style:  litehtml.FontStyleNormal,
			}

			id, _ := c.CreateFont(descr)
			if id == 0 {
				t.Error("expected non-zero font ID for monospace font")
			}
			c.DeleteFont(id)
		})
	}
}

func TestPNGContainer_CreateFont_StyleVariants(t *testing.T) {
	c := newPNGContainer(100, 100)

	tests := []struct {
		name   string
		weight int
		style  litehtml.FontStyle
	}{
		{"regular", 400, litehtml.FontStyleNormal},
		{"bold", 700, litehtml.FontStyleNormal},
		{"bold_750", 750, litehtml.FontStyleNormal},
		{"italic", 400, litehtml.FontStyleItalic},
		{"bolditalic", 700, litehtml.FontStyleItalic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descr := litehtml.FontDescription{
				Family: "serif",
				Size:   16,
				Weight: tt.weight,
				Style:  tt.style,
			}

			id, _ := c.CreateFont(descr)
			if id == 0 {
				t.Errorf("expected non-zero font ID for %s", tt.name)
			}
			c.DeleteFont(id)
		})
	}
}

func TestPNGContainer_TextWidth(t *testing.T) {
	c := newPNGContainer(100, 100)

	descr := litehtml.FontDescription{
		Family: "serif",
		Size:   16,
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	id, _ := c.CreateFont(descr)
	defer c.DeleteFont(id)

	width := c.TextWidth("Hello", id)
	if width <= 0 {
		t.Error("expected positive text width")
	}

	// Longer text should be wider
	width2 := c.TextWidth("Hello World", id)
	if width2 <= width {
		t.Error("longer text should have greater width")
	}
}

func TestPNGContainer_TextWidth_InvalidFont(t *testing.T) {
	c := newPNGContainer(100, 100)

	// Use an invalid font ID
	width := c.TextWidth("Hello", 9999)
	if width <= 0 {
		t.Error("expected positive fallback text width")
	}
}

func TestPNGContainer_DrawText(t *testing.T) {
	c := newPNGContainer(100, 100)

	descr := litehtml.FontDescription{
		Family: "serif",
		Size:   20,
		Weight: 400,
		Style:  litehtml.FontStyleNormal,
	}

	id, metrics := c.CreateFont(descr)
	defer c.DeleteFont(id)

	clr := litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}
	pos := litehtml.Position{X: 10, Y: 10, Width: 80, Height: metrics.Height}

	c.DrawText(0, "Test", id, clr, pos)

	// Check that some pixels changed from white
	hasColor := false
	for y := 5; y < 35; y++ {
		for x := 5; x < 95; x++ {
			r, g, b, a := c.img.At(x, y).RGBA()
			if a > 0 && (r != 0xFFFF || g != 0xFFFF || b != 0xFFFF) {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("expected text to be drawn on canvas")
	}
}

func TestPNGContainer_DrawText_InvalidFont(t *testing.T) {
	c := newPNGContainer(100, 100)

	clr := litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}
	pos := litehtml.Position{X: 10, Y: 10, Width: 80, Height: 20}

	// Should not panic with invalid font
	c.DrawText(0, "Test", 9999, clr, pos)
}

func TestPNGContainer_DrawSolidFill_NoRadius(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 30},
	}
	clr := litehtml.WebColor{Red: 0, Green: 255, Blue: 0, Alpha: 255}

	c.DrawSolidFill(0, layer, clr)

	// Check that pixels in the fill area are green
	r, g, b, _ := c.img.At(30, 20).RGBA()
	if g < 0x8000 || r > 0x4000 || b > 0x4000 {
		t.Error("expected green fill in the solid fill area")
	}
}

func TestPNGContainer_DrawSolidFill_WithRadius(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
		BorderRadius: litehtml.BorderRadiuses{
			TopLeftX:  10,
			TopLeftY:  10,
			TopRightX: 10,
			TopRightY: 10,
		},
	}
	clr := litehtml.WebColor{Red: 255, Green: 0, Blue: 255, Alpha: 255}

	c.DrawSolidFill(0, layer, clr)

	// Check that pixels in the center are magenta
	r, g, b, _ := c.img.At(30, 25).RGBA()
	if r < 0x8000 || b < 0x8000 || g > 0x4000 {
		t.Error("expected magenta fill in the rounded solid fill area")
	}
}

func TestPNGContainer_DrawBorders_RoundedAllCorners(t *testing.T) {
	c := newPNGContainer(200, 200)

	borders := litehtml.Borders{
		Top:    litehtml.Border{Width: 5, Color: litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}, Style: litehtml.BorderStyleSolid},
		Right:  litehtml.Border{Width: 5, Color: litehtml.WebColor{Red: 0, Green: 255, Blue: 0, Alpha: 255}, Style: litehtml.BorderStyleSolid},
		Bottom: litehtml.Border{Width: 5, Color: litehtml.WebColor{Red: 0, Green: 0, Blue: 255, Alpha: 255}, Style: litehtml.BorderStyleSolid},
		Left:   litehtml.Border{Width: 5, Color: litehtml.WebColor{Red: 255, Green: 255, Blue: 0, Alpha: 255}, Style: litehtml.BorderStyleSolid},
		Radius: litehtml.BorderRadiuses{
			TopLeftX:     15,
			TopLeftY:     15,
			TopRightX:    15,
			TopRightY:    15,
			BottomRightX: 15,
			BottomRightY: 15,
			BottomLeftX:  15,
			BottomLeftY:  15,
		},
	}
	pos := litehtml.Position{X: 20, Y: 20, Width: 100, Height: 80}

	c.DrawBorders(0, borders, pos, false)

	// Check that border pixels exist in various locations
	hasColor := false
	for y := 20; y < 100; y++ {
		for x := 20; x < 120; x++ {
			r, g, b, a := c.img.At(x, y).RGBA()
			if a > 0 && (r > 0x8000 || g > 0x8000 || b > 0x8000) {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("expected colored border pixels in rounded border")
	}
}

func TestPNGContainer_DrawLinearGradient_EmptyStops(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
	}
	gradient := litehtml.LinearGradient{
		ColorPoints: []litehtml.ColorPoint{},
	}
	gradient.Start.X = 10
	gradient.Start.Y = 25
	gradient.End.X = 60
	gradient.End.Y = 25

	// Should not panic with empty color points
	c.DrawLinearGradient(0, layer, gradient)
}

func TestPNGContainer_DrawLinearGradient_ZeroLength(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
	}
	// Create a gradient with a structure that has X and Y fields
	gradient := litehtml.LinearGradient{
		ColorPoints: []litehtml.ColorPoint{
			{Offset: 0, Color: litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}},
			{Offset: 1, Color: litehtml.WebColor{Red: 0, Green: 0, Blue: 255, Alpha: 255}},
		},
	}
	gradient.Start.X = 30
	gradient.Start.Y = 30
	gradient.End.X = 30
	gradient.End.Y = 30

	// Should not panic with zero-length gradient
	c.DrawLinearGradient(0, layer, gradient)
}

func TestPNGContainer_DrawRadialGradient_EmptyStops(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
	}
	gradient := litehtml.RadialGradient{
		ColorPoints: []litehtml.ColorPoint{},
	}
	gradient.Position.X = 35
	gradient.Position.Y = 30
	gradient.Radius.X = 25
	gradient.Radius.Y = 20

	// Should not panic with empty color points
	c.DrawRadialGradient(0, layer, gradient)
}

func TestPNGContainer_DrawConicGradient_EmptyStops(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
	}
	gradient := litehtml.ConicGradient{
		Angle:       0,
		ColorPoints: []litehtml.ColorPoint{},
	}
	gradient.Position.X = 35
	gradient.Position.Y = 30

	// Should not panic with empty color points
	c.DrawConicGradient(0, layer, gradient)
}

func TestPNGContainer_DrawImage_MissingImage(t *testing.T) {
	c := newPNGContainer(100, 100)

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 50, Height: 40},
	}

	// Should not panic with missing image
	c.DrawImage(0, layer, "nonexistent.png", "")
}

func TestPNGContainer_DrawImage_ZeroSize(t *testing.T) {
	c := newPNGContainer(100, 100)

	// Add a dummy image
	c.images["test.png"] = image.NewRGBA(image.Rect(0, 0, 10, 10))

	layer := litehtml.BackgroundLayer{
		BorderBox: litehtml.Position{X: 10, Y: 10, Width: 0, Height: 0}, // Zero size
	}

	// Should not panic with zero-size destination
	c.DrawImage(0, layer, "test.png", "")
}

func TestPNGContainer_GetImageSize(t *testing.T) {
	c := newPNGContainer(100, 100)

	// Add a test image
	testImg := image.NewRGBA(image.Rect(0, 0, 42, 37))
	c.images["test.png"] = testImg

	size := c.GetImageSize("test.png", "")
	if size.Width != 42 {
		t.Errorf("expected width 42, got %f", size.Width)
	}
	if size.Height != 37 {
		t.Errorf("expected height 37, got %f", size.Height)
	}
}

func TestPNGContainer_GetImageSize_Missing(t *testing.T) {
	c := newPNGContainer(100, 100)

	size := c.GetImageSize("missing.png", "")
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected zero size for missing image, got %v", size)
	}
}

func TestPNGContainer_TransformText(t *testing.T) {
	c := newPNGContainer(100, 100)

	tests := []struct {
		name      string
		input     string
		transform litehtml.TextTransform
		expected  string
	}{
		{"uppercase", "hello world", litehtml.TextTransformUppercase, "HELLO WORLD"},
		{"lowercase", "HELLO WORLD", litehtml.TextTransformLowercase, "hello world"},
		{"none", "Hello World", litehtml.TextTransformNone, "Hello World"},
		{"capitalize", "hello", litehtml.TextTransformCapitalize, "hello"}, // Not implemented, should return as-is
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.TransformText(tt.input, tt.transform)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPNGContainer_GetViewport(t *testing.T) {
	c := newPNGContainer(320, 480)

	vp := c.GetViewport()
	if vp.Width != 320 {
		t.Errorf("expected width 320, got %f", vp.Width)
	}
	if vp.Height != 480 {
		t.Errorf("expected height 480, got %f", vp.Height)
	}
}

func TestPNGContainer_GetMediaFeatures(t *testing.T) {
	c := newPNGContainer(640, 480)

	mf := c.GetMediaFeatures()
	if mf.Width != 640 {
		t.Errorf("expected width 640, got %f", mf.Width)
	}
	if mf.Height != 480 {
		t.Errorf("expected height 480, got %f", mf.Height)
	}
	if mf.Type != litehtml.MediaTypeScreen {
		t.Errorf("expected media type screen, got %v", mf.Type)
	}
	if mf.Resolution != 96 {
		t.Errorf("expected resolution 96, got %f", mf.Resolution)
	}
}

func TestPNGContainer_GetLanguage(t *testing.T) {
	c := newPNGContainer(100, 100)

	lang, locale := c.GetLanguage()
	if lang != "en" {
		t.Errorf("expected language 'en', got %q", lang)
	}
	if locale != "en-US" {
		t.Errorf("expected locale 'en-US', got %q", locale)
	}
}

func TestPNGContainer_PtToPx(t *testing.T) {
	c := newPNGContainer(100, 100)

	px := c.PtToPx(72)
	expected := float32(96.0) // 72pt * 96/72 = 96px
	if px != expected {
		t.Errorf("expected %f px, got %f", expected, px)
	}
}

func TestPNGContainer_GetDefaultFontSize(t *testing.T) {
	c := newPNGContainer(100, 100)

	size := c.GetDefaultFontSize()
	if size != 16 {
		t.Errorf("expected default font size 16, got %f", size)
	}
}

func TestPNGContainer_GetDefaultFontName(t *testing.T) {
	c := newPNGContainer(100, 100)

	name := c.GetDefaultFontName()
	if name != "serif" {
		t.Errorf("expected default font name 'serif', got %q", name)
	}
}

func TestPNGContainer_NoOpMethods(t *testing.T) {
	c := newPNGContainer(100, 100)

	// These should not panic
	c.SetCaption("test")
	c.SetBaseURL("http://example.com")
	c.Link("href", "rel", "type")
	c.OnAnchorClick("http://example.com")
	var mouseEvt litehtml.MouseEvent
	c.OnMouseEvent(mouseEvt)
	c.SetCursor("pointer")
	c.ImportCSS("style.css", "")
	c.SetClip(litehtml.Position{}, litehtml.BorderRadiuses{})
	c.DelClip()
	c.CreateElement("div", map[string]string{})
}

func TestHelpers_RectFromPos(t *testing.T) {
	pos := litehtml.Position{X: 10.5, Y: 20.7, Width: 30.3, Height: 40.9}
	r := rectFromPos(pos)

	if r.Min.X != 11 || r.Min.Y != 21 {
		t.Errorf("unexpected rect min: %v", r.Min)
	}
	if r.Max.X != 41 || r.Max.Y != 62 {
		t.Errorf("unexpected rect max: %v", r.Max)
	}
}

func TestHelpers_HasRadius(t *testing.T) {
	tests := []struct {
		name     string
		radius   litehtml.BorderRadiuses
		expected bool
	}{
		{"no radius", litehtml.BorderRadiuses{}, false},
		{"top left", litehtml.BorderRadiuses{TopLeftX: 10}, true},
		{"bottom right", litehtml.BorderRadiuses{BottomRightY: 5}, true},
		{"all zero", litehtml.BorderRadiuses{TopLeftX: 0, TopLeftY: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRadius(tt.radius)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHelpers_IsInsideRoundedRect(t *testing.T) {
	pos := litehtml.Position{X: 10, Y: 10, Width: 80, Height: 60}
	radius := litehtml.BorderRadiuses{
		TopLeftX:  10,
		TopLeftY:  10,
		TopRightX: 10,
		TopRightY: 10,
	}

	tests := []struct {
		name     string
		px, py   int
		expected bool
	}{
		{"center", 50, 40, true},
		{"far outside", 0, 0, false},
		{"right edge inside", 89, 40, true},
		{"bottom edge inside", 50, 69, true},
		{"top-left corner center", 20, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInsideRoundedRect(tt.px, tt.py, pos, radius)
			if result != tt.expected {
				t.Errorf("point (%d,%d): expected %v, got %v", tt.px, tt.py, tt.expected, result)
			}
		})
	}
}

func TestHelpers_InterpolateGradient(t *testing.T) {
	stops := []litehtml.ColorPoint{
		{Offset: 0, Color: litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}},
		{Offset: 1, Color: litehtml.WebColor{Red: 0, Green: 0, Blue: 255, Alpha: 255}},
	}

	tests := []struct {
		name   string
		t      float32
		checkR bool
		checkB bool
	}{
		{"start", 0, true, false},
		{"end", 1, false, true},
		{"middle", 0.5, false, false},
		{"before start", -0.5, true, false},
		{"after end", 1.5, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := interpolateGradient(stops, tt.t)
			if tt.checkR && c.R < 200 {
				t.Errorf("expected red color at t=%f, got %v", tt.t, c)
			}
			if tt.checkB && c.B < 200 {
				t.Errorf("expected blue color at t=%f, got %v", tt.t, c)
			}
		})
	}
}

func TestHelpers_InterpolateGradient_EmptyStops(t *testing.T) {
	c := interpolateGradient([]litehtml.ColorPoint{}, 0.5)
	expected := color.NRGBA{}
	if c != expected {
		t.Errorf("expected zero color for empty stops, got %v", c)
	}
}

func TestHelpers_InterpolateGradient_SingleStop(t *testing.T) {
	stops := []litehtml.ColorPoint{
		{Offset: 0.5, Color: litehtml.WebColor{Red: 128, Green: 64, Blue: 32, Alpha: 255}},
	}

	c := interpolateGradient(stops, 0.7)
	if c.R != 128 || c.G != 64 || c.B != 32 {
		t.Errorf("expected constant color from single stop, got %v", c)
	}
}

func TestHelpers_InterpolateGradient_ZeroSpan(t *testing.T) {
	stops := []litehtml.ColorPoint{
		{Offset: 0.5, Color: litehtml.WebColor{Red: 255, Green: 0, Blue: 0, Alpha: 255}},
		{Offset: 0.5, Color: litehtml.WebColor{Red: 0, Green: 0, Blue: 255, Alpha: 255}},
	}

	c := interpolateGradient(stops, 0.5)
	// When span is zero, it returns the color at that stop (the second one at that offset)
	// The actual implementation returns the second stop's color when span is 0
	if c.R < 200 && c.B < 200 {
		t.Errorf("expected a valid color for zero-span stops at exact offset, got %v", c)
	}
}

func TestHelpers_MaxF32(t *testing.T) {
	tests := []struct {
		a, b, expected float32
	}{
		{5, 10, 10},
		{10, 5, 10},
		{-5, -10, -5},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := maxF32(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("maxF32(%f, %f) = %f, expected %f", tt.a, tt.b, result, tt.expected)
		}
	}
}
