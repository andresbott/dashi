package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/themes"
)

func TestListBackgrounds_IncludesShared(t *testing.T) {
	dir := t.TempDir()
	dashStore := dashboard.NewStore(dir)
	created, err := dashStore.Create(dashboard.Dashboard{
		ID:   "test01",
		Name: "Test",
		Icon: "ti-home",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Pages: []dashboard.Page{},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	bs, err := images.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := bs.Save("sunset.jpg", []byte{0xFF, 0xD8, 0xFF}); err != nil {
		t.Fatal(err)
	}

	ts := themes.NewStore("")

	h := NewDashboardHandler(dashStore, ts, bs, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/backgrounds?dashboard="+created.ID, nil)
	w := httptest.NewRecorder()
	h.ListBackgrounds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Shared []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"shared"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Shared) != 1 || body.Shared[0].Value != "shared:sunset.jpg" {
		t.Fatalf("unexpected shared: %+v", body.Shared)
	}
}
