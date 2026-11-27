import { describe, it, expect, afterEach, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, h, nextTick } from 'vue'
import PrimeVue from 'primevue/config'
import draggable from 'vuedraggable'
import StackWidget from './Widget.vue'
import { EDITING_MODE, WIDGET_MOVE } from '@/lib/injectionKeys'
import { dragState, endWidgetDrag, startWidgetDrag } from '@/lib/widgetDrag'
import type { WidgetMove } from '@/lib/widgetMove'
import type { Widget } from '@/types/dashboard'

const child = (id: string): Widget => ({ id, type: 'clock', title: id, width: 4 })
const stack = (children: Widget[]): Widget => ({
    id: 's1',
    type: 'stack',
    title: 'Stack',
    width: 4,
    column: 5,
    config: { widgets: children },
})

// Fixes an element's layout box, since jsdom does no layout.
function setBox(el: Element, left: number, top: number, width: number, height: number) {
    ;(el as HTMLElement).getBoundingClientRect = () =>
        ({ left, top, width, height, right: left + width, bottom: top + height, x: left, y: top }) as DOMRect
}

// The stack as the editor renders it: inside its card's .widget-body.
function mountStack(
    children: Widget[],
    opts: { editing?: boolean; moveWidget?: ((m: WidgetMove) => void) | null } = {},
) {
    const { editing = true, moveWidget = vi.fn() } = opts
    const onUpdate = vi.fn()
    const Card = defineComponent({
        setup: () => () =>
            h('div', { class: 'widget-body' }, [h(StackWidget, { widget: stack(children), 'onUpdate:widget': onUpdate })]),
    })
    const provide: Record<symbol, unknown> = { [EDITING_MODE as symbol]: editing }
    if (moveWidget) provide[WIDGET_MOVE as symbol] = moveWidget
    const wrapper = mount(Card, {
        attachTo: document.body,
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
            provide,
            stubs: { WidgetRenderer: true, AddWidgetDialog: true, WidgetSettingsDialog: true, ConfirmDialog: true },
        },
    })
    const body = wrapper.find('.widget-body').element
    // Card body x 0..400, y 0..300; children cards 40px tall from y 10, 60, 110.
    setBox(body, 0, 0, 400, 300)
    wrapper.findAll('.stack-child-editor').forEach((c, i) => setBox(c.element, 10, 10 + i * 50, 380, 40))
    mounted.push(wrapper)
    return { wrapper, onUpdate, moveWidget }
}

// A widget from some row dragged to (x, y).
const dragRowWidgetTo = (x: number, y: number, draggedType = 'clock') =>
    startWidgetDrag(
        { kind: 'row', rowId: 'r9', widgetId: 'w9' },
        { originColumn: 1, originLeftX: x, grabX: x, grabY: y, span: 3, draggedType },
    )

const list = (wrapper: VueWrapper) => wrapper.findComponent(draggable)

const mounted: VueWrapper[] = []

afterEach(() => {
    endWidgetDrag()
    while (mounted.length) mounted.pop()!.unmount()
})

describe('stack widget drop zone', () => {
    it("registers the card's whole .widget-body as its zone in edit mode", () => {
        mountStack([child('c1')])
        // Inside the body's padding, outside the stack's own (zero) box.
        dragRowWidgetTo(5, 290)
        expect(dragState).toMatchObject({ targetKind: 'stack', targetStackId: 's1', targetIndex: 1 })
    })

    it('unregisters the zone on unmount', () => {
        const { wrapper } = mountStack([child('c1')])
        mounted.splice(mounted.indexOf(wrapper), 1)
        wrapper.unmount()
        dragRowWidgetTo(5, 290)
        expect(dragState.targetKind).toBeNull()
    })

    it('registers no zone in view mode', () => {
        mountStack([child('c1')], { editing: false })
        dragRowWidgetTo(5, 290)
        expect(dragState.targetKind).toBeNull()
    })

    it('highlights and shows the insert marker before the hovered child', async () => {
        const { wrapper } = mountStack([child('c1'), child('c2'), child('c3')])
        expect(wrapper.find('.stack-insert-marker').exists()).toBe(false)
        dragRowWidgetTo(100, 55) // between c1 (mid 30) and c2 (mid 80)
        await nextTick()
        expect(wrapper.find('.widget-stack-editor--target').exists()).toBe(true)
        const items = wrapper.findAll('.stack-item')
        expect(wrapper.findAll('.stack-insert-marker')).toHaveLength(1)
        expect(items[1].find('.stack-insert-marker').exists()).toBe(true)
    })

    it('shows the marker after the last child', async () => {
        const { wrapper } = mountStack([child('c1'), child('c2')])
        dragRowWidgetTo(100, 280)
        await nextTick()
        expect(wrapper.findAll('.stack-insert-marker')).toHaveLength(1)
        expect(wrapper.find('.stack-end-marker .stack-insert-marker').exists()).toBe(true)
    })

    it('shows a dashed placeholder with a marker when an empty stack is hovered', async () => {
        const { wrapper } = mountStack([])
        expect(wrapper.find('.stack-drop-placeholder').exists()).toBe(false)
        dragRowWidgetTo(100, 100)
        await nextTick()
        expect(wrapper.find('.stack-drop-placeholder .stack-insert-marker').exists()).toBe(true)
    })

    it('ignores a dragged stack', async () => {
        const { wrapper } = mountStack([child('c1')])
        dragRowWidgetTo(100, 100, 'stack')
        await nextTick()
        expect(dragState.targetKind).toBeNull()
        expect(wrapper.find('.widget-stack-editor--target').exists()).toBe(false)
    })
})

describe('stack child drag', () => {
    const startChildDrag = async (wrapper: VueWrapper, index: number, x: number, y: number) => {
        await wrapper.findAll('.stack-drag-handle')[index].trigger('pointerdown', { clientX: x, clientY: y })
        list(wrapper).vm.$emit('start', { oldIndex: index, item: wrapper.findAll('.stack-item')[index].element })
        await nextTick()
    }

    it('starts a stack-sourced drag at the grab point', async () => {
        const { wrapper } = mountStack([child('c1'), child('c2')])
        // The pointer-down point, not the item's box (zero in jsdom), seeds
        // the first target: below c2, so after it.
        await startChildDrag(wrapper, 0, 20, 290)
        expect(dragState).toMatchObject({
            active: true,
            sourceKind: 'stack',
            sourceStackId: 's1',
            sourceWidgetId: 'c1',
            draggedType: 'clock',
            targetKind: 'stack',
            targetIndex: 1,
        })
    })

    it('places the marker at the visual position when reordering within the stack', async () => {
        const { wrapper } = mountStack([child('c1'), child('c2'), child('c3')])
        // Grab c1 and hover between c2 and c3: index 1 after removing c1, so
        // the marker sits before c3.
        await startChildDrag(wrapper, 0, 20, 30)
        document.dispatchEvent(new MouseEvent('pointermove', { clientX: 20, clientY: 105 }))
        await nextTick()
        expect(dragState.targetIndex).toBe(1)
        expect(wrapper.findAll('.stack-item')[2].find('.stack-insert-marker').exists()).toBe(true)
        expect(wrapper.findAll('.stack-insert-marker')).toHaveLength(1)
    })

    it('hands the resolved move to WIDGET_MOVE on release, without emitting an update', async () => {
        const { wrapper, moveWidget, onUpdate } = mountStack([child('c1'), child('c2'), child('c3')])
        await startChildDrag(wrapper, 0, 20, 30)
        document.dispatchEvent(new MouseEvent('pointermove', { clientX: 20, clientY: 105 }))
        list(wrapper).vm.$emit('end', { oldIndex: 0, newIndex: 0 })
        expect(moveWidget).toHaveBeenCalledExactlyOnceWith({
            source: { kind: 'stack', stackId: 's1', widgetId: 'c1' },
            target: { kind: 'stack', stackId: 's1', index: 1 },
        })
        expect(onUpdate).not.toHaveBeenCalled()
        expect(dragState.active).toBe(false)
    })

    it('does nothing on release without a WIDGET_MOVE provider', async () => {
        const { wrapper, onUpdate } = mountStack([child('c1'), child('c2')], { moveWidget: null })
        await startChildDrag(wrapper, 0, 20, 30)
        expect(() => list(wrapper).vm.$emit('end', { oldIndex: 0, newIndex: 0 })).not.toThrow()
        expect(onUpdate).not.toHaveBeenCalled()
        expect(dragState.active).toBe(false)
    })

    it('aborts its drag when the stack unmounts mid-gesture', async () => {
        const { wrapper } = mountStack([child('c1')])
        await startChildDrag(wrapper, 0, 20, 30)
        mounted.splice(mounted.indexOf(wrapper), 1)
        wrapper.unmount()
        expect(dragState.active).toBe(false)
    })

    it('keeps SortableJS passive: it never reorders the list itself', () => {
        const { wrapper } = mountStack([child('c1'), child('c2')])
        // vuedraggable declares move as a prop; sort is passed to SortableJS.
        expect((list(wrapper).props('move') as () => boolean)()).toBe(false)
        expect(list(wrapper).vm.$attrs.sort).toBe(false)
        expect(list(wrapper).vm.$attrs['onUpdate:modelValue']).toBeUndefined()
    })

    it('tells the user children can be dragged in and out', () => {
        const { wrapper } = mountStack([child('c1')])
        expect(wrapper.find('.stack-hint').text()).toBe('Drag widgets in or out of the stack.')
        expect(wrapper.find('.stack-drag-handle').attributes('aria-label')).toMatch(/out of the stack/)
    })
})
