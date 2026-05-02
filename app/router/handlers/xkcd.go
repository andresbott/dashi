package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/xkcd"
)

type XkcdHandler struct {
	client *xkcd.Client
	logger *slog.Logger
}

func NewXkcdHandler(client *xkcd.Client, logger *slog.Logger) *XkcdHandler {
	return &XkcdHandler{client: client, logger: logger}
}

func (h *XkcdHandler) GetComic(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")

	var comic xkcd.Comic
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
		h.logger.Error("fetch xkcd comic", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to fetch xkcd comic", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comic)
}
