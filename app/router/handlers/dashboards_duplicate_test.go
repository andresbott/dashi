package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/gorilla/mux"
)

func TestDashboardHandler_Duplicate(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, err := store.Create(dashboard.Dashboard{ID: "dup01", Name: "Hall", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v0/dashboards/"+created.ID+"/duplicate", nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	rec := httptest.NewRecorder()

	h.Duplicate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp dashboard.Dashboard
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.ID == created.ID {
		t.Errorf("expected a new ID, got %q", resp.ID)
	}
	if resp.Name != "Hall (copy)" {
		t.Errorf("name = %q, want %q", resp.Name, "Hall (copy)")
	}
}

func TestDashboardHandler_Duplicate_NotFound(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "missing", id: "missing01"},
		{name: "invalid id", id: "INVALID"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := setupAuthTestHandler(t)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/dashboards/"+tc.id+"/duplicate", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})
			rec := httptest.NewRecorder()

			h.Duplicate(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
		})
	}
}
