package dashboard

import "encoding/json"

// ImageConfig holds rendering settings for image-type dashboards.
type ImageConfig struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

// Container controls the global wrapper around all rows.
type Container struct {
	MaxWidth        string `json:"maxWidth"`
	VerticalAlign   string `json:"verticalAlign"`
	HorizontalAlign string `json:"horizontalAlign"`
	ShowBoxes       bool   `json:"showBoxes,omitempty"`
}

// Dashboard represents a user-defined dashboard with a layout of widgets in rows.
type Dashboard struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	Type        string      `json:"type"`
	Container   Container   `json:"container"`
	ImageConfig *ImageConfig `json:"imageConfig,omitempty"`
	Rows        []Row       `json:"rows"`
}

// DashboardMeta is the lightweight listing representation (no rows).
type DashboardMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
	Type string `json:"type"`
}

// Row represents a horizontal section of the dashboard.
type Row struct {
	ID      string   `json:"id"`
	Title   string   `json:"title,omitempty"`
	Height  string   `json:"height"`
	Width   string   `json:"width"`
	Widgets []Widget `json:"widgets"`
}

// Widget represents a single widget placed within a row.
type Widget struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Title  string          `json:"title"`
	Width  int             `json:"width"`
	Config json.RawMessage `json:"config,omitempty"`
}
