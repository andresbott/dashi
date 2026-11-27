import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent, inject, ref, watch } from 'vue'
import PrimeVue from 'primevue/config'
import DashboardEditView from './DashboardEditView.vue'
import { WIDGET_MOVE } from '@/lib/injectionKeys'
import type { WidgetMove } from '@/lib/widgetMove'
import type { Dashboard, Row, Widget } from '@/types/dashboard'

const draft = ref<Dashboard | null>(null)

vi.mock('@/composables/useDashboardDraft', () => ({
    useDashboardDraft: () => ({
        draft,
        isLoading: ref(false),
        isError: ref(false),
        saveStatus: ref('idle'),
        hasUnsaved: ref(false),
        flush: vi.fn(async () => {}),
    }),
}))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: { id: 'd1' } }),
    useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('@/lib/serverInfo', () => ({ dashboardViewUrl: (id: string) => `http://viewer/${id}` }))

// Every row in the page captures the injected move handler, the way the real
// rows and stacks do.
let moveWidget: ((m: WidgetMove) => void) | undefined
const RowStub = defineComponent({
    props: { row: { type: Object, required: true } },
    setup() {
        moveWidget = inject(WIDGET_MOVE)
        return () => null
    },
})

const w = (id: string, width: number, column?: number, extra: Partial<Widget> = {}): Widget => ({
    id,
    type: 'clock',
    title: id,
    width,
    column,
    ...extra,
})

function rowsOf(): Row[] {
    return draft.value!.pages[0].rows
}

// Counts the draft's change notifications: useDashboardDraft deep-watches the
// draft with flush 'sync', so each assignment is one autosave trigger.
function countChanges() {
    let n = 0
    watch(draft, () => n++, { deep: true, flush: 'sync' })
    return () => n
}

describe('DashboardEditView widget moves', () => {
    beforeEach(() => {
        moveWidget = undefined
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        draft.value = {
            id: 'd1', name: 'Kitchen', icon: '', type: 'interactive',
            container: { maxWidth: '', verticalAlign: 'top', horizontalAlign: 'center' },
            pages: [{
                name: 'Main',
                rows: [
                    {
                        id: 'r1', height: 'auto', width: '100%',
                        widgets: [w('a', 4, 1), w('s1', 4, 5, { type: 'stack', config: { widgets: [w('c1', 4)] } })],
                    },
                    { id: 'r2', height: 'auto', width: '100%', widgets: [w('b', 3, 1)] },
                    { id: 'r3', height: 'auto', width: '100%', widgets: [w('full', 12, 1)] },
                ],
            }],
        } as Dashboard
    })

    afterEach(() => {
        vi.unstubAllGlobals()
    })

    async function mountView() {
        const wrapper = mount(DashboardEditView, {
            global: { plugins: [PrimeVue], stubs: { DashboardRow: RowStub } },
        })
        await flushPromises()
        expect(moveWidget).toBeTypeOf('function')
        return wrapper
    }

    it('moves a widget to another row with one draft assignment', async () => {
        await mountView()
        const changes = countChanges()
        moveWidget!({ source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'row', rowId: 'r1', column: 9 } })
        expect(changes()).toBe(1)
        expect(rowsOf()[0].widgets.map((x) => [x.id, x.column])).toEqual([['a', 1], ['s1', 5], ['b', 9]])
        expect(rowsOf()[1].widgets).toEqual([])
    })

    it('repositions a widget within its own row', async () => {
        await mountView()
        const changes = countChanges()
        moveWidget!({ source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'row', rowId: 'r2', column: 4 } })
        expect(changes()).toBe(1)
        expect(rowsOf()[1].widgets).toEqual([w('b', 3, 4)])
    })

    it('does not reassign the draft for a release that changes nothing', async () => {
        await mountView()
        const rows = rowsOf()
        const changes = countChanges()
        // A click-and-release on the handle: same row, same column.
        moveWidget!({ source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'row', rowId: 'r2', column: 1 } })
        // A stack child dropped back at its own index.
        moveWidget!({ source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'stack', stackId: 's1', index: 0 } })
        expect(changes()).toBe(0)
        expect(rowsOf()).toBe(rows)
    })

    it('drops a row widget into a stack with one draft assignment', async () => {
        await mountView()
        const changes = countChanges()
        moveWidget!({ source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'stack', stackId: 's1', index: 0 } })
        expect(changes()).toBe(1)
        const stack = rowsOf()[0].widgets[1]
        expect((stack.config as { widgets: Widget[] }).widgets.map((x) => [x.id, x.width, x.column])).toEqual([
            ['b', 4, undefined],
            ['c1', 4, undefined],
        ])
        expect(rowsOf()[1].widgets).toEqual([])
    })

    it('drags a stack child out onto a row', async () => {
        await mountView()
        const changes = countChanges()
        moveWidget!({ source: { kind: 'stack', stackId: 's1', widgetId: 'c1' }, target: { kind: 'row', rowId: 'r2', column: 4 } })
        expect(changes()).toBe(1)
        expect(rowsOf()[1].widgets.map((x) => [x.id, x.width, x.column])).toEqual([['b', 3, 1], ['c1', 6, 4]])
        expect((rowsOf()[0].widgets[1].config as { widgets: Widget[] }).widgets).toEqual([])
    })

    it('leaves the draft unchanged when the move is refused', async () => {
        await mountView()
        const before = JSON.parse(JSON.stringify(draft.value))
        const changes = countChanges()
        // No room on the full row.
        moveWidget!({ source: { kind: 'row', rowId: 'r2', widgetId: 'b' }, target: { kind: 'row', rowId: 'r3', column: 1 } })
        // A stack never goes into a stack.
        moveWidget!({ source: { kind: 'row', rowId: 'r1', widgetId: 's1' }, target: { kind: 'stack', stackId: 's1', index: 0 } })
        expect(changes()).toBe(0)
        expect(JSON.parse(JSON.stringify(draft.value))).toEqual(before)
    })
})
