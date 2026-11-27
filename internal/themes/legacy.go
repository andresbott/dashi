package themes

import (
	"encoding/json"
	"fmt"
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

// decodeTheme decodes a stored theme.json. It is the one decoder behind both
// Store.Get and readThemeFile, so every read path sees the same theme. A
// file written before the darkMode flag existed has no key and loads with
// dark mode on: those themes always rendered their dark palette.
func decodeTheme(raw []byte) (Theme, error) {
	t, darkMode, err := decodeWire(raw)
	if err != nil {
		return Theme{}, err
	}
	t.DarkMode = darkMode == nil || *darkMode
	return t, nil
}

// DecodeRequest decodes a theme sent by a client to create or replace one.
// Unlike a stored file, the body must carry the darkMode key: a client that
// omits it predates the flag and would send the old shape, so it is refused
// with ErrInvalidConfig rather than guessed. Malformed JSON returns the
// decoder's error unwrapped.
func DecodeRequest(raw []byte) (Theme, error) {
	t, darkMode, err := decodeWire(raw)
	if err != nil {
		return Theme{}, err
	}
	if darkMode == nil {
		return Theme{}, fmt.Errorf("%w: darkMode is required", ErrInvalidConfig)
	}
	t.DarkMode = *darkMode
	return t, nil
}
