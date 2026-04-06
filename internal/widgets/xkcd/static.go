package xkcd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	xkcdclient "github.com/andresbott/dashi/internal/xkcd"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed static.html
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

func NewStaticRenderer(client *xkcdclient.Client) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg xkcdConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("xkcd config: %w", err)
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
			return "", fmt.Errorf("xkcd fetch: %w", err)
		}

		data := xkcdTemplateData{
			Num:   comic.Num,
			Title: comic.SafeTitle,
			Img:   comic.Img,
			Alt:   comic.Alt,
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("xkcd template: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
