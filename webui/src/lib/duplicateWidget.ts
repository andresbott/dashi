import type { Widget } from '@/types/dashboard'
import { GRID_COLUMNS, placeRow, totalSpan, clampRowToLine } from '@/lib/rowLayout'
import { plainCopy } from '@/lib/plainCopy'

// cloneWidget returns a detached deep copy of widget under a fresh id. A stack
// keeps its children in config.widgets; each child gets a fresh id too, so no
// id on the dashboard is ever shared between the original and the copy.
export function cloneWidget(widget: Widget, newId: () => string): Widget {
    const copy = plainCopy(widget)
    copy.id = newId()
    const children = copy.config?.widgets
    if (Array.isArray(children)) {
        for (const child of children as Widget[]) child.id = newId()
    }
    return copy
}

// duplicateInRow returns a copy of the row with a duplicate of the widget at
// index inserted right after it, starting at the column where the original
// ends. The copy keeps the original's width when the row has room for it and
// shrinks to the free columns otherwise; right-hand neighbours are pushed
// along and the row is kept on one line. A full row (or an index outside it)
// is returned unchanged.
export function duplicateInRow(widgets: Widget[], index: number, newId: () => string): Widget[] {
    const placed = placeRow(widgets)[index]
    if (!placed) return widgets
    const width = Math.min(placed.width, GRID_COLUMNS - totalSpan(widgets))
    if (width < 1) return widgets
    const copy = { ...cloneWidget(widgets[index], newId), width, column: placed.column + placed.width }
    const out = [...widgets]
    out.splice(index + 1, 0, copy)
    return clampRowToLine(out)
}
