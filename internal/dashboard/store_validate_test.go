package dashboard

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// rejectValidator rejects any widget of type "strict" whose config is "bad".
type rejectValidator struct{}

func (rejectValidator) ValidateConfig(widgetType string, config json.RawMessage) error {
	if widgetType == "strict" && string(config) == `"bad"` {
		return errors.New("strict says no")
	}
	return nil
}

func dashWithWidget(name, widgetType, config string) Dashboard {
	return Dashboard{
		Name: name,
		Pages: []Page{{
			Name: "Main",
			Rows: []Row{{ID: "r1", Widgets: []Widget{
				{ID: "ok1", Type: "clock", Config: json.RawMessage(`{}`)},
				{ID: "w1", Type: widgetType, Config: json.RawMessage(config)},
			}}},
		}},
	}
}

func TestStore_WidgetValidation(t *testing.T) {
	tests := []struct {
		name    string
		write   func(s *Store) error
		wantErr bool
	}{
		{
			name:  "create accepts a valid widget",
			write: func(s *Store) error { _, err := s.Create(dashWithWidget("A", "strict", `"ok"`)); return err },
		},
		{
			name:    "create rejects an invalid widget",
			write:   func(s *Store) error { _, err := s.Create(dashWithWidget("A", "strict", `"bad"`)); return err },
			wantErr: true,
		},
		{
			name: "update rejects an invalid widget",
			write: func(s *Store) error {
				created, err := s.Create(dashWithWidget("A", "strict", `"ok"`))
				if err != nil {
					return err
				}
				bad := dashWithWidget("Renamed", "strict", `"bad"`)
				bad.ID = created.ID
				_, err = s.Update(bad)
				return err
			},
			wantErr: true,
		},
		{
			name: "import rejects an invalid widget",
			write: func(s *Store) error {
				data, err := json.Marshal(dashWithWidget("A", "strict", `"bad"`))
				if err != nil {
					return err
				}
				_, err = s.ImportZip(makeZip(t, map[string][]byte{"dashboard.json": data}))
				return err
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore(t.TempDir())
			s.SetWidgetValidator(rejectValidator{})
			err := tc.write(s)
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, ErrInvalidWidgetConfig) {
				t.Fatalf("err = %v, want ErrInvalidWidgetConfig", err)
			}
			// The message names the widget, its page and the rule's reason.
			for _, want := range []string{`strict widget "w1"`, `page "Main"`, "strict says no"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err, want)
				}
			}
		})
	}
}

// A rejected write leaves the disk untouched: no new folder on create or
// import, and no rename or rewrite on update.
func TestStore_WidgetValidation_NothingWritten(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.SetWidgetValidator(rejectValidator{})

	created, err := s.Create(dashWithWidget("Original", "strict", `"ok"`))
	if err != nil {
		t.Fatal(err)
	}

	bad := dashWithWidget("Renamed", "strict", `"bad"`)
	bad.ID = created.ID
	if _, err := s.Update(bad); err == nil {
		t.Fatal("update: want error")
	}
	if _, err := s.Create(dashWithWidget("Other", "strict", `"bad"`)); err == nil {
		t.Fatal("create: want error")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "original" {
		t.Fatalf("dashboards dir = %v, want only the untouched original folder", entries)
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Original" {
		t.Errorf("stored name = %q, want the pre-update %q", got.Name, "Original")
	}
}

// Without a validator the store keeps its old pass-through behaviour.
func TestStore_NoWidgetValidator(t *testing.T) {
	s := NewStore(t.TempDir())
	if _, err := s.Create(dashWithWidget("A", "strict", `"bad"`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
