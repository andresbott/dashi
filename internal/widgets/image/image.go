package image

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

var validFits = map[string]bool{"cover": true, "contain": true, "fill": true}

//go:embed image.html
var imageHTML string

var tmpl = template.Must(template.New("image").Parse(imageHTML))

type imageConfig struct {
	Image string `json:"image"`
	Fit   string `json:"fit"`
}

type imageData struct {
	Mime string
	Data string
	Fit  string
}

func NewStaticRenderer(store *dashboard.Store) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg imageConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("image config: %w", err)
			}
		}

		if cfg.Image == "" {
			return template.HTML(`<div class="widget-image-empty"></div>`), nil
		}

		data, mimeType, err := store.GetAsset(ctx.DashboardID, cfg.Image)
		if err != nil {
			return template.HTML(`<div class="widget-image-empty"></div>`), nil
		}

		fit := cfg.Fit
		if !validFits[fit] {
			fit = "cover"
		}

		if !strings.HasPrefix(mimeType, "image/") {
			mimeType = "image/png"
		}

		d := imageData{
			Mime: mimeType,
			Data: base64.StdEncoding.EncodeToString(data),
			Fit:  fit,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, d); err != nil {
			return "", fmt.Errorf("image render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
