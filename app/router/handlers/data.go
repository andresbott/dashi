package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/data"
	"github.com/andresbott/dashi/internal/data/backgrounds"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/gorilla/mux"
)

const maxDataUpload = 10 << 20 // 10 MB, matches UploadAsset in dashboards.go

// DataHandler exposes /api/v0/data/{notes,images,backgrounds} for the
// shared cross-dashboard user-data layer.
type DataHandler struct {
	notes       *notes.Store
	images      *images.Store
	backgrounds *backgrounds.Store
	logger      *slog.Logger
}

func NewDataHandler(n *notes.Store, i *images.Store, b *backgrounds.Store, logger *slog.Logger) *DataHandler {
	return &DataHandler{notes: n, images: i, backgrounds: b, logger: logger}
}

// RegisterRead mounts the GET endpoints on r. Intended for
// attachReadAPIs (viewer + editor).
func (h *DataHandler) RegisterRead(r *mux.Router) {
	r.Path("/data/notes").Methods(http.MethodGet).HandlerFunc(h.listNotes)
	r.Path("/data/notes/{name}").Methods(http.MethodGet).HandlerFunc(h.getNoteHTML)
	r.Path("/data/notes/{name}/raw").Methods(http.MethodGet).HandlerFunc(h.getNoteRaw)

	r.Path("/data/images").Methods(http.MethodGet).HandlerFunc(h.listImages)
	r.Path("/data/images/{name}").Methods(http.MethodGet).HandlerFunc(h.getImage)

	r.Path("/data/backgrounds").Methods(http.MethodGet).HandlerFunc(h.listBackgrounds)
	r.Path("/data/backgrounds/{name}").Methods(http.MethodGet).HandlerFunc(h.getBackground)
}

// RegisterWrite mounts the POST/DELETE endpoints on r. Intended for
// attachWriteAPIs (editor only).
func (h *DataHandler) RegisterWrite(r *mux.Router) {
	r.Path("/data/notes/{name}").Methods(http.MethodPost).HandlerFunc(h.saveNote)
	r.Path("/data/notes/{name}").Methods(http.MethodDelete).HandlerFunc(h.deleteNote)

	r.Path("/data/images/{name}").Methods(http.MethodPost).HandlerFunc(h.saveImage)
	r.Path("/data/images/{name}").Methods(http.MethodDelete).HandlerFunc(h.deleteImage)

	r.Path("/data/backgrounds/{name}").Methods(http.MethodPost).HandlerFunc(h.saveBackground)
	r.Path("/data/backgrounds/{name}").Methods(http.MethodDelete).HandlerFunc(h.deleteBackground)
}

// ----- notes -----

func (h *DataHandler) listNotes(w http.ResponseWriter, r *http.Request) {
	items, err := h.notes.List()
	if err != nil {
		h.internalErr(w, "list notes", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DataHandler) getNoteHTML(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	html, err := h.notes.GetHTML(name)
	if err != nil {
		h.dataErr(w, "get note", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"html": string(html)})
}

func (h *DataHandler) getNoteRaw(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	raw, err := h.notes.Get(name)
	if err != nil {
		h.dataErr(w, "get note raw", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(raw)) //nolint:gosec // G705: served as text/plain, not HTML; no XSS risk
}

func (h *DataHandler) saveNote(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	r.Body = http.MaxBytesReader(w, r.Body, maxDataUpload)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := h.notes.Save(name, string(body)); err != nil {
		h.dataErr(w, "save note", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *DataHandler) deleteNote(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.notes.Delete(name); err != nil {
		h.dataErr(w, "delete note", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- images -----

func (h *DataHandler) listImages(w http.ResponseWriter, r *http.Request) {
	items, err := h.images.List()
	if err != nil {
		h.internalErr(w, "list images", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DataHandler) getImage(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	body, mime, err := h.images.Get(name)
	if err != nil {
		h.dataErr(w, "get image", err)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(body) //nolint:gosec // G705: binary image data with explicit Content-Type and nosniff; not HTML
}

func (h *DataHandler) saveImage(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	r.Body = http.MaxBytesReader(w, r.Body, maxDataUpload)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := h.images.Save(name, body); err != nil {
		h.dataErr(w, "save image", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *DataHandler) deleteImage(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.images.Delete(name); err != nil {
		h.dataErr(w, "delete image", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- backgrounds -----

func (h *DataHandler) listBackgrounds(w http.ResponseWriter, r *http.Request) {
	items, err := h.backgrounds.List()
	if err != nil {
		h.internalErr(w, "list backgrounds", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DataHandler) getBackground(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	body, mime, err := h.backgrounds.Get(name)
	if err != nil {
		h.dataErr(w, "get background", err)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(body) //nolint:gosec // G705: binary image data with explicit Content-Type and nosniff; not HTML
}

func (h *DataHandler) saveBackground(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	r.Body = http.MaxBytesReader(w, r.Body, maxDataUpload)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := h.backgrounds.Save(name, body); err != nil {
		h.dataErr(w, "save background", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *DataHandler) deleteBackground(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.backgrounds.Delete(name); err != nil {
		h.dataErr(w, "delete background", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- helpers -----

// dataErr maps sentinel errors to HTTP status codes.
func (h *DataHandler) dataErr(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, data.ErrNotFound):
		ErrorJSON(w, "not found", http.StatusNotFound)
	case errors.Is(err, data.ErrInvalidName), errors.Is(err, data.ErrExtNotAllowed):
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
	default:
		h.internalErr(w, op, err)
	}
}

func (h *DataHandler) internalErr(w http.ResponseWriter, op string, err error) {
	if h.logger != nil {
		h.logger.Error(op, slog.String("error", err.Error()))
	}
	ErrorJSON(w, "internal server error", http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
