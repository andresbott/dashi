import type { Widget } from '@/types/dashboard'
import { GRID_COLUMNS, columnFromDrag, fitsInRow } from '@/lib/rowLayout'
import { dragOutSpan, type DragSource, type DropTarget } from '@/lib/widgetMove'

// ---------------------------------------------------------------------------
// Hit testing for the widget drag session (lib/widgetDrag.ts), on plain
// numbers only: the session measures the DOM and hands a snapshot in, so
// every rule here is testable without layout.
// ---------------------------------------------------------------------------

export interface Rect {
    left: number
    top: number
    right: number
    bottom: number
}

// clipRect is the part of rect visible through clip (a scroll container's
// box), or null when nothing is left (a touching edge counts as nothing).
// Without a clip the rect is returned as it is.
export function clipRect(rect: Rect, clip: Rect | null): Rect | null {
    if (!clip) return rect
    const out = {
        left: Math.max(rect.left, clip.left),
        top: Math.max(rect.top, clip.top),
        right: Math.min(rect.right, clip.right),
        bottom: Math.min(rect.bottom, clip.bottom),
    }
    return out.right > out.left && out.bottom > out.top ? out : null
}

// A pointer this close (px) above or below a row still targets it, so the
// margin between two rows is not a dead zone.
export const ROW_SNAP_PX = 24

// nearestRow returns the row whose vertical band is closest to y (0 inside
// it), or null when even the closest is more than snap px away. On a tie the
// first row wins.
export function nearestRow<T extends { top: number; bottom: number }>(
    rows: readonly T[],
    y: number,
    snap: number = ROW_SNAP_PX,
): T | null {
    let best: T | null = null
    let bestDist = Infinity
    for (const row of rows) {
        const dist = y < row.top ? row.top - y : y > row.bottom ? y - row.bottom : 0
        if (dist < bestDist) {
            bestDist = dist
            best = row
        }
    }
    return best && bestDist <= snap ? best : null
}

// hitStackZone returns the stack drop zone containing the point, or null. A
// dragged stack never targets a zone: a stack cannot hold another stack.
export function hitStackZone<T extends { rect: Rect }>(
    zones: readonly T[],
    x: number,
    y: number,
    draggedType: string,
): T | null {
    if (draggedType === 'stack') return null
    return zones.find((z) => x >= z.rect.left && x <= z.rect.right && y >= z.rect.top && y <= z.rect.bottom) ?? null
}

// stackInsertIndex is where a widget dropped at y lands among the stack's
// children (given their vertical midpoints, in order), counted after the
// dragged child at skipIndex (when it comes from this stack) is removed.
export function stackInsertIndex(childMidYs: readonly number[], y: number, skipIndex?: number): number {
    return childMidYs.filter((mid, i) => i !== skipIndex && mid < y).length
}

// rowDropColumn is the start column of a span-wide widget dropped on a row.
// On its own row it moves by whole columns relative to the grab point, so a
// release without moving keeps it exactly where it was; on any other row its
// moved left edge snaps to the nearest column.
export function rowDropColumn(opts: {
    ownRow: boolean
    originColumn: number
    grabStartX: number
    pointerX: number
    originLeftX: number
    gridLeft: number
    colWidth: number
    span: number
}): number {
    const max = Math.max(1, GRID_COLUMNS - opts.span + 1)
    if (opts.ownRow) {
        return columnFromDrag({
            originColumn: opts.originColumn,
            pointerX: opts.pointerX,
            grabStartX: opts.grabStartX,
            colWidth: opts.colWidth,
            min: 1,
            max,
        })
    }
    const leftX = opts.originLeftX + (opts.pointerX - opts.grabStartX)
    return Math.max(1, Math.min(max, Math.round((leftX - opts.gridLeft) / opts.colWidth) + 1))
}

export interface RowSnapshot {
    id: string
    // Vertical band of the whole row card, toolbar included.
    top: number
    bottom: number
    // The 12-column grid's horizontal extent; null while it is not mounted.
    grid: { left: number; width: number } | null
    widgets: Pick<Widget, 'width' | 'column'>[]
}

export interface StackZoneSnapshot {
    id: string
    rect: Rect
    // Vertical midpoints of the stack's children, in order.
    childMidYs: number[]
}

export interface DragSnapshot {
    source: DragSource
    draggedType: string
    // The dragged widget's own span (row sources).
    span: number
    // Index of the dragged child in its stack (stack sources).
    sourceIndex?: number
    originColumn: number
    grabStartX: number
    originLeftX: number
    pointerX: number
    pointerY: number
    rows: RowSnapshot[]
    zones: StackZoneSnapshot[]
}

export interface ResolvedDrop {
    target: DropTarget
    // Columns the widget would occupy in a row target (the guide's span);
    // the dragged span for stack targets.
    span: number
    // False when the row has no room: the guide renders blocked and the
    // release does nothing.
    fits: boolean
}

// resolveDropTarget picks the drop target under the pointer. A stack zone
// wins while the pointer is inside it; otherwise the nearest row is targeted.
// A row widget keeps its width (always fits on its own row); a stack child
// takes dragOutSpan of the row (any row, its stack's own included) and is
// blocked only when the row is full.
export function resolveDropTarget(s: DragSnapshot): ResolvedDrop | null {
    const zone = hitStackZone(s.zones, s.pointerX, s.pointerY, s.draggedType)
    if (zone) {
        const skip = s.source.kind === 'stack' && s.source.stackId === zone.id ? s.sourceIndex : undefined
        return {
            target: { kind: 'stack', stackId: zone.id, index: stackInsertIndex(zone.childMidYs, s.pointerY, skip) },
            span: s.span,
            fits: true,
        }
    }

    const row = nearestRow(s.rows, s.pointerY)
    if (!row || !row.grid) return null

    const ownRow = s.source.kind === 'row' && s.source.rowId === row.id
    let span: number
    let fits: boolean
    if (s.source.kind === 'row') {
        span = s.span
        fits = ownRow || fitsInRow(row.widgets, span)
    } else {
        const free = dragOutSpan(row.widgets)
        // A full row still shows a one-column blocked guide under the pointer.
        span = Math.max(1, free)
        fits = free >= 1
    }
    const column = rowDropColumn({
        ownRow,
        originColumn: s.originColumn,
        grabStartX: s.grabStartX,
        pointerX: s.pointerX,
        originLeftX: s.originLeftX,
        gridLeft: row.grid.left,
        colWidth: row.grid.width / GRID_COLUMNS,
        span,
    })
    return { target: { kind: 'row', rowId: row.id, column }, span, fits }
}
