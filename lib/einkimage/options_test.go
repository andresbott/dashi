package einkimage

import "testing"

func TestEnumZeroValues(t *testing.T) {
	var dt DitheringType
	if dt != ErrorDiffusion {
		t.Errorf("DitheringType zero=%v, want ErrorDiffusion", dt)
	}
	var cm ColorMatchingMode
	if cm != MatchRGB {
		t.Errorf("ColorMatchingMode zero=%v, want MatchRGB", cm)
	}
	var tm ToneMappingMode
	if tm != ToneMapContrast {
		t.Errorf("ToneMappingMode zero=%v, want ToneMapContrast", tm)
	}
	var dr DRCMode
	if dr != DRCDisplay {
		t.Errorf("DRCMode zero=%v, want DRCDisplay", dr)
	}
	var lc LevelCompressionMode
	if lc != LevelPerChannel {
		t.Errorf("LevelCompressionMode zero=%v, want LevelPerChannel", lc)
	}
}

func TestDitherOptionsZeroValueIsValid(t *testing.T) {
	var o DitherOptions
	if o.DitheringType != ErrorDiffusion {
		t.Errorf("default DitheringType=%v", o.DitheringType)
	}
	if o.ToneMapping != nil {
		t.Error("ToneMapping must default to nil")
	}
	if o.DynamicRangeCompression != nil {
		t.Error("DRC must default to nil")
	}
	if o.LevelCompression != nil {
		t.Error("LevelCompression must default to nil")
	}
}

func TestToneMappingOptionsFields(t *testing.T) {
	tm := ToneMappingOptions{
		Mode: ToneMapSCurve, Exposure: 1.1, Saturation: 1.2,
		Contrast: 1.3, Strength: 0.8, ShadowBoost: 0.1,
		HighlightCompress: 1.5, Midpoint: 0.5,
	}
	if tm.Mode != ToneMapSCurve || tm.Midpoint != 0.5 {
		t.Errorf("field access broken: %+v", tm)
	}
}

func TestDRCOptionsFields(t *testing.T) {
	d := DRCOptions{
		Mode: DRCAuto, Black: [3]uint8{10, 10, 10}, White: [3]uint8{240, 240, 240},
		BlackSet: true, WhiteSet: true, Strength: 0.9,
		LowPercentile: 0.02, HighPercentile: 0.98,
	}
	if !d.BlackSet || d.HighPercentile != 0.98 {
		t.Errorf("DRCOptions field access broken: %+v", d)
	}
}

func TestLevelCompressionOptionsFields(t *testing.T) {
	l := LevelCompressionOptions{
		Mode: LevelLuma, Black: [3]uint8{5, 5, 5}, White: [3]uint8{250, 250, 250},
		BlackSet: true, WhiteSet: true, Auto: true, AutoThreshold: 0.02,
	}
	if !l.Auto || l.AutoThreshold != 0.02 {
		t.Errorf("LevelCompressionOptions field access broken: %+v", l)
	}
}
