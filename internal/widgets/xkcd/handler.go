package xkcd

import (
	"encoding/json"
	"log/slog"
	"net/http"

	xkcdclient "github.com/andresbott/dashi/internal/xkcd"
)

type handler struct {
	client *xkcdclient.Client
	logger *slog.Logger
}

func newHandler(client *xkcdclient.Client, logger *slog.Logger) *handler {
	return &handler{client: client, logger: logger}
}

func (h *handler) GetComic(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")

	var comic xkcdclient.Comic
	var err error

	switch mode {
	case "random":
		comic, err = h.client.GetDailyRandom()
	case "random-each":
		comic, err = h.client.GetRandom()
	default:
		comic, err = h.client.GetLatest()
	}

	if err != nil {
		if h.logger != nil {
			h.logger.Error("fetch xkcd comic", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to fetch xkcd comic", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comic)
}

// writeJSONError mirrors handlers.ErrorJSON but lives in-package so
// the widget does not import app/router/handlers.
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
