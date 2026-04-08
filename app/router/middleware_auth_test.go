package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthMiddleware(t *testing.T) (*dashboard.Store, func(http.Handler) http.Handler) {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)
	mid := NewDashboardAuthMiddleware(store)
	return store, mid
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(h)
}

func TestAuthMiddleware_UnprotectedDashboard(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "Test", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/"+created.ID, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "OK" {
		t.Fatalf("expected 'OK', got %q", rec.Body.String())
	}
}

func TestAuthMiddleware_ProtectedDashboard_NoCredentials(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "My Dashboard", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	_ = store.SetAuth(created.ID, "viewer", hashPassword(t, "secret"))

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/"+created.ID, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	wwwAuth := rec.Header().Get("WWW-Authenticate")
	if wwwAuth == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
	if wwwAuth != `Basic realm="Dashboard: My Dashboard"` {
		t.Fatalf("unexpected realm: %q", wwwAuth)
	}
}

func TestAuthMiddleware_ProtectedDashboard_CorrectCredentials(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "Test", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	_ = store.SetAuth(created.ID, "viewer", hashPassword(t, "secret"))

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/"+created.ID, nil)
	req.SetBasicAuth("viewer", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ProtectedDashboard_WrongPassword(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "Test", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	_ = store.SetAuth(created.ID, "viewer", hashPassword(t, "secret"))

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/"+created.ID, nil)
	req.SetBasicAuth("viewer", "wrongpassword")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ProtectedDashboard_WrongUsername(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "Test", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	_ = store.SetAuth(created.ID, "viewer", hashPassword(t, "secret"))

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/"+created.ID, nil)
	req.SetBasicAuth("wronguser", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_APIPath_Protected(t *testing.T) {
	store, mid := setupAuthMiddleware(t)
	created, _ := store.Create(dashboard.Dashboard{
		ID: "abc123", Name: "Test", Type: "interactive",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	_ = store.SetAuth(created.ID, "viewer", hashPassword(t, "secret"))

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Without credentials
	req := httptest.NewRequest("GET", "/api/v0/dashboards/"+created.ID, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("API without creds: expected 401, got %d", rec.Code)
	}

	// With correct credentials
	req = httptest.NewRequest("GET", "/api/v0/dashboards/"+created.ID, nil)
	req.SetBasicAuth("viewer", "secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("API with creds: expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_NonDashboardPath(t *testing.T) {
	_, mid := setupAuthMiddleware(t)

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))

	paths := []string{"/", "/assets/app.js", "/api/v0/health", "/api/v0/widgets/weather"}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200, got %d", p, rec.Code)
		}
	}
}
