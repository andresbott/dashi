import type { Widget } from '@/types/dashboard'

// The 12-column grid a row is divided into. Mirrors dashboard.GridColumns in
// the Go backend (internal/dashboard/layout.go).
export const GRID_COLUMNS = 12

export interface Placement {
    // Resolved 1-based start column.
    column: number
    // Empty columns before this widget on its line (rendered as col-offset-N).
    gap: number
    // Clamped column span (rendered as col-N).
    width: number
}

// Only the layout-relevant slice of a widget, so callers can pass preview
// overrides mid-drag without constructing a full Widget.
type Positioned = Pick<Widget, 'width' | 'column'>

function span(width: number | undefined): number {
    return width && width >= 1 ? width : GRID_COLUMNS
}

// placeRow resolves a row's widgets onto the 12-column grid, one Placement per
// widget in the same order. It is the exact mirror of Go's PlaceRow: a running
// cursor tracks the next free column and any widget whose desired start would
// collide with what precedes it is pushed right, so the result never overlaps.
// Column === 0/undefined flows after the previous widget; width < 1 becomes a
// full 12-column span.
export function placeRow(widgets: Positioned[]): Placement[] {
    const out: Placement[] = []
    let cursor = 1
    for (const w of widgets) {
        const width = span(w.width)
        let start = w.column && w.column >= 1 ? w.column : cursor
        if (start > GRID_COLUMNS) start = GRID_COLUMNS
        if (start < cursor) start = cursor
        out.push({ column: start, gap: start - cursor, width })
        cursor = start + width
    }
    return out
}

// insertByColumn returns a copy of widgets with widget spliced in at the array
// position matching its column, so placeRow honours the column without
// reordering — or disturbing the columns of — the existing widgets. A widget
// with no column (0) flows onto the end. This is the add path: existing
// widgets (including un-positioned ones from older dashboards) keep their spot.
export function insertByColumn(widgets: Widget[], widget: Widget): Widget[] {
    const out = [...widgets]
    if (!widget.column || widget.column < 1) {
        out.push(widget)
        return out
    }
    const placed = placeRow(widgets)
    let idx = out.length
    for (let i = 0; i < placed.length; i++) {
        if (placed[i].column >= widget.column) {
            idx = i
            break
        }
    }
    out.splice(idx, 0, widget)
    return out
}

// firstFreeSpan finds where a newly added widget should go: the first free run
// of columns, with its width shrunk to fit that run. If the row is already full
// it returns column 0 (flow), so the widget wraps onto a new line rather than
// overlapping. desiredWidth is clamped to the 1..12 range.
export function firstFreeSpan(widgets: Widget[], desiredWidth: number): { column: number; width: number } {
    const want = Math.max(1, Math.min(GRID_COLUMNS, desiredWidth))
    const occupied = new Array<boolean>(GRID_COLUMNS + 1).fill(false) // 1-based
    for (const p of placeRow(widgets)) {
        for (let c = p.column; c < p.column + p.width && c <= GRID_COLUMNS; c++) {
            occupied[c] = true
        }
    }
    let c = 1
    while (c <= GRID_COLUMNS) {
        if (occupied[c]) {
            c++
            continue
        }
        const start = c
        let len = 0
        while (c <= GRID_COLUMNS && !occupied[c]) {
            len++
            c++
        }
        return { column: start, width: Math.min(want, len) }
    }
    return { column: 0, width: want } // row full: flow onto a new line
}

// columnFromDrag resolves the start column of a widget being dragged by its
// (centred) move handle. It works from the number of columns the pointer has
// moved since the grab, not the pointer's absolute position — so releasing
// without moving keeps the widget exactly where it was, with a symmetric
// half-column dead zone around the origin. Where within the widget it was
// grabbed is therefore irrelevant. Result is clamped to [min, max].
export function columnFromDrag(opts: {
    originColumn: number
    pointerX: number
    grabStartX: number
    colWidth: number
    min: number
    max: number
}): number {
    const delta = Math.round((opts.pointerX - opts.grabStartX) / opts.colWidth)
    return Math.max(opts.min, Math.min(opts.max, opts.originColumn + delta))
}
