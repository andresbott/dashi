package sysinfo

import (
	"encoding/json"
	"log/slog"
	"net/http"

	sysinfopkg "github.com/andresbott/dashi/internal/providers/sysinfo"
)

type handler struct {
	logger *slog.Logger
}

func newHandler(logger *slog.Logger) *handler {
	return &handler{logger: logger}
}

func (h *handler) GetSysinfo(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfopkg.Get()
	if err != nil {
		if h.logger != nil {
			h.logger.Error("fetch sysinfo", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to fetch system info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
