package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
)

type MarkdownHandler struct {
	store  *dashboard.Store
	logger *slog.Logger
	md     goldmark.Markdown
}

func NewMarkdownHandler(store *dashboard.Store, logger *slog.Logger) *MarkdownHandler {
	return &MarkdownHandler{
		store:  store,
		logger: logger,
		md:     goldmark.New(),
	}
}

func (h *MarkdownHandler) GetMarkdown(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	filename := mux.Vars(r)["filename"]

	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		ErrorJSON(w, "invalid filename", http.StatusBadRequest)
		return
	}

	data, _, err := h.store.GetAsset(id, "md/"+filename)
	if err != nil {
		ErrorJSON(w, "not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := h.md.Convert(data, &buf); err != nil {
		h.logger.Error("markdown render", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to render markdown", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"html": buf.String()})
}

// ListMarkdown lists the .md files in a dashboard's md/ folder.
// Returns 200 with {"files": [...]} on success, 404 on any store error.
func (h *MarkdownHandler) ListMarkdown(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	assets, err := h.store.ListAssets(id)
	if err != nil {
		ErrorJSON(w, "not found", http.StatusNotFound)
		return
	}

	const prefix = "md/"
	files := make([]string, 0, len(assets))
	for _, a := range assets {
		if !strings.HasPrefix(a, prefix) {
			continue
		}
		if !strings.HasSuffix(a, ".md") {
			continue
		}
		files = append(files, a[len(prefix):])
	}
	sort.Strings(files)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"files": files})
}
