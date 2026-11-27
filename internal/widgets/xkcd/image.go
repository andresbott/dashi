package xkcd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	xkcdclient "github.com/andresbott/dashi/internal/providers/xkcd"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed image.html
var staticHTML string

var tmpl = template.Must(template.New("xkcd").Parse(staticHTML))

type xkcdConfig struct {
	Mode string `json:"mode"`
}

type xkcdTemplateData struct {
	Num   int
	Title string
	Img   string
	Alt   string
}

// loadComic decodes the config and fetches the comic its mode selects. Both
// render stacks use it, so they always show the same comic.
func loadComic(client *xkcdclient.Client, config json.RawMessage) (xkcdTemplateData, error) {
	var cfg xkcdConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return xkcdTemplateData{}, fmt.Errorf("xkcd config: %w", err)
		}
	}

	var comic xkcdclient.Comic
	var err error

	switch cfg.Mode {
	case "random":
		comic, err = client.GetDailyRandom()
	case "random-each":
		comic, err = client.GetRandom()
	default:
		comic, err = client.GetLatest()
	}
	if err != nil {
		return xkcdTemplateData{}, fmt.Errorf("xkcd fetch: %w", err)
	}

	return xkcdTemplateData{
		Num:   comic.Num,
		Title: comic.SafeTitle,
		Img:   comic.Img,
		Alt:   comic.Alt,
	}, nil
}

func NewStaticRenderer(client *xkcdclient.Client) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		data, err := loadComic(client, config)
		if err != nil {
			return "", err
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("xkcd template: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
