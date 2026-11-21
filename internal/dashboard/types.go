package dashboard

import "encoding/json"

// Container controls the global wrapper around all rows.
type Container struct {
	MaxWidth        string `json:"maxWidth"`
	VerticalAlign   string `json:"verticalAlign"`
	HorizontalAlign string `json:"horizontalAlign"`
	ShowBoxes       bool   `json:"showBoxes,omitempty"`
}

// Page represents a single page within a dashboard, containing its own rows.
type Page struct {
	Name            string `json:"name"`
	RefreshInterval int    `json:"refreshInterval,omitempty"`
	Rows            []Row  `json:"rows"`
}

// Dashboard represents a user-defined dashboard with a layout of widgets in rows.
type Dashboard struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Type      string    `json:"type"`
	Default   bool      `json:"default,omitempty"`
	Container Container `json:"container"`
	// ThemeID references a theme entity (internal/themes): colours and page
	// background. Empty, unknown or dangling references render with the
	// builtin default theme.
	ThemeID string `json:"themeId,omitempty"`
	// WidgetTheme names a widget theme (internal/widgetthemes): fonts and
	// icons. Empty or unknown names render with the default widget theme.
	WidgetTheme string `json:"widgetTheme,omitempty"`
	ColorMode   string `json:"colorMode,omitempty"`
	// Dither picks how an image dashboard is dithered for the e-ink formats
	// (png-bw, png-spectra6, bw, spectra6): empty is plain Floyd-Steinberg,
	// "auto" lets einkimage pick per render, any other value names an
	// einkimage processing preset. See DitherModes; the store rejects
	// anything else. Interactive dashboards ignore it.
	Dither string `json:"dither,omitempty"`
	Pages  []Page `json:"pages"`
}

// DashboardMeta is the lightweight listing representation (no rows).
type DashboardMeta struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Type    string `json:"type"`
	Default bool   `json:"default,omitempty"`
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
//
// Column is the 1-based grid column (1..12) where the widget starts, letting a
// widget sit anywhere in the row. A value
// of 0 (the zero value, and what older dashboards carry) means "flow after the
// previous widget", so pre-existing dashboards render unchanged. Width is the
// span in columns. Placement (column → leading gap, plus overlap resolution)
// is computed by PlaceRow.
type Widget struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Title  string          `json:"title"`
	Width  int             `json:"width"`
	Column int             `json:"column,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

// Auth holds per-dashboard basic auth credentials.
// Stored as a sidecar auth.json file in the dashboard directory.
type Auth struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}
