import { describe, it, expect } from 'vitest'
import { reactive } from 'vue'
import type { Row, Widget } from '@/types/dashboard'
import { applyWidgetMove, dragOutSpan, findStack, type WidgetMove } from './widgetMove'
import { plainCopy } from './plainCopy'

const w = (id: string, width: number, column?: number, extra: Partial<Widget> = {}): Widget => ({
    id,
    type: 'clock',
    title: id,
    width,
    ...(column === undefined ? {} : { column }),
    ...extra,
})

const stack = (id: string, width: number, column: number | undefined, children: Widget[]): Widget =>
    w(id, width, column, { type: 'stack', config: { widgets: children } })

const row = (id: string, widgets: Widget[]): Row => ({ id, height: 'auto', width: '100%', widgets })

const children = (rows: Row[], stackId: string): Widget[] => {
    const at = findStack(rows, stackId)
    if (!at) throw new Error(`no stack ${stackId}`)
    return (rows[at.rowIndex].widgets[at.widgetIndex].config as { widgets: Widget[] }).widgets
}

// A page with two rows: r1 holds a plain widget and a stack with two
// children, r2 holds a single widget and has room to spare.
const page = (): Row[] => [
    row('r1', [w('a', 4, 1), stack('s1', 4, 5, [w('c1', 4), w('c2', 4)])]),
    row('r2', [w('b', 3, 1)]),
]

describe('dragOutSpan', () => {
    it('is 6 on a row with plenty of room', () => {
        expect(dragOutSpan([w('a', 3)])).toBe(6)
    })

    it('shrinks to the free columns', () => {
        expect(dragOutSpan([w('a', 4), w('b', 4)])).toBe(4)
    })

    it('is below 1 on a full row', () => {
        expect(dragOutSpan([w('a', 6), w('b', 6)])).toBeLessThan(1)
    })
})

describe('findStack', () => {
    it('locates a stack by id', () => {
        expect(findStack(page(), 's1')).toEqual({ rowIndex: 0, widgetIndex: 1 })
    })

    it('ignores non-stack widgets and unknown ids', () => {
        expect(findStack(page(), 'a')).toBeNull()
        expect(findStack(page(), 'nope')).toBeNull()
    })
})

describe('applyWidgetMove', () => {
    const cases: { name: string; rows?: () => Row[]; move: WidgetMove; check: (out: Row[]) => void }[] = [
        {
            name: 'row → same row repositions within the line',
            move: { source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'row', rowId: 'r2', column: 5 } },
            check: (out) => {
                expect(out[1].widgets.map((x) => x.id)).toEqual(['b', 's2'])
                expect(out[1].widgets[0]).toEqual(w('b', 3, 5))
            },
        },
        {
            name: 'row → other row inserts at the column and keeps the width',
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'row', rowId: 'r2', column: 6 } },
            check: (out) => {
                expect(out[0].widgets.map((x) => x.id)).toEqual(['s1'])
                expect(out[1].widgets.map((x) => x.id)).toEqual(['b', 'a', 's2'])
                expect(out[1].widgets[1]).toEqual(w('a', 4, 6))
            },
        },
        {
            name: 'row → stack inserts at the index with the stack width and no column',
            move: { source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'stack', stackId: 's1', index: 1 } },
            check: (out) => {
                expect(out[1].widgets.map((x) => x.id)).toEqual(['s2'])
                expect(children(out, 's1')).toEqual([w('c1', 4), w('b', 4), w('c2', 4)])
            },
        },
        {
            name: 'row → stack in the same row',
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'stack', stackId: 's1', index: 0 } },
            check: (out) => {
                expect(out[0].widgets.map((x) => x.id)).toEqual(['s1'])
                expect(children(out, 's1').map((x) => x.id)).toEqual(['a', 'c1', 'c2'])
            },
        },
        {
            name: 'stack → other row takes the drag-out span at the column',
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r2', column: 4 } },
            check: (out) => {
                expect(children(out, 's1').map((x) => x.id)).toEqual(['c2'])
                expect(out[1].widgets.map((x) => x.id)).toEqual(['b', 'c1', 's2'])
                expect(out[1].widgets[1]).toEqual(w('c1', 6, 4))
            },
        },
        {
            name: 'stack → its own row shrinks to the free columns',
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c2' }, target: { kind: 'row', rowId: 'r1', column: 9 } },
            check: (out) => {
                expect(out[0].widgets.map((x) => x.id)).toEqual(['a', 's1', 'c2'])
                expect(out[0].widgets[2]).toEqual(w('c2', 4, 9))
                expect(children(out, 's1').map((x) => x.id)).toEqual(['c1'])
            },
        },
        {
            name: 'stack → row clamps the column so the span stays on the grid',
            rows: page,
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r2', column: 12 } },
            check: (out) => expect(out[1].widgets[1]).toEqual(w('c1', 6, 7)),
        },
        {
            name: 'stack → same stack reorders (index counted after removal)',
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'stack', stackId: 's1', index: 1 } },
            check: (out) => expect(children(out, 's1').map((x) => x.id)).toEqual(['c2', 'c1']),
        },
        {
            name: 'stack → other stack takes the target stack width',
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c2' }, target: { kind: 'stack', stackId: 's2', index: 0 } },
            check: (out) => {
                expect(children(out, 's1').map((x) => x.id)).toEqual(['c1'])
                expect(children(out, 's2')).toEqual([w('c2', 3), w('d1', 3)])
            },
        },
        {
            name: 'an index past the end clamps to append',
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'stack', stackId: 's1', index: 99 } },
            check: (out) => expect(children(out, 's1').map((x) => x.id)).toEqual(['c1', 'c2', 'a']),
        },
    ]

    const withSecondStack = (): Row[] => {
        const rows = page()
        rows[1].widgets.push(stack('s2', 3, 7, [w('d1', 3)]))
        return rows
    }

    for (const c of cases) {
        it(c.name, () => {
            const out = applyWidgetMove((c.rows ?? withSecondStack)(), c.move)
            expect(out).not.toBeNull()
            c.check(out!)
        })
    }

    const refusals: { name: string; rows: () => Row[]; move: WidgetMove }[] = [
        {
            name: 'a row widget into a full row',
            rows: () => [row('r1', [w('a', 4, 1)]), row('r2', [w('x', 6, 1), w('y', 6, 7)])],
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
        {
            name: 'a row widget wider than the free columns',
            rows: () => [row('r1', [w('a', 4, 1)]), row('r2', [w('x', 9, 1)])],
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'row', rowId: 'r2', column: 10 } },
        },
        {
            name: 'a stack into a stack',
            rows: () => [row('r1', [stack('s1', 4, 1, []), stack('s2', 4, 5, [])])],
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 's2' }, target: { kind: 'stack', stackId: 's1', index: 0 } },
        },
        {
            name: 'a stack child into a full row',
            rows: () => [row('r1', [stack('s1', 6, 1, [w('c1', 6)])]), row('r2', [w('x', 12, 1)])],
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
        {
            name: 'a stack child into its own row when only the stack is left of room',
            rows: () => [row('r1', [w('a', 8, 1), stack('s1', 4, 9, [w('c1', 4)])])],
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r1', column: 1 } },
        },
        {
            name: 'an unknown source row',
            rows: page,
            move: { source: { kind: 'row', rowId: 'nope', widgetId: 'a' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
        {
            name: 'an unknown source widget',
            rows: page,
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'nope' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
        {
            name: 'an unknown target row',
            rows: page,
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'row', rowId: 'nope', column: 1 } },
        },
        {
            name: 'an unknown target stack',
            rows: page,
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 'a' }, target: { kind: 'stack', stackId: 'nope', index: 0 } },
        },
        {
            name: 'a target stack id that is not a stack',
            rows: page,
            move: { source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'stack', stackId: 'a', index: 0 } },
        },
        {
            name: 'an unknown source stack',
            rows: page,
            move: { source: { kind: 'stack', stackId: 'nope', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
        {
            name: 'an unknown stack child',
            rows: page,
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'nope' }, target: { kind: 'row', rowId: 'r2', column: 1 } },
        },
    ]

    for (const c of refusals) {
        it(`refuses ${c.name}`, () => {
            expect(applyWidgetMove(c.rows(), c.move)).toBeNull()
        })
    }

    const noOps: { name: string; move: WidgetMove }[] = [
        {
            name: 'a stack child dropped back at its own index',
            move: { source: { kind: 'stack', stackId: 's1', widgetId: 'c2' }, target: { kind: 'stack', stackId: 's1', index: 1 } },
        },
        {
            name: 'a row widget dropped back at its own column',
            move: { source: { kind: 'row', rowId: 'r1', widgetId: 's1' }, target: { kind: 'row', rowId: 'r1', column: 5 } },
        },
    ]

    for (const c of noOps) {
        it(`has nothing to do for ${c.name}`, () => {
            expect(applyWidgetMove(withSecondStack(), c.move)).toBeNull()
        })
    }

    it('still applies a same-column drop that clamps a legacy overflow row', () => {
        // y overflows the line (columns 9..14); the drop pulls it back to 7.
        const rows = [row('r1', [w('x', 6, 1), w('y', 6, 9)])]
        const out = applyWidgetMove(rows, {
            source: { kind: 'row', rowId: 'r1', widgetId: 'x' },
            target: { kind: 'row', rowId: 'r1', column: 1 },
        })
        expect(out).not.toBeNull()
        expect(out![0].widgets).toEqual([w('x', 6, 1), w('y', 6, 7)])
    })

    it('never mutates its input', () => {
        const moves: WidgetMove[] = [
            ...cases.map((c) => c.move),
            ...refusals.map((c) => c.move),
        ]
        for (const move of moves) {
            const rows = withSecondStack()
            const before = plainCopy(rows)
            applyWidgetMove(rows, move)
            expect(rows).toEqual(before)
        }
    })

    it('leaves untouched rows as they were', () => {
        const rows = withSecondStack()
        const out = applyWidgetMove(rows, {
            source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
            target: { kind: 'stack', stackId: 's1', index: 1 },
        })
        expect(out![1]).toBe(rows[1])
    })

    it('works on reactive rows', () => {
        // The editor passes the reactive draft straight in.
        const rows = reactive(withSecondStack())
        const before = plainCopy(rows)
        const out = applyWidgetMove(rows, {
            source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
            target: { kind: 'row', rowId: 'r2', column: 4 },
        })
        expect(out![1].widgets.map((x) => x.id)).toEqual(['b', 'c1', 's2'])
        expect(children(out!, 's1').map((x) => x.id)).toEqual(['c2'])
        expect(plainCopy(rows)).toEqual(before)
    })
})
