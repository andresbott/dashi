package themes

import (
	"net/url"
	"strings"
)

// variant is one colour mode's fully resolved background.
type variant struct {
	image     *Image // the image layer, or nil when there is none
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
	if bg := browserValue(t.ID, resolve(t.bgFor(colorMode), nil)); bg != "" {
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
	v := resolve(t.bgFor(colorMode), func(stop string) string { return flattenOver(stop, bg) })
	if v.image != nil {
		return "", v.image.Ref
	}
	if !safeBaseValue(v.baseValue) {
		return "", ""
	}
	return v.baseValue, ""
}

// BackgroundImageURL returns the URL of colorMode's background image, as
// BrowserCSS references it, or "" when that mode has no image (or one
// BrowserCSS would not emit). The browser page preloads it, so the image
// starts downloading with the HTML instead of after the theme stylesheet.
func BackgroundImageURL(t Theme, colorMode string) string {
	img := resolve(t.bgFor(colorMode), nil).image
	if img == nil || validateImageEnums(img) != nil {
		return ""
	}
	u, ok := RefURL(t.ID, img.Ref)
	if !ok {
		return ""
	}
	return u
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

// resolve builds one colour mode's background. Callers pass the mode's own
// ModeBackground from Theme.bgFor; there is no fallback to the other mode.
// mapStop, when non-nil, rewrites each gradient stop before it is joined.
func resolve(mb ModeBackground, mapStop func(string) string) variant {
	var v variant
	if mb.Image != nil && mb.Image.Ref != "" {
		v.image = mb.Image
	}
	if g := mb.Gradient; g != nil && len(g.Stops) > 0 {
		stops := g.Stops
		if mapStop != nil {
			stops = make([]string, len(g.Stops))
			for i, s := range g.Stops {
				stops[i] = mapStop(s)
			}
		}
		v.baseValue = "linear-gradient(" + g.Direction + "," + strings.Join(stops, ",") + ")"
	}
	return v
}

// browserValue builds the CSS background shorthand for one variant: the
// image layer first, using that mode's own fit, position and repeat, and the
// gradient as the final layer. Returns "" for a plain background, which then
// shows the palette bg.
//
// The image's enum fields are checked here too: Validate rejects bad values
// on save, but a hand-edited theme.json bypasses it, and this output lands
// in a <style> element where html/template escapes nothing.
func browserValue(themeID string, v variant) string {
	layers := make([]string, 0, 2)
	if img := v.image; img != nil && validateImageEnums(img) == nil {
		if u, ok := RefURL(themeID, img.Ref); ok {
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
