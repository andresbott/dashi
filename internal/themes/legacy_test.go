package themes

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeThemeMigratesLegacyBackground(t *testing.T) {
	tcs := []struct {
		name string
		raw  string
		want Background
	}{
		{
			name: "light image only is copied into dark",
			raw: `{"background":{"image":{"light":"asset:a.png","fit":"cover","position":"center","repeat":"no-repeat"}}}`,
			want: Background{
				Light: ModeBackground{Image: &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}},
				Dark:  ModeBackground{Image: &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}},
			},
		},
		{
			name: "dark image overrides the copied ref",
			raw: `{"background":{"image":{"light":"asset:a.png","dark":"asset:b.png","fit":"contain","position":"top","repeat":"repeat"}}}`,
			want: Background{
				Light: ModeBackground{Image: &Image{Ref: "asset:a.png", Fit: "contain", Position: "top", Repeat: "repeat"}},
				Dark:  ModeBackground{Image: &Image{Ref: "asset:b.png", Fit: "contain", Position: "top", Repeat: "repeat"}},
			},
		},
		{
			name: "light gradient only is copied into dark",
			raw:  `{"background":{"gradient":{"direction":"to bottom","light":["#fff","#eee"]}}}`,
			want: Background{
				Light: ModeBackground{Gradient: &Gradient{Direction: "to bottom", Stops: []string{"#fff", "#eee"}}},
				Dark:  ModeBackground{Gradient: &Gradient{Direction: "to bottom", Stops: []string{"#fff", "#eee"}}},
			},
		},
		{
			name: "dark stops override the copied stops and the image combines",
			raw: `{"background":{"gradient":{"direction":"45deg","light":["#fff","#eee"],"dark":["#000","#111"]},` +
				`"image":{"light":"asset:a.png","fit":"cover","position":"center","repeat":"no-repeat"}}}`,
			want: Background{
				Light: ModeBackground{
					Gradient: &Gradient{Direction: "45deg", Stops: []string{"#fff", "#eee"}},
					Image:    &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
				},
				Dark: ModeBackground{
					Gradient: &Gradient{Direction: "45deg", Stops: []string{"#000", "#111"}},
					Image:    &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"},
				},
			},
		},
		{
			name: "new shape is left untouched",
			raw:  `{"background":{"light":{"image":{"ref":"asset:a.png","fit":"cover","position":"center","repeat":"no-repeat"}},"dark":{}}}`,
			want: Background{
				Light: ModeBackground{Image: &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}},
			},
		},
		{
			name: "plain background stays plain",
			raw:  `{"background":{}}`,
			want: Background{},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeTheme([]byte(tc.raw))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(got.Background, tc.want) {
				t.Errorf("background:\n got %s\nwant %s", fmtBackground(got.Background), fmtBackground(tc.want))
			}
		})
	}
}

func TestDecodeThemeLegacyDarkCopyIsNotAliased(t *testing.T) {
	raw := `{"background":{"gradient":{"direction":"to bottom","light":["#fff","#eee"]},` +
		`"image":{"light":"asset:a.png","fit":"cover","position":"center","repeat":"no-repeat"}}}`
	got, err := decodeTheme([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	bg := got.Background
	if bg.Light.Gradient == bg.Dark.Gradient || bg.Light.Image == bg.Dark.Image {
		t.Fatal("light and dark must not share pointers")
	}
	bg.Dark.Gradient.Stops[0] = "#000"
	if bg.Light.Gradient.Stops[0] != "#fff" {
		t.Error("dark stops must not alias the light stops")
	}
}

func TestDecodeRequestBackgroundShape(t *testing.T) {
	tcs := []struct {
		name    string
		raw     string
		wantErr bool
		want    Background
	}{
		{
			name:    "legacy gradient is rejected",
			raw:     `{"darkMode":true,"background":{"gradient":{"direction":"to bottom","light":["#fff","#eee"]}}}`,
			wantErr: true,
		},
		{
			name:    "legacy image is rejected",
			raw:     `{"darkMode":true,"background":{"image":{"light":"asset:a.png","fit":"cover"}}}`,
			wantErr: true,
		},
		{
			name:    "legacy keys next to the new shape are rejected",
			raw:     `{"darkMode":true,"background":{"light":{},"dark":{},"image":{"light":"asset:a.png"}}}`,
			wantErr: true,
		},
		{
			name: "new shape decodes",
			raw:  `{"darkMode":true,"background":{"light":{"image":{"ref":"asset:a.png","fit":"cover","position":"center","repeat":"no-repeat"}},"dark":{}}}`,
			want: Background{
				Light: ModeBackground{Image: &Image{Ref: "asset:a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}},
			},
		},
		{
			name: "no background decodes",
			raw:  `{"darkMode":false}`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeRequest([]byte(tc.raw))
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidConfig) {
					t.Fatalf("expected ErrInvalidConfig, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(got.Background, tc.want) {
				t.Errorf("background:\n got %s\nwant %s", fmtBackground(got.Background), fmtBackground(tc.want))
			}
		})
	}
}

func fmtBackground(bg Background) string {
	return fmtMode(bg.Light) + " | " + fmtMode(bg.Dark)
}

func fmtMode(mb ModeBackground) string {
	s := "{"
	if mb.Gradient != nil {
		s += fmt.Sprintf("gradient:%+v", *mb.Gradient)
	}
	if mb.Image != nil {
		s += fmt.Sprintf(" image:%+v", *mb.Image)
	}
	return s + "}"
}
