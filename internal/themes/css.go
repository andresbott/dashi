package themes

import (
	"net/url"
	"strings"
)

// variant is one colour mode's fully resolved background.
type variant struct {
	image     string // ref, or "" when there is no image layer
	baseValue string // linear-gradient(...), or "" for the plain palette bg
}

// BrowserCSS returns the body of the <style> block the browser page shell
// emits: the palette vars, plus --dashi-page-bg when the theme has a
// gradient or image, for light on :root and for dark under the
// [data-color-mode="dark"] attribute the shell sets on <html>.
//
// It is always non-nil. Palettes go through Theme.Palette, so a malformed
// role falls back to its default instead of reaching the CSS.
func BrowserCSS(t Theme) []byte {
	var sb strings.Builder
	writeRule(&sb, ":root", t, "light")
	writeRule(&sb, `:root[data-color-mode="dark"]`, t, "dark")
	return []byte(sb.String())
}

func writeRule(sb *strings.Builder, selector string, t Theme, colorMode string) {
	sb.WriteString(selector)
	sb.WriteByte('{')
	writePaletteVars(sb, t.Palette(colorMode))
	if bg := browserValue(t, resolve(t.Background, t.mode(colorMode), nil)); bg != "" {
		sb.WriteString("--dashi-page-bg:")
		sb.WriteString(bg)
		sb.WriteByte(';')
	}
	sb.WriteString("}\n")
}

// ImageCSS resolves one variant for the litehtml stack, which cannot read
// CSS custom properties and needs concrete values. A plain background
// returns two empty strings: the caller paints the palette bg.
//
// When an image is present it returns an empty css string and the image
// ref: internal/dashboard/image pre-paints the image onto the canvas
// *before* litehtml renders, so an opaque body background would paint over
// it. Nothing is lost — that stack is cover-only, so the image always
// covers the canvas and a base could never show.
//
// litehtml cannot parse #rrggbbaa, so translucent gradient stops are
// flattened over this mode's palette bg first.
func ImageCSS(t Theme, colorMode string) (css string, imageRef string) {
	bg, _ := parseOpaqueHex(t.Palette(colorMode).BG) // Palette guarantees an opaque BG
	v := resolve(t.Background, t.mode(colorMode), func(stop string) string { return flattenOver(stop, bg) })
	if v.image != "" {
		return "", v.image
	}
	if !safeBaseValue(v.baseValue) {
		return "", ""
	}
	return v.baseValue, ""
}

// RefURL maps an image reference onto the HTTP route that serves it.
func RefURL(themeID, ref string) (string, bool) {
	scheme, path, ok := strings.Cut(ref, ":")
	if !ok || path == "" {
		return "", false
	}
	if scheme != "asset" {
		return "", false
	}
	segments := strings.Split(path, "/")
	for i, s := range segments {
		segments[i] = urlSegment(s)
	}
	return "/api/v0/themes/" + urlSegment(themeID) + "/assets/" + strings.Join(segments, "/"), true
}

// urlSegment percent-encodes one path segment and additionally escapes the
// single quote, which would terminate the url('…') value this package builds.
// url.PathEscape does escape the single quote to %27, but this extra call
// is defence in depth should the encoding ever change.
func urlSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "'", "%27")
}

// resolve picks the values for one colour mode; callers pass the effective
// mode from Theme.mode, so a theme with dark mode off resolves light. Dark falls
// back to light per field. mapStop, when non-nil, rewrites each gradient
// stop before it is joined.
func resolve(b Background, colorMode string, mapStop func(string) string) variant {
	dark := colorMode == "dark"
	var v variant

	if b.Image != nil {
		v.image = b.Image.Light
		if dark && b.Image.Dark != "" {
			v.image = b.Image.Dark
		}
	}
	if b.Gradient != nil {
		stops := b.Gradient.Light
		if dark && len(b.Gradient.Dark) > 0 {
			stops = b.Gradient.Dark
		}
		if len(stops) > 0 && mapStop != nil {
			mapped := make([]string, len(stops))
			for i, s := range stops {
				mapped[i] = mapStop(s)
			}
			stops = mapped
		}
		if len(stops) > 0 {
			v.baseValue = "linear-gradient(" + b.Gradient.Direction + "," + strings.Join(stops, ",") + ")"
		}
	}
	return v
}

// browserValue builds the CSS background shorthand for one variant: the
// image layer first, the gradient as the final layer. Returns "" for a
// plain background, which then shows the palette bg.
//
// The image's enum fields are checked here too: Validate rejects bad values
// on save, but a hand-edited theme.json bypasses it, and this output lands
// in a <style> element where html/template escapes nothing.
func browserValue(t Theme, v variant) string {
	layers := make([]string, 0, 2)
	img := t.Background.Image
	if v.image != "" && img != nil && validateImageEnums(img) == nil {
		if u, ok := RefURL(t.ID, v.image); ok {
			layers = append(layers, "url('"+u+"') "+
				img.Position+"/"+cssSize(img.Fit)+" "+img.Repeat)
		}
	}
	if v.baseValue != "" && safeBaseValue(v.baseValue) {
		layers = append(layers, v.baseValue)
	}
	return strings.Join(layers, ",")
}

// cssSize maps a fit mode onto a background-size value.
func cssSize(fit string) string {
	switch fit {
	case "contain":
		return "contain"
	case "stretch":
		return "100% 100%"
	case "original":
		return "auto"
	default:
		return "cover"
	}
}

// safeBaseValue is a defence-in-depth check for hand-edited theme.json
// files that bypassed Validate. Generated CSS lands inside a <style>
// element, where html/template performs no escaping at all.
func safeBaseValue(s string) bool {
	if !strings.HasPrefix(s, "linear-gradient(") || !strings.HasSuffix(s, ")") {
		return false
	}
	return !strings.ContainsAny(s, "<>;{}\"'\\")
}
