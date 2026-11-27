package themes

import (
	"encoding/json"
	"fmt"
	"slices"
)

// themeWire decodes a theme plus a probe for the darkMode key: the outer
// DarkMode shadows Theme.DarkMode for encoding/json, so a missing key stays
// nil instead of collapsing into false.
type themeWire struct {
	Theme
	DarkMode *bool `json:"darkMode"`
}

func decodeWire(raw []byte) (Theme, *bool, error) {
	var w themeWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return Theme{}, nil, err
	}
	return w.Theme, w.DarkMode, nil
}

// legacyWire is the background shape stored before per-mode backgrounds: one
// gradient and one image whose direction and fit/position/repeat were shared
// by both modes, with optional dark stops and a dark image.
type legacyWire struct {
	Background struct {
		Gradient *struct {
			Direction string   `json:"direction"`
			Light     []string `json:"light"`
			Dark      []string `json:"dark"`
		} `json:"gradient"`
		Image *struct {
			Light    string `json:"light"`
			Dark     string `json:"dark"`
			Fit      string `json:"fit"`
			Position string `json:"position"`
			Repeat   string `json:"repeat"`
		} `json:"image"`
	} `json:"background"`
}

// decodeTheme decodes a stored theme.json. It is the one decoder behind both
// Store.Get and readThemeFile, so every read path sees the same theme. A
// file written before the darkMode flag existed has no key and loads with
// dark mode on: those themes always rendered their dark palette. A file with
// the legacy background shape is converted in memory; it is rewritten in the
// new shape only on its next save.
func decodeTheme(raw []byte) (Theme, error) {
	t, darkMode, err := decodeWire(raw)
	if err != nil {
		return Theme{}, err
	}
	t.DarkMode = darkMode == nil || *darkMode
	if t.Background == (Background{}) {
		var lw legacyWire
		if err := json.Unmarshal(raw, &lw); err != nil {
			return Theme{}, err
		}
		t.Background = lw.background()
	}
	return t, nil
}

// background converts the legacy shape. Light takes the shared settings and
// the light values; dark is a deep copy of light with the legacy dark stops
// and dark image applied on top, so a theme keeps rendering as before.
func (lw legacyWire) background() Background {
	var bg Background
	if g := lw.Background.Gradient; g != nil {
		bg.Light.Gradient = &Gradient{Direction: g.Direction, Stops: slices.Clone(g.Light)}
		dark := g.Light
		if len(g.Dark) > 0 {
			dark = g.Dark
		}
		bg.Dark.Gradient = &Gradient{Direction: g.Direction, Stops: slices.Clone(dark)}
	}
	if img := lw.Background.Image; img != nil {
		light := Image{Ref: img.Light, Fit: img.Fit, Position: img.Position, Repeat: img.Repeat}
		dark := light
		if img.Dark != "" {
			dark.Ref = img.Dark
		}
		bg.Light.Image, bg.Dark.Image = &light, &dark
	}
	return bg
}

// DecodeRequest decodes a theme sent by a client to create or replace one.
// Unlike a stored file, the body must carry the darkMode key: a client that
// omits it predates the flag and would send the old shape, so it is refused
// with ErrInvalidConfig rather than guessed. A background in the legacy
// gradient/image shape is refused the same way: migration is a load-time path
// for stored files, and decoding it as the new shape would silently drop it.
// Malformed JSON returns the decoder's error unwrapped.
func DecodeRequest(raw []byte) (Theme, error) {
	t, darkMode, err := decodeWire(raw)
	if err != nil {
		return Theme{}, err
	}
	if darkMode == nil {
		return Theme{}, fmt.Errorf("%w: darkMode is required", ErrInvalidConfig)
	}
	var lw legacyWire
	if err := json.Unmarshal(raw, &lw); err != nil {
		return Theme{}, err
	}
	if lw.Background.Gradient != nil || lw.Background.Image != nil {
		return Theme{}, fmt.Errorf("%w: background must use the light/dark shape, not gradient/image", ErrInvalidConfig)
	}
	t.DarkMode = *darkMode
	return t, nil
}
