import { describe, it, expect, afterEach, vi } from 'vitest'
import type { Widget } from '@/types/dashboard'
import { DRAG_GUARD_CLASS } from './dragGuard'
import {
    abortIfSource,
    dragState,
    endWidgetDrag,
    registerRow,
    registerStackZone,
    startWidgetDrag,
} from './widgetDrag'

const w = (id: string, width: number, column?: number): Widget => ({ id, type: 'clock', title: id, width, column })

// An element whose layout box is fixed, since jsdom does no layout.
function box(left: number, top: number, width: number, height: number): HTMLElement {
    const el = document.createElement('div')
    el.getBoundingClientRect = () =>
        ({ left, top, width, height, right: left + width, bottom: top + height, x: left, y: top }) as DOMRect
    return el
}

const cleanups: (() => void)[] = []

// Page under test: rows r1 (y 0..110) and r2 (y 150..250), 1200px grids
// (100px columns). r1 holds stack s1 whose body (x 400..800, y 40..100) has
// children with midpoints 55 and 85.
function setup() {
    const r1Widgets = [w('a', 4, 1), { ...w('s1', 4, 5), type: 'stack' }]
    const r2Widgets = [w('b', 3, 1)]
    cleanups.push(
        registerRow('r1', { root: box(0, 0, 1200, 110), grid: () => box(0, 30, 1200, 80), widgets: () => r1Widgets }),
        registerRow('r2', { root: box(0, 150, 1200, 100), grid: () => box(0, 180, 1200, 70), widgets: () => r2Widgets }),
        registerStackZone('s1', {
            zone: () => box(400, 40, 400, 60),
            children: () => [box(400, 45, 400, 20), box(400, 75, 400, 20)],
        }),
    )
}

const startRowDrag = () =>
    startWidgetDrag(
        { kind: 'row', rowId: 'r2', widgetId: 'b' },
        { originColumn: 1, originLeftX: 0, grabX: 50, grabY: 200, span: 3, draggedType: 'clock' },
    )

const pointerTo = (clientX: number, clientY: number) =>
    document.dispatchEvent(new MouseEvent('pointermove', { clientX, clientY }))

afterEach(() => {
    endWidgetDrag()
    while (cleanups.length) cleanups.pop()!()
    vi.restoreAllMocks()
})

describe('widget drag session', () => {
    it('targets the row under the grab point on start, guarded', () => {
        setup()
        startRowDrag()
        expect(dragState).toMatchObject({
            active: true,
            sourceKind: 'row',
            sourceRowId: 'r2',
            sourceWidgetId: 'b',
            draggedType: 'clock',
            targetKind: 'row',
            targetRowId: 'r2',
            targetColumn: 1,
            targetSpan: 3,
            targetFits: true,
        })
        expect(document.body.classList.contains(DRAG_GUARD_CLASS)).toBe(true)
    })

    it('targets a stack zone and its insert index on pointer move', () => {
        setup()
        startRowDrag()
        pointerTo(600, 70)
        expect(dragState).toMatchObject({ targetKind: 'stack', targetStackId: 's1', targetIndex: 1, targetRowId: null })
        expect(endWidgetDrag()).toEqual({
            source: { kind: 'row', rowId: 'r2', widgetId: 'b' },
            target: { kind: 'stack', stackId: 's1', index: 1 },
        })
    })

    it('returns the row move on release', () => {
        setup()
        startRowDrag()
        pointerTo(350, 200)
        expect(endWidgetDrag()).toEqual({
            source: { kind: 'row', rowId: 'r2', widgetId: 'b' },
            target: { kind: 'row', rowId: 'r2', column: 4 },
        })
    })

    it('skips the dragged child when a stack child is reordered', () => {
        setup()
        startWidgetDrag(
            { kind: 'stack', stackId: 's1', widgetId: 'c1' },
            { originColumn: 1, originLeftX: 400, grabX: 420, grabY: 55, span: 4, draggedType: 'clock', sourceIndex: 0 },
        )
        pointerTo(600, 95)
        expect(dragState).toMatchObject({ sourceKind: 'stack', sourceStackId: 's1', targetIndex: 1 })
    })

    it('refuses the release over a row without room', () => {
        setup()
        startWidgetDrag(
            { kind: 'row', rowId: 'r2', widgetId: 'b' },
            { originColumn: 1, originLeftX: 0, grabX: 50, grabY: 200, span: 5, draggedType: 'clock' },
        )
        pointerTo(50, 20)
        expect(dragState).toMatchObject({ targetKind: 'row', targetRowId: 'r1', targetFits: false })
        expect(endWidgetDrag()).toBeNull()
    })

    it('clears the target away from every row', () => {
        setup()
        startRowDrag()
        pointerTo(50, 900)
        expect(dragState).toMatchObject({ targetKind: null, targetRowId: null, targetColumn: null })
        expect(endWidgetDrag()).toBeNull()
    })

    it('re-resolves the target on scroll', () => {
        setup()
        startRowDrag()
        pointerTo(600, 70)
        cleanups.pop()!() // the stack zone scrolls away
        window.dispatchEvent(new Event('scroll'))
        expect(dragState.targetKind).toBe('row')
    })

    it('cancels on Escape: the release finishes the drag and returns nothing', () => {
        setup()
        startRowDrag()
        const mouseup = vi.fn()
        document.addEventListener('mouseup', mouseup)
        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
        document.removeEventListener('mouseup', mouseup)
        expect(mouseup).toHaveBeenCalledOnce()
        expect(dragState).toMatchObject({ cancelled: true, targetKind: null })
        pointerTo(600, 70) // ignored once cancelled
        expect(dragState.targetKind).toBeNull()
        expect(endWidgetDrag()).toBeNull()
        expect(dragState.cancelled).toBe(false)
    })

    it('releases the body guard and resets the state on release', () => {
        setup()
        startRowDrag()
        endWidgetDrag()
        expect(document.body.classList.contains(DRAG_GUARD_CLASS)).toBe(false)
        expect(dragState).toMatchObject({ active: false, sourceKind: null, targetKind: null, cancelled: false })
    })

    it('aborts when the source unregisters mid-drag', () => {
        setup()
        startRowDrag()
        const removed = vi.spyOn(document, 'removeEventListener')
        // The source row unmounts (page switch, row delete).
        cleanups.splice(1, 1)[0]()
        abortIfSource('r2')
        expect(removed).toHaveBeenCalledWith('pointermove', expect.any(Function))
        expect(removed).toHaveBeenCalledWith('keydown', expect.any(Function))
        expect(document.body.classList.contains(DRAG_GUARD_CLASS)).toBe(false)
        expect(dragState.active).toBe(false)
        pointerTo(600, 70)
        expect(dragState.targetKind).toBeNull()
    })

    it('aborts a stack child drag when its stack goes away, and ignores other ids', () => {
        setup()
        startWidgetDrag(
            { kind: 'stack', stackId: 's1', widgetId: 'c1' },
            { originColumn: 1, originLeftX: 400, grabX: 420, grabY: 55, span: 4, draggedType: 'clock', sourceIndex: 0 },
        )
        abortIfSource('r1')
        expect(dragState.active).toBe(true)
        abortIfSource('s1')
        expect(dragState.active).toBe(false)
    })

    it('unregistering keeps a newer registration under the same id', () => {
        const stale = registerRow('r1', { root: box(0, 0, 1200, 110), grid: () => null, widgets: () => [] })
        cleanups.push(
            registerRow('r1', { root: box(0, 0, 1200, 110), grid: () => box(0, 30, 1200, 80), widgets: () => [] }),
        )
        stale()
        startWidgetDrag(
            { kind: 'row', rowId: 'r1', widgetId: 'a' },
            { originColumn: 1, originLeftX: 0, grabX: 50, grabY: 50, span: 3, draggedType: 'clock' },
        )
        expect(dragState.targetRowId).toBe('r1')
    })
})
