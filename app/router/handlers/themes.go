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

	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

type ThemeHandler struct {
	store  *themes.Store
	logger *slog.Logger
}

func NewThemeHandler(store *themes.Store, logger *slog.Logger) *ThemeHandler {
	return &ThemeHandler{store: store, logger: logger}
}

func (h *ThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.store.List())
}

func (h *ThemeHandler) GetIcon(w http.ResponseWriter, r *http.Request) {
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
	case themes.ThemeTypeFont:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"class": resolved.CSSClass})
	case themes.ThemeTypeImage:
		contentType := mime.TypeByExtension(filepath.Ext(resolved.FilePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, resolved.FilePath)
	}
}

func (h *ThemeHandler) GetFont(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	fontName := vars["font"]

	data, err := h.store.GetDisplayFontData(themeName, fontName)
	if err != nil {
		h.logger.Error("get font data", slog.String("error", err.Error()))
		ErrorJSON(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "font/ttf")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write(data); err != nil { //nolint:gosec // G705: font bytes served with explicit Content-Type and nosniff; not HTML
		// Error already committed to response, log only
		return
	}
}

func (h *ThemeHandler) GetBackground(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	fileName := vars["file"]

	if strings.ContainsAny(fileName, "/\\") || strings.Contains(fileName, "..") {
		ErrorJSON(w, "invalid filename", http.StatusBadRequest)
		return
	}

	data, err := h.store.GetBackgroundData(themeName, fileName)
	if err != nil {
		ErrorJSON(w, "not found", http.StatusNotFound)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write(data); err != nil { //nolint:gosec // G705: background image bytes served with explicit Content-Type and nosniff; not HTML
		// Error already committed to response, log only
		return
	}
}

// ---------- admin CRUD ----------

// Upload extracts a zipped theme into the on-disk themes directory.
// The archive must contain a valid theme.yaml at its root with `name`
// and `type` (icon|style) fields.
func (h *ThemeHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, themes.MaxUploadSize)
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
func (h *ThemeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.Delete(name); err != nil {
		h.themeWriteErr(w, "delete theme", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Download re-packages an uploaded theme's directory into a zip for
// download. Builtin themes return 403.
func (h *ThemeHandler) Download(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	data, err := h.store.Zip(name)
	if err != nil {
		h.themeWriteErr(w, "zip theme", err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.zip"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data) //nolint:gosec // G705: zip bytes served with explicit Content-Type and nosniff; not HTML
}

func (h *ThemeHandler) themeWriteErr(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, themes.ErrNotFound):
		ErrorJSON(w, "not found", http.StatusNotFound)
	case errors.Is(err, themes.ErrConflict):
		ErrorJSON(w, err.Error(), http.StatusConflict)
	case errors.Is(err, themes.ErrBuiltin):
		ErrorJSON(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, themes.ErrInvalidName),
		errors.Is(err, themes.ErrInvalidArchive),
		errors.Is(err, themes.ErrInvalidType):
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	default:
		if h.logger != nil {
			h.logger.Error(op, slog.String("error", err.Error()))
		}
		ErrorJSON(w, "internal server error", http.StatusInternalServerError)
	}
}
