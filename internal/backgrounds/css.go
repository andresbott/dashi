package backgrounds

import (
	"net/url"
	"strings"
)

// variant is one colour mode's fully resolved background.
type variant struct {
	image     string // ref, or "" when there is no image layer
	baseValue string // concrete colour or linear-gradient(...), or ""
}

// BrowserCSS returns the body of the <style> block the browser page shell
// emits: --dashi-page-bg for light on :root, for dark under the
// [data-color-mode="dark"] attribute the shell sets on <html>. This mirrors
// themes.ThemeCSS deliberately — same mechanism, same override order.
//
// Returns nil when b defines no background, so the caller emits no element
// at all and the theme background shows through.
func BrowserCSS(b Background) []byte {
	light := browserValue(b, resolve(b, "light"))
	dark := browserValue(b, resolve(b, "dark"))
	if light == "" && dark == "" {
		return nil
	}
	var sb strings.Builder
	sb.WriteString(":root{--dashi-page-bg:")
	sb.WriteString(light)
	sb.WriteString(";}\n")
	sb.WriteString(`:root[data-color-mode="dark"]{--dashi-page-bg:`)
	sb.WriteString(dark)
	sb.WriteString(";}\n")
	return []byte(sb.String())
}

// ImageCSS resolves one variant for the litehtml stack, which cannot read
// CSS custom properties and needs concrete values.
//
// When an image is present it returns an empty css string and the image
// ref: internal/dashboard/image pre-paints the image onto the canvas
// *before* litehtml renders, so an opaque body background would paint over
// it. Nothing is lost — that stack is cover-only, so the image always
// covers the canvas and a base could never show.
func ImageCSS(b Background, colorMode string) (css string, imageRef string) {
	v := resolve(b, colorMode)
	if v.image != "" {
		return "", v.image
	}
	return v.baseValue, ""
}

// RefURL maps an image reference onto the HTTP route that serves it.
func RefURL(bgID, ref string) (string, bool) {
	scheme, path, ok := strings.Cut(ref, ":")
	if !ok || path == "" {
		return "", false
	}
	switch scheme {
	case "asset":
		segments := strings.Split(path, "/")
		for i, s := range segments {
			segments[i] = urlSegment(s)
		}
		return "/api/v0/backgrounds/" + urlSegment(bgID) + "/assets/" + strings.Join(segments, "/"), true
	case "shared":
		return "/api/v0/data/backgrounds/" + urlSegment(path), true
	}
	return "", false
}

// urlSegment percent-encodes one path segment and additionally escapes the
// single quote, which url.PathEscape leaves alone but which would terminate
// the url('…') value this package builds.
func urlSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "'", "%27")
}

// resolve picks the values for one colour mode. Any mode other than "dark"
// resolves light, matching how the renderers collapse colorMode. Dark falls
// back to light per field.
func resolve(b Background, colorMode string) variant {
	dark := colorMode == "dark"
	var v variant

	if b.Image != nil {
		v.image = b.Image.Light
		if dark && b.Image.Dark != "" {
			v.image = b.Image.Dark
		}
	}
	switch {
	case b.Color != nil:
		v.baseValue = b.Color.Light
		if dark && b.Color.Dark != "" {
			v.baseValue = b.Color.Dark
		}
	case b.Gradient != nil:
		stops := b.Gradient.Light
		if dark && len(b.Gradient.Dark) > 0 {
			stops = b.Gradient.Dark
		}
		if len(stops) > 0 {
			v.baseValue = "linear-gradient(" + b.Gradient.Direction + "," + strings.Join(stops, ",") + ")"
		}
	}
	return v
}

// browserValue builds the CSS background shorthand for one variant: the
// image layer first, the base as the final layer. Only the final layer of
// the shorthand may carry a colour, which is why the base goes last.
func browserValue(b Background, v variant) string {
	layers := make([]string, 0, 2)
	if v.image != "" && b.Image != nil {
		if u, ok := RefURL(b.ID, v.image); ok {
			layers = append(layers, "url('"+u+"') "+
				b.Image.Position+"/"+cssSize(b.Image.Fit)+" "+b.Image.Repeat)
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

// safeBaseValue is a defence-in-depth check for hand-edited background.json
// files that bypassed Validate. Generated CSS lands inside a <style>
// element, where html/template performs no escaping at all.
func safeBaseValue(s string) bool {
	if hexColorRe.MatchString(s) {
		return true
	}
	if !strings.HasPrefix(s, "linear-gradient(") || !strings.HasSuffix(s, ")") {
		return false
	}
	return !strings.ContainsAny(s, "<>;{}\"'\\")
}
