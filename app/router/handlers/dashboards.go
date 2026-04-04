package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/gorilla/mux"
)

type DashboardHandler struct {
	store  *dashboard.Store
	logger *slog.Logger
}

func NewDashboardHandler(store *dashboard.Store, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{store: store, logger: logger}
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.List()
	if err != nil {
		h.logger.Error("list dashboards", slog.String("error", err.Error()))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": list})
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	d, err := h.store.Get(id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d dashboard.Dashboard
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(d)
	if err != nil {
		h.logger.Error("create dashboard", slog.String("error", err.Error()))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var d dashboard.Dashboard
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	d.ID = id

	updated, err := h.store.Update(d)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.Delete(id); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DashboardHandler) DeletePreviews(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.DeletePreviews()
	if err != nil {
		h.logger.Error("delete previews", slog.String("error", err.Error()))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"deleted": count})
}

func (h *DashboardHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	assets, err := h.store.ListAssets(id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": assets})
}

func (h *DashboardHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	assetPath := mux.Vars(r)["path"]
	data, mimeType, err := h.store.GetAsset(id, assetPath)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.Write(data)
}

func (h *DashboardHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	assetPath := mux.Vars(r)["path"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.SaveAsset(id, assetPath, data); err != nil {
		h.logger.Error("save asset", slog.String("error", err.Error()))
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *DashboardHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	assetPath := mux.Vars(r)["path"]
	if err := h.store.DeleteAsset(id, assetPath); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
