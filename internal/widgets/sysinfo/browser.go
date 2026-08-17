package sysinfo

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/sysinfo"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed browser.html
var browserHTML string

//go:embed sysinfo.css
var browserCSS []byte

//go:embed sysinfo.js
var browserJS []byte

var browserTmpl = template.Must(template.New("sysinfo-browser").Parse(browserHTML))

// RenderBrowser renders the sysinfo skeleton with current values filled
// in. One slot per configured disk: the slot count comes from config, so
// sysinfo.js only ever writes into existing elements.
func (m *Module) RenderBrowser(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
	var cfg sysinfoConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return "", fmt.Errorf("sysinfo config: %w", err)
		}
	}

	// A failed local probe must not blank the page: render the skeleton
	// with empty values and let sysinfo.js fill it on the next tick.
	var data sysinfoData
	if info, err := sysinfo.Get(); err == nil {
		data = buildSysinfoData(cfg, info, diskRowsConfigured)
	} else {
		// Probe failed: render the skeleton with empty values rather than
		// humanised zeros, so the widget shows blanks until the first poll.
		data = sysinfoData{
			ShowMemory: cfg.ShowMemory,
			ShowUptime: cfg.ShowUptime,
			Disks:      buildDiskRows(cfg.Disks, nil, diskRowsConfigured),
		}
	}

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("sysinfo browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// CSS implements widgets.BrowserAssets.
func (m *Module) CSS() []byte { return browserCSS }

// JS implements widgets.BrowserAssets.
func (m *Module) JS() []byte { return browserJS }
