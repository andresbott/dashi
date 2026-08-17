package dashboard

import "regexp"

// styleCloseRe matches an HTML end-tag for <style> in every form an HTML
// parser accepts: any ASCII case, and any whitespace (or nothing) between
// the tag name and `>` — including newlines, as in "</style\n>". The `>`
// is deliberately not required, because a parser inside a raw-text element
// terminates on `</style` followed by whitespace, `/` or `>`.
var styleCloseRe = regexp.MustCompile(`(?i)</\s*style`)

// SanitizeCustomCSS neutralises `</style>` sequences in a dashboard's
// owner-uploaded custom.css.
//
// Both render stacks inject custom.css into a <style> element. html/template
// applies no escaping inside a <style> element (unlike a quoted attribute,
// which still gets the attribute escaper), so a stylesheet containing
// `</style><script>…</script>` would break out of the element and execute in
// a real browser. custom.css only reaches this path through the deliberately
// trusted editor, but before the vanilla-viewer refactor it never reached a
// browser at all, so the surface is new and worth closing.
//
// The replacement keeps the file valid CSS-ish rather than rejecting the
// whole stylesheet: a broken selector is a better failure mode for a
// dashboard owner than a blank page.
func SanitizeCustomCSS(css string) string {
	if css == "" {
		return ""
	}
	return styleCloseRe.ReplaceAllString(css, "/* blocked */")
}
