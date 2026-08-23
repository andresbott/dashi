package dashboard

// GridColumns is the fixed number of columns a row is divided into.
const GridColumns = 12

// Placement is the resolved layout of one widget on the 12-column grid: Gap is
// the number of empty columns before it on its line, and Width is its clamped
// column span. PlaceRow produces one Placement per widget.
type Placement struct {
	Gap   int
	Width int
}

// PlaceRow resolves a row's widgets onto the 12-column grid, returning one
// Placement per widget in the same order.
//
// Each widget's Column (1-based) is where it wants to start; Column == 0 means
// "flow immediately after the previous widget" — the pre-column behaviour, so
// dashboards created before column support render unchanged. Width < 1 defaults
// to a full 12-column span, matching the historical renderer rule.
//
// Placement is overlap-proof: a running cursor tracks the next free column, and
// any widget whose desired start would collide with what precedes it is pushed
// right to the cursor. The leading Gap is the distance from the cursor to the
// resolved start. Rows whose spans exceed 12 columns still overflow and wrap in
// the templates, exactly as before — which is why start is only clamped up to
// GridColumns, never the running cursor.
func PlaceRow(widgets []Widget) []Placement {
	placements := make([]Placement, len(widgets))
	cursor := 1
	for i, w := range widgets {
		width := w.Width
		if width < 1 {
			width = GridColumns
		}
		start := cursor
		if w.Column >= 1 {
			start = w.Column
		}
		if start > GridColumns {
			start = GridColumns
		}
		if start < cursor {
			start = cursor
		}
		placements[i] = Placement{Gap: start - cursor, Width: width}
		cursor = start + width
	}
	return placements
}
