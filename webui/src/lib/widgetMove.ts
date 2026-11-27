import type { Row, Widget } from '@/types/dashboard'
import {
    GRID_COLUMNS,
    clampRowToLine,
    fitsInRow,
    insertByColumn,
    moveWidgetToColumn,
    totalSpan,
} from '@/lib/rowLayout'

// ---------------------------------------------------------------------------
// Widget moves in the dashboard editor, as data: where a dragged widget came
// from and where it was dropped. applyWidgetMove is the single reducer every
// drop goes through (row ↔ row, row ↔ stack, stack ↔ stack), so the editor
// assigns the page's rows exactly once per drop.
// ---------------------------------------------------------------------------

export type DragSource =
    | { kind: 'row'; rowId: string; widgetId: string }
    | { kind: 'stack'; stackId: string; widgetId: string }

export type DropTarget =
    | { kind: 'row'; rowId: string; column: number }
    // index is the insert position counted after the source was removed.
    | { kind: 'stack'; stackId: string; index: number }

export interface WidgetMove {
    source: DragSource
    target: DropTarget
}

// The widest a widget dragged out of a stack lands in a row: half the grid.
const DRAG_OUT_WIDTH = 6

// dragOutSpan is the width a stack child takes when dropped into a row: half
// the grid, shrunk to the row's free columns. Below 1 the row is full and the
// drop is refused. The child's stored width is never used: it is stale by
// design (children always render at their stack's width).
export function dragOutSpan(rowWidgets: Pick<Widget, 'width' | 'column'>[]): number {
    return Math.min(DRAG_OUT_WIDTH, GRID_COLUMNS - totalSpan(rowWidgets))
}

// findStack locates the stack widget with the given id, or null when there is
// no such widget or it is not a stack.
export function findStack(rows: Row[], stackId: string): { rowIndex: number; widgetIndex: number } | null {
    for (let rowIndex = 0; rowIndex < rows.length; rowIndex++) {
        const widgetIndex = rows[rowIndex].widgets.findIndex((w) => w.id === stackId && w.type === 'stack')
        if (widgetIndex >= 0) return { rowIndex, widgetIndex }
    }
    return null
}

function stackChildren(stack: Widget): Widget[] {
    const list = (stack.config as { widgets?: Widget[] } | undefined)?.widgets
    return Array.isArray(list) ? list : []
}

function withChildren(stack: Widget, widgets: Widget[]): Widget {
    return { ...stack, config: { ...stack.config, widgets } }
}

function replaceRow(rows: Row[], index: number, widgets: Widget[]): Row[] {
    const out = [...rows]
    out[index] = { ...rows[index], widgets }
    return out
}

// asChild shapes a widget for life inside a stack: the stack's width and no
// grid column (children stack vertically).
function asChild(widget: Widget, width: number): Widget {
    const child: Widget = { ...widget, width }
    delete child.column
    return child
}

function insertAt<T>(list: T[], index: number, item: T): T[] {
    const out = [...list]
    out.splice(Math.max(0, Math.min(list.length, index)), 0, item)
    return out
}

// insertIntoStack inserts widget into the stack stackId at index, looking the
// stack up in rows (so it works after the source was removed from the same
// row). Null when the stack does not exist.
function insertIntoStack(rows: Row[], stackId: string, index: number, widget: Widget): Row[] | null {
    const at = findStack(rows, stackId)
    if (!at) return null
    const row = rows[at.rowIndex]
    const stack = row.widgets[at.widgetIndex]
    const widgets = [...row.widgets]
    widgets[at.widgetIndex] = withChildren(stack, insertAt(stackChildren(stack), index, asChild(widget, stack.width)))
    return replaceRow(rows, at.rowIndex, widgets)
}

function moveFromRow(rows: Row[], source: Extract<DragSource, { kind: 'row' }>, target: DropTarget): Row[] | null {
    const fromIndex = rows.findIndex((r) => r.id === source.rowId)
    if (fromIndex < 0) return null
    const fromRow = rows[fromIndex]
    const widgetIndex = fromRow.widgets.findIndex((w) => w.id === source.widgetId)
    if (widgetIndex < 0) return null
    const widget = fromRow.widgets[widgetIndex]

    if (target.kind === 'row' && target.rowId === source.rowId) {
        return replaceRow(rows, fromIndex, clampRowToLine(moveWidgetToColumn(fromRow.widgets, widgetIndex, target.column)))
    }

    if (target.kind === 'stack' && widget.type === 'stack') return null // no stack in a stack

    const removed = replaceRow(rows, fromIndex, fromRow.widgets.filter((_, i) => i !== widgetIndex))

    if (target.kind === 'stack') return insertIntoStack(removed, target.stackId, target.index, widget)

    const toIndex = removed.findIndex((r) => r.id === target.rowId)
    if (toIndex < 0) return null
    const toRow = removed[toIndex]
    if (!fitsInRow(toRow.widgets, widget.width)) return null
    return replaceRow(removed, toIndex, clampRowToLine(insertByColumn(toRow.widgets, { ...widget, column: target.column })))
}

function moveFromStack(rows: Row[], source: Extract<DragSource, { kind: 'stack' }>, target: DropTarget): Row[] | null {
    const at = findStack(rows, source.stackId)
    if (!at) return null
    const stack = rows[at.rowIndex].widgets[at.widgetIndex]
    const list = stackChildren(stack)
    const childIndex = list.findIndex((c) => c.id === source.widgetId)
    if (childIndex < 0) return null
    const child = list[childIndex]

    const rowWidgets = [...rows[at.rowIndex].widgets]
    rowWidgets[at.widgetIndex] = withChildren(stack, list.filter((_, i) => i !== childIndex))
    const removed = replaceRow(rows, at.rowIndex, rowWidgets)

    if (target.kind === 'stack') return insertIntoStack(removed, target.stackId, target.index, child)

    const toIndex = removed.findIndex((r) => r.id === target.rowId)
    if (toIndex < 0) return null
    const toRow = removed[toIndex]
    const width = dragOutSpan(toRow.widgets)
    if (width < 1) return null
    const column = Math.max(1, Math.min(GRID_COLUMNS - width + 1, target.column))
    return replaceRow(removed, toIndex, clampRowToLine(insertByColumn(toRow.widgets, { ...child, width, column })))
}

// applyWidgetMove returns the page's rows with the move applied, or null when
// the move is refused (unknown ids, a stack dropped into a stack, no room in
// the target row). It never mutates rows and accepts reactive input: rows,
// widgets and stack configs on the moved path are rebuilt, untouched rows keep
// their identity.
//
// Placement rules:
//   - row → same row: reposition/reorder at the column, kept on one line.
//   - row → other row: keeps its width; refused unless it fits.
//   - row → stack: inserted at index with the stack's width, no column.
//   - stack → row (its own row included): width = dragOutSpan(target row),
//     column clamped so the span stays on the grid.
//   - stack → stack: inserted at index with the target stack's width.
export function applyWidgetMove(rows: Row[], move: WidgetMove): Row[] | null {
    return move.source.kind === 'row'
        ? moveFromRow(rows, move.source, move.target)
        : moveFromStack(rows, move.source, move.target)
}
