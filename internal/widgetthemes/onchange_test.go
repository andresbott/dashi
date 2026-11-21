package widgetthemes

import (
	"bytes"
	"sync"
	"testing"
)

const fontPackYAML = `name: "fancy"
fonts:
  - name: "Fancy Sans"
    file: "fancy.ttf"
icons:
  type: font
  classPrefix: "fi-"
  fontFile: "icons.ttf"
  icons:
    sun:
      class: "sun"
      codepoint: "e001"
`

func uploadFontPack(t *testing.T, s *Store, displayBytes, iconBytes string) {
	t.Helper()
	zipBytes := buildThemeZip(t, fontPackYAML, map[string]string{
		"fancy.ttf": displayBytes,
		"icons.ttf": iconBytes,
	})
	if _, err := s.Upload(zipBytes); err != nil {
		t.Fatalf("upload: %v", err)
	}
}

func TestAllFontsDefaultOnly(t *testing.T) {
	fonts := NewStore("").AllFonts()
	if !bytes.Equal(fonts["Inter"], interRegularTTF) {
		t.Error("default display font Inter missing")
	}
	if !bytes.Equal(fonts[IconFontFamily(DefaultName)], defaultFontTTF) {
		t.Error("default icon font missing")
	}
	if len(fonts) != 2 {
		t.Errorf("got %d fonts, want 2: %v", len(fonts), keys(fonts))
	}
}

func TestAllFontsTracksUploadReplaceAndDelete(t *testing.T) {
	s := NewStore(t.TempDir())
	uploadFontPack(t, s, "display-v1", "icons-v1")

	fonts := s.AllFonts()
	if string(fonts["Fancy Sans"]) != "display-v1" {
		t.Errorf("display font = %q, want display-v1", fonts["Fancy Sans"])
	}
	if string(fonts[IconFontFamily("fancy")]) != "icons-v1" {
		t.Errorf("icon font = %q, want icons-v1", fonts[IconFontFamily("fancy")])
	}

	// Replace = delete + re-upload under the same name.
	if err := s.Delete("fancy"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	fonts = s.AllFonts()
	if _, ok := fonts["Fancy Sans"]; ok {
		t.Error("deleted pack's display font still listed")
	}
	if _, ok := fonts[IconFontFamily("fancy")]; ok {
		t.Error("deleted pack's icon font still listed")
	}

	uploadFontPack(t, s, "display-v2", "icons-v2")
	if got := string(s.AllFonts()["Fancy Sans"]); got != "display-v2" {
		t.Errorf("after re-upload display font = %q, want display-v2", got)
	}
}

func TestOnChangeRunsAfterUploadAndDelete(t *testing.T) {
	s := NewStore(t.TempDir())
	var seen []bool // whether "fancy" was visible when each call ran
	s.OnChange(func() {
		_, ok := s.Get("fancy")
		seen = append(seen, ok)
	})

	uploadFontPack(t, s, "d", "i")
	if err := s.Delete("fancy"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(seen) != 2 || !seen[0] || seen[1] {
		t.Fatalf("listener saw %v, want [true false] (change visible when called)", seen)
	}
}

func TestOnChangeNotCalledOnFailedWrites(t *testing.T) {
	s := NewStore(t.TempDir())
	calls := 0
	s.OnChange(func() { calls++ })

	if _, err := s.Upload([]byte("not a zip")); err == nil {
		t.Fatal("expected upload error")
	}
	if err := s.Delete("missing"); err == nil {
		t.Fatal("expected delete error")
	}
	if err := s.Delete(DefaultName); err == nil {
		t.Fatal("expected builtin delete error")
	}
	if calls != 0 {
		t.Errorf("listener called %d times on failed writes, want 0", calls)
	}
}

// TestOnChangeListenersAreSerialised runs uploads concurrently; the
// listener's unguarded counter is only race-free (under -race) if calls
// never overlap.
func TestOnChangeListenersAreSerialised(t *testing.T) {
	s := NewStore(t.TempDir())
	calls := 0
	s.OnChange(func() {
		calls++
		_ = s.AllFonts() // listeners may read the store without deadlocking
	})

	names := []string{"p1", "p2", "p3", "p4", "p5", "p6"}
	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			zipBytes := buildThemeZip(t, "name: \""+name+"\"\n", nil)
			if _, err := s.Upload(zipBytes); err != nil {
				t.Errorf("upload %s: %v", name, err)
			}
		}(name)
	}
	wg.Wait()
	if calls != len(names) {
		t.Errorf("listener called %d times, want %d", calls, len(names))
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
