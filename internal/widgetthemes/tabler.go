package widgetthemes

import (
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"
)

// tablerGlyphs maps every glyph name in the embedded Tabler icon font
// ("brand-github") to its code point. The font names its glyphs after the
// icons, so this covers every icon the editor's picker offers, not only the
// handful the default manifest maps. Built once, on first use.
var tablerGlyphs = sync.OnceValue(func() map[string]rune {
	glyphs := make(map[string]rune)
	f, err := sfnt.Parse(defaultFontTTF)
	if err != nil {
		return glyphs
	}
	var b sfnt.Buffer
	// Tabler starts in the BMP private-use area and has long since run past
	// its end (U+F8FF) into the code points above it, so scan to the end of
	// the BMP, plus supplementary plane 15 for future releases.
	ranges := [][2]rune{{0xe000, 0xfffd}, {0xf0000, 0xffffd}}
	for _, rg := range ranges {
		for r := rg[0]; r <= rg[1]; r++ {
			idx, err := f.GlyphIndex(&b, r)
			if err != nil || idx == 0 {
				continue
			}
			name, err := f.GlyphName(&b, idx)
			if err != nil || name == "" {
				continue
			}
			if _, seen := glyphs[name]; !seen {
				glyphs[name] = r
			}
		}
	}
	return glyphs
})

// tablerCodepoint returns the hex code point (as manifests write it) of a
// Tabler icon in the embedded font. name may carry the "ti-" class prefix
// the editor stores ("ti-brand-github") or not ("brand-github").
func tablerCodepoint(name string) (string, bool) {
	name = strings.TrimPrefix(name, "ti-")
	if name == "" {
		return "", false
	}
	r, ok := tablerGlyphs()[name]
	if !ok {
		return "", false
	}
	return strconv.FormatInt(int64(r), 16), true
}
