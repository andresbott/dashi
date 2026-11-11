package themes

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Palette is the set of colour roles for one colour mode. Every value is a
// hex colour; BG must be opaque because it is the page base everything else
// is composited onto.
type Palette struct {
	BG        string `json:"bg"`
	Surface   string `json:"surface"`
	Border    string `json:"border"`
	Text      string `json:"text"`
	TextMuted string `json:"textMuted"`
	Primary   string `json:"primary"`
	OnPrimary string `json:"onPrimary"`
	Secondary string `json:"secondary"`
	Success   string `json:"success"`
	Warning   string `json:"warning"`
	Danger    string `json:"danger"`
}

// Colors holds the palette for each colour mode.
type Colors struct {
	Light Palette `json:"light"`
	Dark  Palette `json:"dark"`
}

var opaqueHexRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// role binds a role's JSON name and CSS variable to its field in a palette.
type role struct {
	name   string
	cssVar string
	field  func(p *Palette) *string
	opaque bool
}

// paletteRoles is in CSS variable output order.
var paletteRoles = []role{
	{"bg", "--dashi-bg", func(p *Palette) *string { return &p.BG }, true},
	{"surface", "--dashi-surface", func(p *Palette) *string { return &p.Surface }, false},
	{"border", "--dashi-border", func(p *Palette) *string { return &p.Border }, false},
	{"text", "--dashi-text", func(p *Palette) *string { return &p.Text }, false},
	{"textMuted", "--dashi-text-muted", func(p *Palette) *string { return &p.TextMuted }, false},
	{"primary", "--dashi-primary", func(p *Palette) *string { return &p.Primary }, false},
	{"onPrimary", "--dashi-on-primary", func(p *Palette) *string { return &p.OnPrimary }, false},
	{"secondary", "--dashi-secondary", func(p *Palette) *string { return &p.Secondary }, false},
	{"success", "--dashi-success", func(p *Palette) *string { return &p.Success }, false},
	{"warning", "--dashi-warning", func(p *Palette) *string { return &p.Warning }, false},
	{"danger", "--dashi-danger", func(p *Palette) *string { return &p.Danger }, false},
}

func (r role) valid(v string) bool {
	if r.opaque {
		return opaqueHexRe.MatchString(v)
	}
	return hexColorRe.MatchString(v)
}

// DefaultPalette returns the built-in palette: dark for "dark", light for
// anything else.
func DefaultPalette(colorMode string) Palette {
	if colorMode == "dark" {
		return Palette{
			BG: "#1a1a2e", Surface: "#23233d", Border: "#33334d",
			Text: "#e0e0e0", TextMuted: "#999999",
			Primary: "#4a90d9", OnPrimary: "#ffffff",
			Secondary: "#a78bfa", Success: "#22c55e", Warning: "#f59e0b", Danger: "#ef4444",
		}
	}
	return Palette{
		BG: "#ffffff", Surface: "#f5f5f7", Border: "#e0e0e5",
		Text: "#333333", TextMuted: "#666666",
		Primary: "#0a84ff", OnPrimary: "#ffffff",
		Secondary: "#8b5cf6", Success: "#16a34a", Warning: "#d97706", Danger: "#dc2626",
	}
}

// Validate reports whether every role holds a valid hex colour and BG is
// opaque. Errors wrap ErrInvalidConfig and name the offending role.
func (p Palette) Validate() error {
	for _, r := range paletteRoles {
		v := *r.field(&p)
		if r.valid(v) {
			continue
		}
		if r.opaque {
			return fmt.Errorf("%w: %s must be an opaque hex colour, got %q", ErrInvalidConfig, r.name, v)
		}
		return fmt.Errorf("%w: %s must be a hex colour, got %q", ErrInvalidConfig, r.name, v)
	}
	return nil
}

// withDefaults keeps each valid role and replaces every other one with the
// default for colorMode, so malformed stored data can never reach CSS.
func (p Palette) withDefaults(colorMode string) Palette {
	def := DefaultPalette(colorMode)
	for _, r := range paletteRoles {
		dst := r.field(&p)
		if !r.valid(*dst) {
			*dst = *r.field(&def)
		}
	}
	return p
}

// Flatten composites every #rrggbbaa value over BG and returns it as a
// lowercase #rrggbb. Other values are left unchanged. It is for consumers
// that cannot render alpha, such as the e-ink pipeline.
func (p Palette) Flatten() Palette {
	bg, ok := parseOpaqueHex(p.BG)
	if !ok {
		return p
	}
	for _, r := range paletteRoles {
		dst := r.field(&p)
		v := *dst
		if len(v) != 9 || !hexColorRe.MatchString(v) {
			continue
		}
		var ch [4]int
		for i := range ch {
			n, err := strconv.ParseUint(v[1+2*i:3+2*i], 16, 8)
			if err != nil {
				break
			}
			ch[i] = int(n)
		}
		a := float64(ch[3]) / 255
		blend := func(fg, base int) int {
			return int(math.Round(float64(fg)*a + float64(base)*(1-a)))
		}
		*dst = fmt.Sprintf("#%02x%02x%02x", blend(ch[0], bg[0]), blend(ch[1], bg[1]), blend(ch[2], bg[2]))
	}
	return p
}

// parseOpaqueHex parses #rgb or #rrggbb into its channels.
func parseOpaqueHex(s string) ([3]int, bool) {
	var out [3]int
	if !opaqueHexRe.MatchString(s) {
		return out, false
	}
	h := s[1:]
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	for i := range out {
		n, err := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
		if err != nil {
			return out, false
		}
		out[i] = int(n)
	}
	return out, true
}

// writePaletteVars writes the palette as CSS custom declarations. Callers
// must pass a validated or withDefaults-sanitised palette.
func writePaletteVars(b *strings.Builder, p Palette) {
	for _, r := range paletteRoles {
		b.WriteString(r.cssVar)
		b.WriteByte(':')
		b.WriteString(*r.field(&p))
		b.WriteByte(';')
	}
}
