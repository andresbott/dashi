package dashboard

import (
	"regexp"
	"strings"
	"testing"
)

func TestSanitizeCustomCSSNeutralisesStyleEndTags(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"plain", "</style><script>alert(1)</script>"},
		{"uppercase", "</STYLE><script>alert(1)</script>"},
		{"mixed case", "</Style ><script>alert(1)</script>"},
		{"newline before gt", "</style\n><script>alert(1)</script>"},
		{"space after slash", "</ style><script>alert(1)</script>"},
		{"tab", "</\tstyle><script>alert(1)</script>"},
	}

	// Mirrors how an HTML parser leaves a raw-text element: any ASCII-case
	// "</style" ends it, whitespace between the slash and the name included.
	breakout := regexp.MustCompile(`(?i)</\s*style`)

	for _, tc := range cases {
		got := SanitizeCustomCSS(tc.in)
		if breakout.MatchString(got) {
			t.Errorf("%s: sanitized value still contains a style end tag: %q", tc.name, got)
		}
	}
}

func TestSanitizeCustomCSSLeavesLegitimateCSSAlone(t *testing.T) {
	css := ".dashi-clock__time{color:red}\n@media (width < 600px){.dashi-row{display:block}}"
	if got := SanitizeCustomCSS(css); got != css {
		t.Errorf("legitimate CSS was modified:\ngot:  %q\nwant: %q", got, css)
	}
	if got := SanitizeCustomCSS(""); got != "" {
		t.Errorf("empty input should stay empty, got %q", got)
	}
}

func TestSanitizeCustomCSSKeepsSurroundingRules(t *testing.T) {
	got := SanitizeCustomCSS(".a{color:red}</style>.b{color:blue}")
	if !strings.Contains(got, ".a{color:red}") || !strings.Contains(got, ".b{color:blue}") {
		t.Errorf("sanitizer must only replace the end tag, got %q", got)
	}
}
