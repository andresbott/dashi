package handlers

import (
	"encoding/json"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

type ThemeHandler struct {
	store *themes.Store
}

func NewThemeHandler(store *themes.Store) *ThemeHandler {
	return &ThemeHandler{store: store}
}

func (h *ThemeHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.store.List())
}

func (h *ThemeHandler) GetIcon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	iconName := vars["icon"]

	if strings.ContainsAny(iconName, "/\\") || strings.Contains(iconName, "..") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid icon name"})
		return
	}

	resolved, err := h.store.ResolveIcon(themeName, iconName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	switch resolved.Type {
	case themes.ThemeTypeFont:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"class": resolved.CSSClass})
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "font/ttf")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func (h *ThemeHandler) GetBackground(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	themeName := vars["name"]
	fileName := vars["file"]

	if strings.ContainsAny(fileName, "/\\") || strings.Contains(fileName, "..") {
		http.Error(w, `{"error":"invalid filename"}`, http.StatusBadRequest)
		return
	}

	data, err := h.store.GetBackgroundData(themeName, fileName)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}
