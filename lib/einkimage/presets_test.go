package einkimage

import "testing"

func TestGetProcessingPresetKnown(t *testing.T) {
	for _, name := range []string{"balanced", "dynamic", "vivid", "soft", "grayscale"} {
		p, ok := GetProcessingPreset(name)
		if !ok {
			t.Errorf("%s not found", name)
			continue
		}
		if p.Name != name {
			t.Errorf("preset %s returned Name=%s", name, p.Name)
		}
	}
}

func TestGetProcessingPresetUnknown(t *testing.T) {
	if _, ok := GetProcessingPreset("nope"); ok {
		t.Error("unknown preset returned ok=true")
	}
}

func TestGetProcessingPresetCaseInsensitive(t *testing.T) {
	if _, ok := GetProcessingPreset("BALANCED"); !ok {
		t.Error("preset lookup must be case-insensitive")
	}
}

func TestPresetHasExpectedFields(t *testing.T) {
	p, _ := GetProcessingPreset("dynamic")
	if p.ToneMapping == nil || p.ToneMapping.Mode != ToneMapSCurve {
		t.Errorf("dynamic preset ToneMapping=%+v", p.ToneMapping)
	}
	if p.ErrorDiffusionMatrix == "" {
		t.Error("dynamic preset must set ErrorDiffusionMatrix")
	}
}

func TestGetProcessingPresetNamesReturnsAll(t *testing.T) {
	names := GetProcessingPresetNames()
	if len(names) != 5 {
		t.Errorf("want 5 preset names, got %d: %v", len(names), names)
	}
}
