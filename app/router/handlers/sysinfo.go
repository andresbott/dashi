package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/sysinfo"
)

type SysinfoHandler struct {
	logger *slog.Logger
}

func NewSysinfoHandler(logger *slog.Logger) *SysinfoHandler {
	return &SysinfoHandler{logger: logger}
}

func (h *SysinfoHandler) GetSysinfo(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.Get()
	if err != nil {
		h.logger.Error("fetch sysinfo", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to fetch system info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		ErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
	}
}
