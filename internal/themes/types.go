// Package themes stores dashboard themes as first-class entities: a
// light/dark colour palette plus a page background. Each stored theme is one
// folder holding theme.json and an assets/ folder of uploaded images; the
// read-only builtin themes ("default", "ember") live in code and are never
// written to disk. Dashboards reference a theme by ID. Every field other than a colour
// is a closed enum, so the CSS this package generates never contains
// user-authored syntax.
package themes

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// DefaultID is the ID of the builtin theme.
const DefaultID = "default"

var (
	ErrNotFound      = errors.New("theme not found")
	ErrInvalidID     = errors.New("invalid theme ID")
	ErrInvalidConfig = errors.New("invalid theme configuration")
	ErrBuiltin       = errors.New("the builtin theme is read-only")
)

// Gradient is a linear-gradient base. Direction is shared by both modes;
// only the colour stops vary.
type Gradient struct {
	Direction string   `json:"direction"`
	Light     []string `json:"light"`
	Dark      []string `json:"dark,omitempty"`
}

// Image is an image layer painted over the base. Fit, Position and Repeat
// are shared by both modes; only the image itself varies.
type Image struct {
	Light    string `json:"light"`
	Dark     string `json:"dark,omitempty"`
	Fit      string `json:"fit"`
	Position string `json:"position"`
	Repeat   string `json:"repeat"`
}

// Background is a theme's page background. Gradient and Image may combine,
// the image painted over the gradient. Both absent is the legal plain state:
// the page shows the palette bg. There is no solid colour: the base is
// always Palette.BG.
type Background struct {
	Gradient *Gradient `json:"gradient,omitempty"`
	Image    *Image    `json:"image,omitempty"`
}

// Theme is one theme: a background plus a palette per colour mode.
//
// DarkMode says whether the theme has a dark variant. When it is false every
// colour mode renders the light palette and background; the dark data stays
// stored but unused, so switching it back on loses nothing. The tag has no
// omitempty on purpose: a file without the key loads with dark mode on (see
// decodeTheme), so a saved false must be written out.
type Theme struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	DarkMode   bool       `json:"darkMode"`
	Background Background `json:"background"`
	Colors     Colors     `json:"colors"`
}

// Meta is the listing representation. Colors is the rendered view: with
// DarkMode off its dark palette is the light one.
type Meta struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Builtin    bool   `json:"builtin,omitempty"`
	DarkMode   bool   `json:"darkMode"`
	UsedBy     int    `json:"usedBy"`
	PreviewCSS string `json:"previewCss"`
	Colors     Colors `json:"colors"`
}

// DefaultTheme returns the builtin theme: the default palettes over a plain
// background.
func DefaultTheme() Theme {
	return Theme{
		ID:       DefaultID,
		Name:     "Default",
		DarkMode: true,
		Colors: Colors{
			Light: DefaultPalette("light"),
			Dark:  DefaultPalette("dark"),
		},
	}
}

// mode maps a requested colour mode onto the variant the theme renders:
// "dark" only when the theme has dark mode on and dark was asked for, "light"
// for everything else. Every per-mode lookup goes through it.
func (t Theme) mode(colorMode string) string {
	if t.DarkMode && colorMode == "dark" {
		return "dark"
	}
	return "light"
}

// Palette returns the palette rendered for colorMode (see mode), leniently:
// every role that is not a valid colour falls back to the default, so data
// that bypassed Validate still renders and cannot inject.
func (t Theme) Palette(colorMode string) Palette {
	if t.mode(colorMode) == "dark" {
		return t.Colors.Dark.withDefaults("dark")
	}
	return t.Colors.Light.withDefaults("light")
}

var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

var gradientKeywords = map[string]bool{
	"to top": true, "to bottom": true, "to left": true, "to right": true,
	"to top left": true, "to top right": true,
	"to bottom left": true, "to bottom right": true,
}

var fitValues = map[string]bool{
	"cover": true, "contain": true, "stretch": true, "original": true,
}

var positionValues = map[string]bool{
	"center": true, "top": true, "bottom": true, "left": true, "right": true,
	"top left": true, "top right": true, "bottom left": true, "bottom right": true,
}

var repeatValues = map[string]bool{
	"no-repeat": true, "repeat": true, "repeat-x": true, "repeat-y": true,
}

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".svg": true,
}

// Validate reports whether t is safe to store and to render. It is the only
// gate between user input and generated CSS, so it must reject anything it
// does not positively recognise. Every light palette role is required; the
// dark palette is required in full only with dark mode on. With it off, the
// dark roles that are set are still checked, because they stay stored and
// render again once dark mode is switched back on.
func (t Theme) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidConfig)
	}
	if err := t.Colors.Light.Validate(); err != nil {
		return fmt.Errorf("light palette: %w", err)
	}
	validateDark := t.Colors.Dark.Validate
	if !t.DarkMode {
		validateDark = t.Colors.Dark.validateSet
	}
	if err := validateDark(); err != nil {
		return fmt.Errorf("dark palette: %w", err)
	}
	if err := validateGradient(t.Background.Gradient); err != nil {
		return err
	}
	if err := validateImage(t.Background.Image); err != nil {
		return err
	}
	return nil
}

func validateGradient(g *Gradient) error {
	if g == nil {
		return nil
	}
	if err := validateDirection(g.Direction); err != nil {
		return err
	}
	if err := validateStops(g.Light); err != nil {
		return err
	}
	if len(g.Dark) > 0 {
		if err := validateStops(g.Dark); err != nil {
			return err
		}
	}
	return nil
}

func validateImage(img *Image) error {
	if img == nil {
		return nil
	}
	if err := validateImageRefs(img); err != nil {
		return err
	}
	return validateImageEnums(img)
}

func validateImageRefs(img *Image) error {
	if err := ValidateRef(img.Light); err != nil {
		return err
	}
	if img.Dark != "" {
		if err := ValidateRef(img.Dark); err != nil {
			return err
		}
	}
	return nil
}

func validateImageEnums(img *Image) error {
	if !fitValues[img.Fit] {
		return fmt.Errorf("%w: unknown fit %q", ErrInvalidConfig, img.Fit)
	}
	if !positionValues[img.Position] {
		return fmt.Errorf("%w: unknown position %q", ErrInvalidConfig, img.Position)
	}
	if !repeatValues[img.Repeat] {
		return fmt.Errorf("%w: unknown repeat %q", ErrInvalidConfig, img.Repeat)
	}
	return nil
}

func validateColor(c string) error {
	if !hexColorRe.MatchString(c) {
		return fmt.Errorf("%w: %q is not a hex colour", ErrInvalidConfig, c)
	}
	return nil
}

func validateDirection(d string) error {
	if gradientKeywords[d] {
		return nil
	}
	if rest, ok := strings.CutSuffix(d, "deg"); ok {
		n, err := strconv.Atoi(rest)
		if err == nil && n >= 0 && n <= 360 {
			return nil
		}
	}
	return fmt.Errorf("%w: unknown gradient direction %q", ErrInvalidConfig, d)
}

func validateStops(stops []string) error {
	if len(stops) < 2 {
		return fmt.Errorf("%w: a gradient needs at least 2 colour stops", ErrInvalidConfig)
	}
	for _, s := range stops {
		if err := validateColor(s); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRef checks an image reference. Only one scheme exists: asset:,
// this theme's own uploads, so a theme is self-contained and a copy carries
// every image it uses. theme:, dashboard: and the former shared: pool are
// deliberately unsupported.
func ValidateRef(ref string) error {
	scheme, path, ok := strings.Cut(ref, ":")
	if !ok {
		return fmt.Errorf("%w: reference %q has no scheme", ErrInvalidConfig, ref)
	}
	if scheme != "asset" {
		return fmt.Errorf("%w: unsupported reference scheme %q", ErrInvalidConfig, scheme)
	}
	if path == "" {
		return fmt.Errorf("%w: empty reference path", ErrInvalidConfig)
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("%w: absolute paths not allowed", ErrInvalidConfig)
	}
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return fmt.Errorf("%w: path traversal not allowed", ErrInvalidConfig)
	}
	if !imageExts[strings.ToLower(filepath.Ext(cleaned))] {
		return fmt.Errorf("%w: %q is not an image", ErrInvalidConfig, filepath.Ext(cleaned))
	}
	return nil
}
