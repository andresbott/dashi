package handlers

import (
	"encoding/json"
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
