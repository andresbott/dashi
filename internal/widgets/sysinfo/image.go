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

//go:embed image.html
var imageHTML string

var imageTmpl = template.Must(template.New("sysinfo").Parse(imageHTML))

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

// diskRowMode selects how disk rows are derived from config and probe
// results. The two render stacks genuinely need different answers, which is
// why this is a parameter rather than one shared behaviour.
type diskRowMode int

const (
	// diskRowsMounted emits a row only for configured mountpoints the probe
	// actually reported, in probe order. This is the image stack's
	// behaviour: a PNG is rendered once and never updated, so a row for an
	// absent mountpoint would permanently show " free / " next to an empty
	// bar.
	diskRowsMounted diskRowMode = iota

	// diskRowsConfigured emits a row for every configured mountpoint in
	// config order, mounted or not. This is the browser stack's behaviour:
	// sysinfo.js writes into existing elements only (decision D3), so a
	// mountpoint that appears after page load needs a slot waiting for it.
	diskRowsConfigured
)

// buildSysinfoData assembles the viewmodel from config and probe results.
// Disk rows follow mode; memory, uptime and all humanised formatting are
// shared by both stacks.
func buildSysinfoData(cfg sysinfoConfig, info sysinfo.SystemInfo, mode diskRowMode) sysinfoData {
	data := sysinfoData{
		ShowMemory: cfg.ShowMemory,
		ShowUptime: cfg.ShowUptime,
		Disks:      buildDiskRows(cfg.Disks, info.Disks, mode),
	}

	if cfg.ShowMemory {
		data.MemUsedPct = info.MemUsedPct
		data.MemTotalHuman = humanBytes(info.MemTotal)
		data.MemUsedHuman = humanBytes(info.MemUsed)
	}

	if cfg.ShowUptime {
		data.UptimeHuman = humanUptime(info.UptimeSeconds)
	}

	return data
}

// buildDiskRows turns the configured mountpoints and the probe results into
// rows according to mode. See diskRowMode for why the two stacks differ.
func buildDiskRows(configured []string, probed []sysinfo.DiskInfo, mode diskRowMode) []diskData {
	enabled := make(map[string]bool, len(configured))
	for _, mp := range configured {
		enabled[mp] = true
	}

	if mode == diskRowsMounted {
		var rows []diskData
		for _, d := range probed {
			if !enabled[d.Mountpoint] {
				continue
			}
			rows = append(rows, diskRow(d))
		}
		return rows
	}

	byMountpoint := make(map[string]sysinfo.DiskInfo, len(probed))
	for _, d := range probed {
		byMountpoint[d.Mountpoint] = d
	}
	var rows []diskData
	for _, mp := range configured {
		if d, found := byMountpoint[mp]; found {
			rows = append(rows, diskRow(d))
			continue
		}
		rows = append(rows, diskData{Mountpoint: mp})
	}
	return rows
}

func diskRow(d sysinfo.DiskInfo) diskData {
	return diskData{
		Mountpoint: d.Mountpoint,
		UsedPct:    d.UsedPct,
		TotalHuman: humanBytes(d.Total),
		FreeHuman:  humanBytes(d.Free),
	}
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

		data := buildSysinfoData(cfg, info, diskRowsMounted)

		var buf bytes.Buffer
		if err := imageTmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("sysinfo render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
