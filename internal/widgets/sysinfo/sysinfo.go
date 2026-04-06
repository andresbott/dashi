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

//go:embed sysinfo.html
var sysinfoHTML string

var tmpl = template.Must(template.New("sysinfo").Parse(sysinfoHTML))

type sysinfoConfig struct {
	ShowMemory bool     `json:"showMemory"`
	ShowUptime bool     `json:"showUptime"`
	Disks      []string `json:"disks"`
}

type diskData struct {
	Mountpoint string
	UsedPct    float64
	TotalHuman string
	FreeHuman  string
}

type sysinfoData struct {
	Disks         []diskData
	ShowMemory    bool
	MemUsedPct    float64
	MemTotalHuman string
	MemUsedHuman  string
	ShowUptime    bool
	UptimeHuman   string
}

func humanBytes(b uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)
	switch {
	case b >= tb:
		return fmt.Sprintf("%.1f TB", float64(b)/float64(tb))
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func humanUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		var cfg sysinfoConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("sysinfo config: %w", err)
			}
		}

		info, err := sysinfo.Get()
		if err != nil {
			return "", fmt.Errorf("sysinfo get: %w", err)
		}

		enabledDisks := make(map[string]bool)
		for _, mp := range cfg.Disks {
			enabledDisks[mp] = true
		}

		data := sysinfoData{
			ShowMemory: cfg.ShowMemory,
			ShowUptime: cfg.ShowUptime,
		}

		for _, d := range info.Disks {
			if !enabledDisks[d.Mountpoint] {
				continue
			}
			data.Disks = append(data.Disks, diskData{
				Mountpoint: d.Mountpoint,
				UsedPct:    d.UsedPct,
				TotalHuman: humanBytes(d.Total),
				FreeHuman:  humanBytes(d.Free),
			})
		}

		if cfg.ShowMemory {
			data.MemUsedPct = info.MemUsedPct
			data.MemTotalHuman = humanBytes(info.MemTotal)
			data.MemUsedHuman = humanBytes(info.MemUsed)
		}

		if cfg.ShowUptime {
			data.UptimeHuman = humanUptime(info.UptimeSeconds)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("sysinfo render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
