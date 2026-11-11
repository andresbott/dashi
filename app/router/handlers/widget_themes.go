package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andresbott/dashi/internal/widgetthemes"
	"github.com/gorilla/mux"
)

type WidgetThemeHandler struct {
	store  *widgetthemes.Store
	logger *slog.Logger
}

func NewWidgetThemeHandler(store *widgetthemes.Store, logger *slog.Logger) *WidgetThemeHandler {
	return &WidgetThemeHandler{store: store, logger: logger}
}

func (h *WidgetThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.store.List())
}

func (h *WidgetThemeHandler) GetIcon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	iconName := vars["icon"]

	if strings.ContainsAny(iconName, "/\\") || strings.Contains(iconName, "..") {
		ErrorJSON(w, "invalid icon name", http.StatusBadRequest)
		return
	}

	resolved, err := h.store.ResolveIcon(themeName, iconName)
	if err != nil {
		h.logger.Error("resolve icon", slog.String("error", err.Error()))
		ErrorJSON(w, err.Error(), http.StatusNotFound)
		return
	}

	switch resolved.Type {
	case widgetthemes.ThemeTypeFont:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"class": resolved.CSSClass})
	case widgetthemes.ThemeTypeImage:
		contentType := mime.TypeByExtension(filepath.Ext(resolved.FilePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, resolved.FilePath)
	}
}

func (h *WidgetThemeHandler) GetFont(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	fontName := vars["font"]

	// FontData resolves display fonts and, under the icon-font-{theme}
	// family name, the theme's icon font — the browser stack needs both as
	// @font-face sources (see widgetthemes.Store.FontFacesCSS).
	data, err := h.store.FontData(themeName, fontName)
	if err != nil {
		h.logger.Error("get font data", slog.String("error", err.Error()))
		ErrorJSON(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "font/ttf")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	// Font bytes served with an explicit Content-Type and nosniff; not HTML.
	if _, err := w.Write(data); err != nil {
		// Error already committed to response, log only
		return
	}
}

// ---------- admin CRUD ----------

// Upload extracts a zipped theme into the on-disk themes directory.
// The archive must contain a valid widget_theme.yaml at its root with a
// `name` field.
func (h *WidgetThemeHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, widgetthemes.MaxUploadSize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	info, err := h.store.Upload(body)
	if err != nil {
		h.themeWriteErr(w, "upload theme", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(info)
}

// Delete removes an uploaded theme. Builtin themes cannot be deleted.
func (h *WidgetThemeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.Delete(name); err != nil {
		h.themeWriteErr(w, "delete theme", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Download re-packages an uploaded theme's directory into a zip for
// download. Builtin themes return 403.
func (h *WidgetThemeHandler) Download(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	data, err := h.store.Zip(name)
	if err != nil {
		h.themeWriteErr(w, "zip theme", err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.zip"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Zip bytes served with an explicit Content-Type and nosniff; not HTML.
	_, _ = w.Write(data)
}

func (h *WidgetThemeHandler) themeWriteErr(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, widgetthemes.ErrNotFound):
		ErrorJSON(w, "not found", http.StatusNotFound)
	case errors.Is(err, widgetthemes.ErrConflict):
		ErrorJSON(w, err.Error(), http.StatusConflict)
	case errors.Is(err, widgetthemes.ErrBuiltin):
		ErrorJSON(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, widgetthemes.ErrInvalidName),
		errors.Is(err, widgetthemes.ErrInvalidArchive):
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	default:
		if h.logger != nil {
			h.logger.Error(op, slog.String("error", err.Error()))
		}
		ErrorJSON(w, "internal server error", http.StatusInternalServerError)
	}
}
