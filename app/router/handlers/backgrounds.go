package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/backgrounds"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/gorilla/mux"
)

const maxBackgroundAssetUpload = 10 << 20 // 10 MB, same cap as other uploads

// BackgroundHandler serves the background entity and its assets. It holds
// the dashboard store only to count references for the listing — the shared
// instance from newSharedDeps, never a second one.
type BackgroundHandler struct {
	store     *backgrounds.Store
	dashStore *dashboard.Store
	logger    *slog.Logger
}

func NewBackgroundHandler(store *backgrounds.Store, dashStore *dashboard.Store, logger *slog.Logger) *BackgroundHandler {
	return &BackgroundHandler{store: store, dashStore: dashStore, logger: logger}
}

func (h *BackgroundHandler) List(w http.ResponseWriter, r *http.Request) {
	usage, err := h.usageCounts()
	if err != nil {
		h.fail(w, "count background usage", err)
		return
	}
	items, err := h.store.List(usage)
	if err != nil {
		h.fail(w, "list backgrounds", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// usageCounts maps background ID -> number of dashboards referencing it.
func (h *BackgroundHandler) usageCounts() (map[string]int, error) {
	metas, err := h.dashStore.List()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(metas))
	for _, m := range metas {
		d, err := h.dashStore.Get(m.ID)
		if err != nil {
			continue // a broken dashboard must not break the background list
		}
		if d.BackgroundID != "" {
			counts[d.BackgroundID]++
		}
	}
	return counts, nil
}

func (h *BackgroundHandler) Get(w http.ResponseWriter, r *http.Request) {
	b, err := h.store.Get(mux.Vars(r)["id"])
	if err != nil {
		h.fail(w, "get background", err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *BackgroundHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b backgrounds.Background
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}
	b.ID = "" // the store assigns it
	created, err := h.store.Create(b)
	if err != nil {
		h.fail(w, "create background", err)
		return
	}
	// 200, not 201: a middleware rewrites 201 bodies into
	// {"error":<body>,"code":201}, which would corrupt the created object.
	// DashboardHandler.Create returns 200 with the object for the same reason.
	writeJSON(w, http.StatusOK, created)
}

func (h *BackgroundHandler) Update(w http.ResponseWriter, r *http.Request) {
	var b backgrounds.Background
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}
	b.ID = mux.Vars(r)["id"]
	updated, err := h.store.Update(b)
	if err != nil {
		h.fail(w, "update background", err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes a background even when dashboards reference it. A dangling
// reference resolves to nothing and the dashboard falls back to the theme
// background; refusing the delete would strand the entity instead.
func (h *BackgroundHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(mux.Vars(r)["id"]); err != nil {
		h.fail(w, "delete background", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BackgroundHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.ListAssets(mux.Vars(r)["id"])
	if err != nil {
		h.fail(w, "list background assets", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *BackgroundHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, mimeType, err := h.store.GetAsset(vars["id"], vars["path"])
	if err != nil {
		h.fail(w, "get background asset", err)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data) //nolint:gosec // G705: binary image data with explicit Content-Type and nosniff; not HTML
}

func (h *BackgroundHandler) SaveAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxBackgroundAssetUpload)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := h.store.SaveAsset(vars["id"], vars["path"], body); err != nil {
		h.fail(w, "save background asset", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *BackgroundHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.store.DeleteAsset(vars["id"], vars["path"]); err != nil {
		h.fail(w, "delete background asset", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// fail maps store errors onto status codes.
func (h *BackgroundHandler) fail(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, backgrounds.ErrNotFound), errors.Is(err, backgrounds.ErrInvalidID):
		ErrorJSON(w, "not found", http.StatusNotFound)
	case errors.Is(err, backgrounds.ErrInvalidConfig):
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	default:
		if h.logger != nil {
			h.logger.Error(op, slog.String("error", err.Error()))
		}
		ErrorJSON(w, "internal server error", http.StatusInternalServerError)
	}
}
