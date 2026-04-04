package dashboard

import (
	"encoding/json"
	"testing"
)

func TestDashboard_ImageConfigSerialization(t *testing.T) {
	d := Dashboard{
		ID:   "test01",
		Name: "Image Test",
		Type: "image",
		Container: Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		ImageConfig: &ImageConfig{
			Width:  1920,
			Height: 1080,
		},
		Rows: []Row{},
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Dashboard
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ImageConfig.Width != 1920 {
		t.Errorf("expected width 1920, got %d", got.ImageConfig.Width)
	}
	if got.ImageConfig.Height != 1080 {
		t.Errorf("expected height 1080, got %d", got.ImageConfig.Height)
	}
}

func TestDashboard_ImageConfigOmittedWhenEmpty(t *testing.T) {
	d := Dashboard{
		ID:   "test02",
		Name: "Static Test",
		Type: "static",
		Container: Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Rows: []Row{},
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if _, exists := raw["imageConfig"]; exists {
		t.Error("imageConfig should be omitted when empty")
	}
}
