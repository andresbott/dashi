package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthTestHandler(t *testing.T) (*DashboardHandler, *dashboard.Store) {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)
	themeStore := themes.NewStore("")
	h := NewDashboardHandler(store, themeStore, slog.Default())
	return h, store
}

func TestDashboardHandler_GetAuth_Unprotected(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, _ := store.Create(dashboard.Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v0/dashboards/"+created.ID+"/auth", nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	rec := httptest.NewRecorder()

	h.GetAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["enabled"] != false {
		t.Fatalf("expected enabled=false, got %v", resp["enabled"])
	}
}

func TestDashboardHandler_SetAuth(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, _ := store.Create(dashboard.Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})

	body := `{"username":"viewer","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v0/dashboards/"+created.ID+"/auth", strings.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	rec := httptest.NewRecorder()

	h.SetAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", resp["enabled"])
	}
	if resp["username"] != "viewer" {
		t.Fatalf("expected username 'viewer', got %v", resp["username"])
	}
	if _, ok := resp["password"]; ok {
		t.Fatal("password hash should not be in response")
	}

	// Verify bcrypt hash was stored
	auth, _ := store.GetAuth(created.ID)
	if auth == nil {
		t.Fatal("expected auth to be set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte("secret123")); err != nil {
		t.Fatalf("stored password does not match: %v", err)
	}
}

func TestDashboardHandler_SetAuth_EmptyFields(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, _ := store.Create(dashboard.Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})

	tests := []struct {
		name string
		body string
	}{
		{"empty username", `{"username":"","password":"secret"}`},
		{"empty password", `{"username":"viewer","password":""}`},
		{"both empty", `{"username":"","password":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/v0/dashboards/"+created.ID+"/auth", strings.NewReader(tt.body))
			req = mux.SetURLVars(req, map[string]string{"id": created.ID})
			rec := httptest.NewRecorder()
			h.SetAuth(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestDashboardHandler_GetAuth_Protected(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, _ := store.Create(dashboard.Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})
	_ = store.SetAuth(created.ID, "viewer", "$2a$10$fakehashvalue")

	req := httptest.NewRequest(http.MethodGet, "/api/v0/dashboards/"+created.ID+"/auth", nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	rec := httptest.NewRecorder()

	h.GetAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", resp["enabled"])
	}
	if resp["username"] != "viewer" {
		t.Fatalf("expected username 'viewer', got %v", resp["username"])
	}
	if _, ok := resp["password"]; ok {
		t.Fatal("password hash should not be in response")
	}
}

func TestDashboardHandler_DeleteAuth(t *testing.T) {
	h, store := setupAuthTestHandler(t)
	created, _ := store.Create(dashboard.Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []dashboard.Page{}})
	_ = store.SetAuth(created.ID, "viewer", "$2a$10$fakehash")

	req := httptest.NewRequest(http.MethodDelete, "/api/v0/dashboards/"+created.ID+"/auth", nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	rec := httptest.NewRecorder()

	h.DeleteAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["enabled"] != false {
		t.Fatalf("expected enabled=false, got %v", resp["enabled"])
	}

	auth, _ := store.GetAuth(created.ID)
	if auth != nil {
		t.Fatal("expected auth to be nil after delete")
	}
}

func TestDashboardHandler_GetAuth_InvalidID(t *testing.T) {
	h, _ := setupAuthTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/dashboards/INVALID/auth", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "INVALID"})
	rec := httptest.NewRecorder()

	h.GetAuth(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
