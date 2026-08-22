// Package backgrounds stores reusable dashboard backgrounds as first-class
// entities: one folder per background holding background.json plus an
// assets/ folder of uploaded images. Dashboards reference a background by
// ID. Every field other than a colour is a closed enum, so the CSS this
// package generates never contains user-authored syntax.
package backgrounds

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotFound      = errors.New("background not found")
	ErrInvalidID     = errors.New("invalid background ID")
	ErrInvalidConfig = errors.New("invalid background configuration")
)

// Color is a solid colour base, light with an optional dark override.
type Color struct {
	Light string `json:"light"`
	Dark  string `json:"dark,omitempty"`
}

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

// Background is one stored background. Color and Gradient are mutually
// exclusive; either may combine with Image. All three absent is the legal
// empty state and renders the theme background.
type Background struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Color    *Color    `json:"color,omitempty"`
	Gradient *Gradient `json:"gradient,omitempty"`
	Image    *Image    `json:"image,omitempty"`
}

// Meta is the listing representation.
type Meta struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsedBy     int    `json:"usedBy"`
	PreviewCSS string `json:"previewCss"`
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

// Validate reports whether b is safe to store and to render. It is the only
// gate between user input and generated CSS, so it must reject anything it
// does not positively recognise.
func (b Background) Validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidConfig)
	}
	if err := validateBase(b); err != nil {
		return err
	}
	if err := validateImage(b); err != nil {
		return err
	}
	return nil
}

func validateBase(b Background) error {
	if b.Color != nil && b.Gradient != nil {
		return fmt.Errorf("%w: color and gradient are mutually exclusive", ErrInvalidConfig)
	}
	if b.Color != nil {
		if err := validateBaseColor(b.Color); err != nil {
			return err
		}
	}
	if b.Gradient != nil {
		if err := validateBaseGradient(b.Gradient); err != nil {
			return err
		}
	}
	return nil
}

func validateBaseColor(c *Color) error {
	if err := validateColor(c.Light); err != nil {
		return err
	}
	if c.Dark != "" {
		if err := validateColor(c.Dark); err != nil {
			return err
		}
	}
	return nil
}

func validateBaseGradient(g *Gradient) error {
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

func validateImage(b Background) error {
	if b.Image != nil {
		if err := validateImageRefs(b.Image); err != nil {
			return err
		}
		if err := validateImageEnums(b.Image); err != nil {
			return err
		}
	}
	return nil
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

// ValidateRef checks an image reference. Only two schemes exist: asset:
// (this background's own uploads) and shared: (the admin pool). theme: and
// dashboard: are deliberately unsupported — a background is global and
// cannot reach into one dashboard's assets.
func ValidateRef(ref string) error {
	scheme, path, ok := strings.Cut(ref, ":")
	if !ok {
		return fmt.Errorf("%w: reference %q has no scheme", ErrInvalidConfig, ref)
	}
	if scheme != "asset" && scheme != "shared" {
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
	if scheme == "shared" && strings.ContainsAny(path, `/\`) {
		return fmt.Errorf("%w: the shared pool is flat", ErrInvalidConfig)
	}
	if !imageExts[strings.ToLower(filepath.Ext(cleaned))] {
		return fmt.Errorf("%w: %q is not an image", ErrInvalidConfig, filepath.Ext(cleaned))
	}
	return nil
}
