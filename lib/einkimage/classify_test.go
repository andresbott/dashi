package einkimage

import (
	"image"
	"image/color"
	"testing"
)

func TestClassifierEmptyImageReturnsUnknown(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// All transparent.
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	got := ClassifyImageStyle(img, ClassifyOptions{})
	if got.Style != StyleUnknown {
		t.Errorf("style=%v, want StyleUnknown", got.Style)
	}
}

func TestClassifierFlatIllustrationDetectsFlatness(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	fillRGBA(img, color.RGBA{200, 50, 50, 255})
	got := ClassifyImageStyle(img, ClassifyOptions{})
	if got.Metrics.FlatRatio < 0.9 {
		t.Errorf("solid color should have flatRatio >= 0.9, got %v",
			got.Metrics.FlatRatio)
	}
}

func TestClassifierGradientProducesSoftChanges(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			v := uint8(x * 4)
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
	got := ClassifyImageStyle(img, ClassifyOptions{})
	if got.Metrics.SoftChangeRatio < 0.2 {
		t.Errorf("gradient should have softChangeRatio >= 0.2, got %v",
			got.Metrics.SoftChangeRatio)
	}
}

func TestIsPhotoIsIllustrationConsistent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fillRGBA(img, color.RGBA{200, 50, 50, 255})
	cls := ClassifyImageStyle(img, ClassifyOptions{})
	if IsPhotoImage(img, ClassifyOptions{}) != (cls.Style == StylePhoto) {
		t.Error("IsPhoto inconsistent with ClassifyImageStyle")
	}
	if IsIllustrationImage(img, ClassifyOptions{}) != (cls.Style == StyleIllustration) {
		t.Error("IsIllustration inconsistent with ClassifyImageStyle")
	}
}
