package markdown

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

type handler struct {
	store  *dashboard.Store
	logger *slog.Logger
	md     goldmark.Markdown
}

func newHandler(store *dashboard.Store, logger *slog.Logger) *handler {
	return &handler{
		store:  store,
		logger: logger,
		md:     goldmark.New(),
	}
}

func (h *handler) GetMarkdown(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	filename := mux.Vars(r)["filename"]

	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		writeJSONError(w, "invalid filename", http.StatusBadRequest)
		return
	}

	data, _, err := h.store.GetAsset(id, "md/"+filename)
	if err != nil {
		writeJSONError(w, "not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := h.md.Convert(data, &buf); err != nil {
		if h.logger != nil {
			h.logger.Error("markdown render", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to render markdown", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"html": buf.String()})
}

func (h *handler) ListMarkdown(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	assets, err := h.store.ListAssets(id)
	if err != nil {
		writeJSONError(w, "not found", http.StatusNotFound)
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

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
