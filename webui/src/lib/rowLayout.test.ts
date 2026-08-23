import { describe, it, expect } from 'vitest'
import type { Widget } from '@/types/dashboard'
import { placeRow, insertByColumn, firstFreeSpan, columnFromDrag, moveWidgetToColumn } from './rowLayout'

// Minimal widget factory — only width/column matter to the layout.
let seq = 0
function w(width: number, column?: number): Widget {
    return { id: `w${seq++}`, type: 'test', title: 't', width, column }
}

describe('placeRow', () => {
    it('flows widgets left to right when no column is set', () => {
        expect(placeRow([w(6), w(6)])).toEqual([
            { column: 1, gap: 0, width: 6 },
            { column: 7, gap: 0, width: 6 },
        ])
    })

    it('defaults a width below one to a full span', () => {
        expect(placeRow([w(0)])).toEqual([{ column: 1, gap: 0, width: 12 }])
    })

    it('turns an explicit column into a leading gap', () => {
        expect(placeRow([w(6, 4)])).toEqual([{ column: 4, gap: 3, width: 6 }])
    })

    it('honours the second widget column relative to the cursor', () => {
        expect(placeRow([w(3), w(6, 7)])).toEqual([
            { column: 1, gap: 0, width: 3 },
            { column: 7, gap: 3, width: 6 },
        ])
    })

    it('pushes an overlapping column right of the previous widget', () => {
        expect(placeRow([w(6), w(6, 3)])).toEqual([
            { column: 1, gap: 0, width: 6 },
            { column: 7, gap: 0, width: 6 },
        ])
    })

    it('shoves packed neighbours along when a widget is widened into them (resize push)', () => {
        // Packed row A|B|C, each width 4. Widening A to 6 pushes B and C right
        // rather than overlapping; C spills past column 12 and wraps (the grid
        // is flex-wrap). This is the contract the editor resize handle relies on.
        expect(placeRow([w(6, 1), w(4, 5), w(4, 9)])).toEqual([
            { column: 1, gap: 0, width: 6 },
            { column: 7, gap: 0, width: 4 },
            { column: 11, gap: 0, width: 4 },
        ])
    })

    it('clamps a column beyond the grid to the last column', () => {
        expect(placeRow([w(4, 20)])).toEqual([{ column: 12, gap: 11, width: 4 }])
    })
})

describe('insertByColumn', () => {
    it('appends a widget with no column onto the end (flow)', () => {
        const existing = [w(3, 1)]
        const added = w(3)
        const out = insertByColumn(existing, added)
        expect(out.map((x) => x.id)).toEqual([existing[0].id, added.id])
    })

    it('splices ahead of a widget with a higher column', () => {
        const a = w(3, 7)
        const added = w(3, 1)
        const out = insertByColumn([a], added)
        expect(out.map((x) => x.id)).toEqual([added.id, a.id])
    })

    it('fills a middle gap without reordering existing widgets', () => {
        // A at 1..3, B at 7..12; the free gap is 4..6.
        const a = w(3, 1)
        const b = w(6, 7)
        const added = w(3, 4)
        const out = insertByColumn([a, b], added)
        expect(out.map((x) => x.id)).toEqual([a.id, added.id, b.id])
        // Placement stays clean: A 1..3, added 4..6, B 7..12.
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 4, 7])
    })

    it('does not move an un-positioned existing widget when inserting ahead of it', () => {
        // Old-dashboard widget with no column resolves to column 1; a new
        // column-4 widget must go after it, leaving it flush left.
        const legacy = w(3)
        const added = w(6, 4)
        const out = insertByColumn([legacy], added)
        expect(out.map((x) => x.id)).toEqual([legacy.id, added.id])
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 4])
    })
})

describe('firstFreeSpan', () => {
    it('places the first widget flush left at the desired width', () => {
        expect(firstFreeSpan([], 6)).toEqual({ column: 1, width: 6 })
    })

    it('fills the leading gap before a positioned widget, shrinking to fit', () => {
        // A span-6 widget at column 4 occupies 4..9; the first free run is 1..3.
        expect(firstFreeSpan([w(6, 4)], 6)).toEqual({ column: 1, width: 3 })
    })

    it('appends in the trailing free space', () => {
        expect(firstFreeSpan([w(6, 1)], 6)).toEqual({ column: 7, width: 6 })
    })

    it('flows onto a new line when the row is full', () => {
        expect(firstFreeSpan([w(12, 1)], 6)).toEqual({ column: 0, width: 6 })
    })
})

describe('columnFromDrag', () => {
    // 100px per column; widget starts at column 4; valid start range 1..10.
    const base = { colWidth: 100, originColumn: 4, min: 1, max: 10 }

    it('keeps the widget in place when released without moving (the reported bug)', () => {
        expect(columnFromDrag({ ...base, pointerX: 500, grabStartX: 500 })).toBe(4)
    })

    it('ignores jitter smaller than half a column in either direction', () => {
        expect(columnFromDrag({ ...base, pointerX: 540, grabStartX: 500 })).toBe(4)
        expect(columnFromDrag({ ...base, pointerX: 460, grabStartX: 500 })).toBe(4)
    })

    it('moves a 3-wide widget from cols 4-6 to cols 1-3 (three columns left)', () => {
        expect(columnFromDrag({ ...base, pointerX: 200, grabStartX: 500 })).toBe(1)
    })

    it('moves right by whole columns', () => {
        expect(columnFromDrag({ ...base, pointerX: 700, grabStartX: 500 })).toBe(6)
    })

    it('clamps to the min and max start columns', () => {
        expect(columnFromDrag({ ...base, pointerX: -9000, grabStartX: 500 })).toBe(1)
        expect(columnFromDrag({ ...base, pointerX: 9000, grabStartX: 500 })).toBe(10)
    })
})

describe('moveWidgetToColumn', () => {
    it('moves a widget to an earlier slot, reordering the array (the reported bug)', () => {
        // A at 1..6, B at 7..12; dragging B to column 1 must put B first.
        const a = w(6, 1)
        const b = w(6, 7)
        const out = moveWidgetToColumn([a, b], 1, 1)
        expect(out.map((x) => x.id)).toEqual([b.id, a.id])
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 7])
    })

    it('moves a widget to a later slot, reordering the array', () => {
        // Mirror case: dragging A right onto B's column swaps them cleanly.
        const a = w(6, 1)
        const b = w(6, 7)
        const out = moveWidgetToColumn([a, b], 0, 7)
        expect(out.map((x) => x.id)).toEqual([b.id, a.id])
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 7])
    })

    it('reorders into the middle of a full row and packs cleanly', () => {
        const a = w(4, 1)
        const b = w(4, 5)
        const c = w(4, 9)
        const out = moveWidgetToColumn([a, b, c], 0, 5)
        expect(out.map((x) => x.id)).toEqual([b.id, a.id, c.id])
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 5, 9])
    })

    it('honours a leading gap for the dragged widget on reorder', () => {
        // Drag A past B to the far right; B packs left, A keeps its drop column.
        const a = w(4, 1)
        const b = w(4, 5)
        const out = moveWidgetToColumn([a, b], 0, 9)
        expect(out.map((x) => x.id)).toEqual([b.id, a.id])
        expect(placeRow(out).map((p) => p.column)).toEqual([1, 9])
    })

    it('repositions within its own slot without reordering or disturbing others', () => {
        // A at 1..3, B parked at 9..11 (gap 4..8). Nudging A to column 2 must
        // not touch B's column, so its intentional gap survives.
        const a = w(3, 1)
        const b = w(3, 9)
        const out = moveWidgetToColumn([a, b], 0, 2)
        expect(out.map((x) => x.id)).toEqual([a.id, b.id])
        expect(out[0].column).toBe(2)
        expect(out[1].column).toBe(9)
    })

    it('returns the widgets unchanged for an out-of-range index', () => {
        const a = w(6, 1)
        const out = moveWidgetToColumn([a], 5, 1)
        expect(out.map((x) => x.id)).toEqual([a.id])
    })
})
