package router

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	"golang.org/x/crypto/bcrypt"
)

// NewDashboardAuthMiddleware returns middleware that checks basic auth
// for protected dashboards in the viewer. Unprotected dashboards pass through.
// It checks both page-load paths (/{id}) and API paths (/api/v0/dashboards/{id}).
func NewDashboardAuthMiddleware(store *dashboard.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			id, dashName := extractDashboardID(r.URL.Path, store)
			if id == "" {
				next.ServeHTTP(w, r)
				return
			}

			auth, err := store.GetAuth(id)
			if err != nil || auth == nil {
				// Unprotected or error reading auth — pass through
				next.ServeHTTP(w, r)
				return
			}

			// Dashboard is protected — check credentials
			username, password, ok := r.BasicAuth()
			if !ok || !checkCredentials(username, password, auth) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Dashboard: `+dashName+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractDashboardID extracts a dashboard ID from the request path.
// Returns the ID and dashboard name, or empty strings if no dashboard matches.
func extractDashboardID(path string, store *dashboard.Store) (id string, name string) {
	trimmed := strings.TrimPrefix(path, "/")

	// Check direct dashboard path: /{id}
	if !strings.Contains(trimmed, "/") && trimmed != "" {
		if dash, err := store.Get(trimmed); err == nil {
			return dash.ID, dash.Name
		}
		return "", ""
	}

	// Check API path: /api/v0/dashboards/{id} or /api/v0/dashboards/{id}/...
	const prefix = "api/v0/dashboards/"
	if strings.HasPrefix(trimmed, prefix) {
		rest := trimmed[len(prefix):]
		id := rest
		if idx := strings.Index(rest, "/"); idx >= 0 {
			id = rest[:idx]
		}
		if id != "" {
			if dash, err := store.Get(id); err == nil {
				return dash.ID, dash.Name
			}
		}
	}

	return "", ""
}

// checkCredentials compares provided credentials against stored auth.
func checkCredentials(username, password string, auth *dashboard.Auth) bool {
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(auth.Username)) == 1
	passwordMatch := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password)) == nil
	return usernameMatch && passwordMatch
}
