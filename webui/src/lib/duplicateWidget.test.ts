import { describe, it, expect } from 'vitest'
import { reactive } from 'vue'
import type { Widget } from '@/types/dashboard'
import { cloneWidget, duplicateInRow } from './duplicateWidget'
import { placeRow } from './rowLayout'

function ids() {
    let n = 0
    return () => `new-${++n}`
}

const w = (id: string, width: number, column?: number, extra: Partial<Widget> = {}): Widget => ({
    id,
    type: 'clock',
    title: id,
    width,
    column,
    ...extra,
})

describe('cloneWidget', () => {
    it('copies title and config with a fresh id', () => {
        const original = w('a', 4, 1, { title: 'Kitchen', config: { tz: 'UTC', list: [1, 2] } })
        const copy = cloneWidget(original, ids())
        expect(copy).toEqual({ ...original, id: 'new-1' })
    })

    it('detaches nested config from the original', () => {
        const original = w('a', 4, 1, { config: { list: [1, 2] } })
        const copy = cloneWidget(original, ids())
        ;(copy.config!.list as number[]).push(3)
        expect(original.config!.list).toEqual([1, 2])
    })

    it('gives every stack child a fresh id too', () => {
        const original = w('s', 4, 1, {
            type: 'stack',
            config: { widgets: [w('c1', 4, undefined, { config: { x: 1 } }), w('c2', 4)] },
        })
        const copy = cloneWidget(original, ids())
        const children = copy.config!.widgets as Widget[]
        expect(copy.id).toBe('new-1')
        expect(children.map((c) => c.id)).toEqual(['new-2', 'new-3'])
        expect(children[0].config).toEqual({ x: 1 })
        expect((original.config!.widgets as Widget[]).map((c) => c.id)).toEqual(['c1', 'c2'])
    })

    it('clones widgets holding reactive proxies deep inside', () => {
        // The editor builds new rows/configs around existing (reactive)
        // children, so a raw widget can still hold proxies below the top.
        const child = reactive(w('c1', 4))
        const original = reactive(w('s', 4, 1, { type: 'stack', config: { widgets: [child] } }))
        const copy = cloneWidget(original, ids())
        expect((copy.config!.widgets as Widget[])[0]).toEqual({ ...w('c1', 4), id: 'new-2' })
    })
})

describe('duplicateInRow', () => {
    it('inserts the copy right after the original, starting where it ends', () => {
        const row = [w('a', 3, 1), w('b', 3, 10)]
        const out = duplicateInRow(row, 0, ids())
        expect(out.map((x) => x.id)).toEqual(['a', 'new-1', 'b'])
        expect(out[1]).toMatchObject({ width: 3, column: 4, title: 'a' })
        expect(out[2]).toMatchObject({ id: 'b', column: 10 })
    })

    it('shrinks the copy to the free columns', () => {
        const row = [w('a', 6, 1), w('b', 4, 7)]
        const out = duplicateInRow(row, 0, ids())
        expect(out[1]).toMatchObject({ width: 2, column: 7 })
    })

    it('pushes right-hand neighbours along and keeps the row on one line', () => {
        const row = [w('a', 4, 1), w('b', 4, 5)]
        const out = duplicateInRow(row, 0, ids())
        expect(out.map((x) => x.id)).toEqual(['a', 'new-1', 'b'])
        expect(placeRow(out).map((p) => [p.column, p.width])).toEqual([[1, 4], [5, 4], [9, 4]])
    })

    it('pulls the row left when the original sits at the right edge', () => {
        const row = [w('a', 4, 9)]
        const out = duplicateInRow(row, 0, ids())
        expect(out.map((x) => x.id)).toEqual(['a', 'new-1'])
        expect(placeRow(out).map((p) => [p.column, p.width])).toEqual([[5, 4], [9, 4]])
    })

    it('resolves a flowing (column-less) original from its placement', () => {
        const row = [w('a', 3), w('b', 3)]
        const out = duplicateInRow(row, 1, ids())
        expect(out[2]).toMatchObject({ id: 'new-1', column: 7, width: 3 })
    })

    it('leaves a full row untouched', () => {
        const row = [w('a', 6, 1), w('b', 6, 7)]
        expect(duplicateInRow(row, 0, ids())).toBe(row)
    })

    it('ignores an index outside the row', () => {
        const row = [w('a', 3, 1)]
        expect(duplicateInRow(row, 4, ids())).toBe(row)
    })
})
