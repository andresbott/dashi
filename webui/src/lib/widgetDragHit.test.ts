import { describe, it, expect } from 'vitest'
import type { Widget } from '@/types/dashboard'
import {
    ROW_SNAP_PX,
    hitStackZone,
    nearestRow,
    resolveDropTarget,
    rowDropColumn,
    stackInsertIndex,
    type DragSnapshot,
} from './widgetDragHit'

const w = (id: string, width: number, column?: number): Widget => ({ id, type: 'clock', title: id, width, column })

describe('nearestRow', () => {
    const rows = [
        { id: 'r1', top: 0, bottom: 100 },
        { id: 'r2', top: 140, bottom: 240 },
    ]

    it('returns the row under the pointer', () => {
        expect(nearestRow(rows, 50)?.id).toBe('r1')
        expect(nearestRow(rows, 200)?.id).toBe('r2')
    })

    it('snaps to the closer row inside the margin between two rows', () => {
        expect(nearestRow(rows, 110)?.id).toBe('r1')
        expect(nearestRow(rows, 130)?.id).toBe('r2')
    })

    it('misses beyond the snap margin', () => {
        expect(nearestRow(rows, 240 + ROW_SNAP_PX)?.id).toBe('r2')
        expect(nearestRow(rows, 240 + ROW_SNAP_PX + 1)).toBeNull()
        expect(nearestRow(rows, -ROW_SNAP_PX - 1)).toBeNull()
    })

    it('honours a custom snap and an empty list', () => {
        expect(nearestRow(rows, 105, 4)).toBeNull()
        expect(nearestRow([], 0)).toBeNull()
    })
})

describe('hitStackZone', () => {
    const zones = [
        { id: 's1', rect: { left: 0, top: 0, right: 100, bottom: 100 } },
        { id: 's2', rect: { left: 200, top: 0, right: 300, bottom: 100 } },
    ]

    it('returns the zone containing the point', () => {
        expect(hitStackZone(zones, 50, 50, 'clock')?.id).toBe('s1')
        expect(hitStackZone(zones, 300, 100, 'clock')?.id).toBe('s2')
    })

    it('misses outside every zone', () => {
        expect(hitStackZone(zones, 150, 50, 'clock')).toBeNull()
        expect(hitStackZone(zones, 50, 101, 'clock')).toBeNull()
    })

    it('ignores every zone while a stack is dragged', () => {
        expect(hitStackZone(zones, 50, 50, 'stack')).toBeNull()
    })
})

describe('stackInsertIndex', () => {
    const mids = [10, 30, 50]

    it('counts the children above the pointer', () => {
        expect(stackInsertIndex(mids, 0)).toBe(0)
        expect(stackInsertIndex(mids, 20)).toBe(1)
        expect(stackInsertIndex(mids, 99)).toBe(3)
        expect(stackInsertIndex([], 99)).toBe(0)
    })

    it('counts after removing the dragged child', () => {
        // Dragging child 0 below child 1: lands at index 1 once it is removed.
        expect(stackInsertIndex(mids, 40, 0)).toBe(1)
        // Dragging child 2 above child 1.
        expect(stackInsertIndex(mids, 20, 2)).toBe(1)
        // Over itself: stays where it is.
        expect(stackInsertIndex(mids, 30, 1)).toBe(1)
    })
})

describe('rowDropColumn', () => {
    const base = { originColumn: 3, grabStartX: 250, originLeftX: 200, gridLeft: 0, colWidth: 100, span: 4 }

    it('keeps the origin column in the own row without movement', () => {
        expect(rowDropColumn({ ...base, ownRow: true, pointerX: 250 })).toBe(3)
    })

    it('moves by whole columns relative to the grab in the own row', () => {
        expect(rowDropColumn({ ...base, ownRow: true, pointerX: 460 })).toBe(5)
        expect(rowDropColumn({ ...base, ownRow: true, pointerX: 2000 })).toBe(9)
    })

    it('snaps the moved left edge in another row', () => {
        // Left edge 200 + 140 = 340 against a grid starting at 100 → column 3.
        expect(rowDropColumn({ ...base, ownRow: false, gridLeft: 100, pointerX: 390 })).toBe(3)
        expect(rowDropColumn({ ...base, ownRow: false, pointerX: -500 })).toBe(1)
        expect(rowDropColumn({ ...base, ownRow: false, pointerX: 5000 })).toBe(9)
    })
})

describe('resolveDropTarget', () => {
    // Two 1200px-wide rows (100px columns): r1 holds a widget and stack s1
    // (zone: its body, x 400..800, y 40..100), r2 holds one 3-column widget.
    const snapshot = (over: Partial<DragSnapshot> = {}): DragSnapshot => ({
        source: { kind: 'row', rowId: 'r2', widgetId: 'b' },
        draggedType: 'clock',
        span: 3,
        originColumn: 1,
        grabStartX: 50,
        originLeftX: 0,
        pointerX: 50,
        pointerY: 200,
        rows: [
            { id: 'r1', top: 0, bottom: 110, grid: { left: 0, width: 1200 }, widgets: [w('a', 4, 1), w('s1', 4, 5)] },
            { id: 'r2', top: 150, bottom: 250, grid: { left: 0, width: 1200 }, widgets: [w('b', 3, 1)] },
        ],
        zones: [{ id: 's1', rect: { left: 400, top: 40, right: 800, bottom: 100 }, childMidYs: [55, 85] }],
        ...over,
    })

    it('targets the stack zone over its body', () => {
        const out = resolveDropTarget(snapshot({ pointerX: 600, pointerY: 90 }))
        expect(out).toEqual({ target: { kind: 'stack', stackId: 's1', index: 2 }, span: 3, fits: true })
    })

    it('falls through to the row over the stack header or the gaps', () => {
        // Left edge 0 + 500 → column 6.
        const out = resolveDropTarget(snapshot({ pointerX: 550, pointerY: 20 }))
        expect(out).toEqual({ target: { kind: 'row', rowId: 'r1', column: 6 }, span: 3, fits: true })
        // Between the two rows, inside the snap margin.
        expect(resolveDropTarget(snapshot({ pointerX: 550, pointerY: 120 }))?.target.kind).toBe('row')
    })

    it('ignores zones while a stack is dragged', () => {
        const out = resolveDropTarget(snapshot({ draggedType: 'stack', pointerX: 600, pointerY: 90 }))
        expect(out?.target.kind).toBe('row')
    })

    it('skips the dragged child when reordering inside its own stack', () => {
        const out = resolveDropTarget(
            snapshot({
                source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
                sourceIndex: 0,
                pointerX: 600,
                pointerY: 95,
            }),
        )
        expect(out?.target).toEqual({ kind: 'stack', stackId: 's1', index: 1 })
    })

    it('does not skip a child of another stack', () => {
        const out = resolveDropTarget(
            snapshot({
                source: { kind: 'stack', stackId: 'other', widgetId: 'c1' },
                sourceIndex: 0,
                pointerX: 600,
                pointerY: 95,
            }),
        )
        expect(out?.target).toEqual({ kind: 'stack', stackId: 's1', index: 2 })
    })

    it('keeps the origin column on the own row without movement', () => {
        const out = resolveDropTarget(snapshot({ originColumn: 1 }))
        expect(out).toEqual({ target: { kind: 'row', rowId: 'r2', column: 1 }, span: 3, fits: true })
    })

    it('marks another row without room as not fitting', () => {
        const out = resolveDropTarget(snapshot({ span: 5, pointerY: 20 }))
        expect(out).toMatchObject({ target: { kind: 'row', rowId: 'r1' }, span: 5, fits: false })
    })

    it('always fits on the own row', () => {
        const out = resolveDropTarget(
            snapshot({ source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, span: 4, pointerY: 20 }),
        )
        expect(out).toMatchObject({ target: { kind: 'row', rowId: 'r1' }, fits: true })
    })

    it('shrinks a stack child to the free columns of a partly full row', () => {
        const out = resolveDropTarget(
            snapshot({
                source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
                sourceIndex: 0,
                span: 4,
                originLeftX: 400,
                grabStartX: 420,
                pointerX: 1120,
                pointerY: 20,
            }),
        )
        // Left edge 400 + 700 = 1100 → column 12, clamped to 13 - 4 = 9.
        expect(out).toEqual({ target: { kind: 'row', rowId: 'r1', column: 9 }, span: 4, fits: true })
    })

    it('gives a stack child half the grid on a roomy row, snapping the left edge', () => {
        const out = resolveDropTarget(
            snapshot({
                source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
                sourceIndex: 0,
                originLeftX: 400,
                grabStartX: 420,
                pointerX: 320,
                pointerY: 200,
            }),
        )
        // Left edge 400 - 100 = 300 → column 4.
        expect(out).toEqual({ target: { kind: 'row', rowId: 'r2', column: 4 }, span: 6, fits: true })
    })

    it('blocks a stack child on a full row', () => {
        const rows = snapshot().rows
        rows[0] = { ...rows[0], widgets: [w('a', 8, 1), w('s1', 4, 9)] }
        const out = resolveDropTarget(
            snapshot({ source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, sourceIndex: 0, rows, pointerY: 20 }),
        )
        expect(out).toMatchObject({ target: { kind: 'row', rowId: 'r1' }, fits: false })
    })

    it('misses when the pointer is far from every row and zone', () => {
        expect(resolveDropTarget(snapshot({ pointerY: 1000 }))).toBeNull()
    })

    it('misses a row whose grid is not mounted', () => {
        const rows = snapshot().rows
        rows[1] = { ...rows[1], grid: null }
        expect(resolveDropTarget(snapshot({ rows }))).toBeNull()
    })
})
