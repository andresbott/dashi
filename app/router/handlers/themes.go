package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

const maxThemeAssetUpload = 10 << 20 // 10 MB, same cap as other uploads

// ThemeHandler serves the theme entity and its assets. It holds the
// dashboard store only to count references for the listing — the shared
// instance from newSharedDeps, never a second one.
type ThemeHandler struct {
	store     *themes.Store
	dashStore *dashboard.Store
	logger    *slog.Logger
}

func NewThemeHandler(store *themes.Store, dashStore *dashboard.Store, logger *slog.Logger) *ThemeHandler {
	return &ThemeHandler{store: store, dashStore: dashStore, logger: logger}
}

func (h *ThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	usage, err := h.usageCounts()
	if err != nil {
		h.fail(w, "count theme usage", err)
		return
	}
	items, err := h.store.List(usage)
	if err != nil {
		h.fail(w, "list themes", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// usageCounts maps theme ID -> number of dashboards rendering with it.
// Dashboards without a theme render the builtin, so they count toward
// themes.DefaultID. A dangling ID only ever lands on a key no listed theme
// reads, so it counts toward nothing.
func (h *ThemeHandler) usageCounts() (map[string]int, error) {
	metas, err := h.dashStore.List()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(metas))
	for _, m := range metas {
		d, err := h.dashStore.Get(m.ID)
		if err != nil {
			continue // a broken dashboard must not break the theme list
		}
		id := d.ThemeID
		if id == "" {
			id = themes.DefaultID
		}
		counts[id]++
	}
	return counts, nil
}

// themeResponse is a theme plus the read-only flag the editor needs. The flag
// lives on the response, not on themes.Theme, so it is never persisted.
type themeResponse struct {
	themes.Theme
	Builtin bool `json:"builtin,omitempty"`
}

func (h *ThemeHandler) Get(w http.ResponseWriter, r *http.Request) {
	t, err := h.store.Get(mux.Vars(r)["id"])
	if err != nil {
		h.fail(w, "get theme", err)
		return
	}
	writeJSON(w, http.StatusOK, themeResponse{Theme: t, Builtin: themes.IsBuiltin(t.ID)})
}

func (h *ThemeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var t themes.Theme
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}
	t.ID = "" // the store assigns it
	created, err := h.store.Create(t)
	if err != nil {
		h.fail(w, "create theme", err)
		return
	}
	// 200, not 201: a middleware rewrites 201 bodies into
	// {"error":<body>,"code":201}, which would corrupt the created object.
	// DashboardHandler.Create returns 200 with the object for the same reason.
	writeJSON(w, http.StatusOK, created)
}

func (h *ThemeHandler) Update(w http.ResponseWriter, r *http.Request) {
	var t themes.Theme
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}
	t.ID = mux.Vars(r)["id"]
	updated, err := h.store.Update(t)
	if err != nil {
		h.fail(w, "update theme", err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes a theme even when dashboards reference it. A dangling
// reference resolves to the builtin default; refusing the delete would
// strand the entity instead.
func (h *ThemeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(mux.Vars(r)["id"]); err != nil {
		h.fail(w, "delete theme", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Copy duplicates a theme, assets included. It takes no body. Like Create
// it answers 200, not 201, for the reason given there.
func (h *ThemeHandler) Copy(w http.ResponseWriter, r *http.Request) {
	created, err := h.store.Copy(mux.Vars(r)["id"])
	if err != nil {
		h.fail(w, "copy theme", err)
		return
	}
	writeJSON(w, http.StatusOK, created)
}

func (h *ThemeHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.ListAssets(mux.Vars(r)["id"])
	if err != nil {
		h.fail(w, "list theme assets", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *ThemeHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, mimeType, err := h.store.GetAsset(vars["id"], vars["path"])
	if err != nil {
		h.fail(w, "get theme asset", err)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data) //nolint:gosec // G705: binary image data with explicit Content-Type and nosniff; not HTML
}

func (h *ThemeHandler) SaveAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxThemeAssetUpload)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := h.store.SaveAsset(vars["id"], vars["path"], body); err != nil {
		h.fail(w, "save theme asset", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *ThemeHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.store.DeleteAsset(vars["id"], vars["path"]); err != nil {
		h.fail(w, "delete theme asset", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// fail maps store errors onto status codes.
func (h *ThemeHandler) fail(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, themes.ErrNotFound):
		ErrorJSON(w, "not found", http.StatusNotFound)
	case errors.Is(err, themes.ErrInvalidID), errors.Is(err, themes.ErrInvalidConfig):
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, themes.ErrBuiltin):
		ErrorJSON(w, err.Error(), http.StatusForbidden)
	default:
		if h.logger != nil {
			h.logger.Error(op, slog.String("error", err.Error()))
		}
		ErrorJSON(w, "internal server error", http.StatusInternalServerError)
	}
}
