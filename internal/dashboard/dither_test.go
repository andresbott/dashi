package dashboard

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDitherModes(t *testing.T) {
	got := DitherModes()
	want := []string{"auto", "balanced", "dynamic", "vivid", "soft", "grayscale"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("DitherModes() = %v, want %v", got, want)
	}
}

func TestStore_DitherValidation(t *testing.T) {
	valid := append([]string{""}, DitherModes()...)
	invalid := []string{"bogus", "Auto", "GRAYSCALE", " vivid"}

	for _, mode := range valid {
		t.Run("create accepts "+mode, func(t *testing.T) {
			s := NewStore(t.TempDir())
			d := Dashboard{Name: "A", Type: "image", Dither: mode}
			created, err := s.Create(d)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			got, err := s.Get(created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.Dither != mode {
				t.Errorf("stored dither = %q, want %q", got.Dither, mode)
			}
		})
	}

	for _, mode := range invalid {
		t.Run("rejects "+mode, func(t *testing.T) {
			s := NewStore(t.TempDir())
			if _, err := s.Create(Dashboard{Name: "A", Dither: mode}); !errors.Is(err, ErrInvalidDither) {
				t.Errorf("create err = %v, want ErrInvalidDither", err)
			}

			created, err := s.Create(Dashboard{Name: "B"})
			if err != nil {
				t.Fatal(err)
			}
			bad := created
			bad.Dither = mode
			if _, err := s.Update(bad); !errors.Is(err, ErrInvalidDither) {
				t.Errorf("update err = %v, want ErrInvalidDither", err)
			}

			data, err := json.Marshal(Dashboard{Name: "C", Dither: mode})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.ImportZip(makeZip(t, map[string][]byte{"dashboard.json": data})); !errors.Is(err, ErrInvalidDither) {
				t.Errorf("import err = %v, want ErrInvalidDither", err)
			}
		})
	}
}

// A dashboard without the field writes no "dither" key, so existing files
// stay byte-identical on their next save.
func TestDitherOmittedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	created, err := s.Create(Dashboard{Name: "Plain"})
	if err != nil {
		t.Fatal(err)
	}
	folder, err := s.DashDir(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(folder, dashboardFile))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "dither") {
		t.Errorf("dashboard.json mentions dither: %s", raw)
	}
}
