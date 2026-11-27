import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import DashboardWidget from './DashboardWidget.vue'
import type { Widget } from '@/types/dashboard'

// The card's header controls are under test, not the preview or the dialogs:
// the renderer and settings dialog are stubbed, and the confirm prompt is a
// stub that shows whether it is open and lets the test press Ok.
const ConfirmStub = {
    name: 'ConfirmDialog',
    props: ['visible', 'name', 'title', 'message'],
    emits: ['confirm', 'update:visible'],
    template: '<div v-if="visible" class="confirm-stub"><button class="confirm-ok" @click="$emit(\'confirm\')">Ok</button></div>',
}

function mountCard(props: Record<string, unknown> = {}) {
    const widget: Widget = { id: 'w1', type: 'clock', title: 'Clock', width: 4, column: 1 }
    return mount(DashboardWidget, {
        props: { widget, ...props },
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
            stubs: { WidgetRenderer: true, WidgetSettingsDialog: true, ConfirmDialog: ConfirmStub },
        },
    })
}

describe('DashboardWidget header', () => {
    it('has a delete button next to the edit button', () => {
        const wrapper = mountCard()
        const header = wrapper.find('.widget-header')
        expect(header.find('[aria-label="Widget settings"]').exists()).toBe(true)
        expect(header.find('[aria-label="Delete widget"]').exists()).toBe(true)
    })

    it('asks for confirmation before deleting', async () => {
        const wrapper = mountCard()
        await wrapper.find('[aria-label="Delete widget"]').trigger('click')
        expect(wrapper.find('.confirm-stub').exists()).toBe(true)
        expect(wrapper.emitted('delete')).toBeUndefined()

        await wrapper.find('.confirm-ok').trigger('click')
        expect(wrapper.emitted('delete')).toHaveLength(1)
        expect(wrapper.find('.confirm-stub').exists()).toBe(false)
    })

    it('emits duplicate from the duplicate button', async () => {
        const wrapper = mountCard()
        await wrapper.find('.widget-header [aria-label="Duplicate widget"]').trigger('click')
        expect(wrapper.emitted('duplicate')).toHaveLength(1)
    })

    it('disables duplicate when the row is full', () => {
        const wrapper = mountCard({ rowFull: true })
        expect(wrapper.find('[aria-label="Duplicate widget"]').attributes('disabled')).toBeDefined()
    })
})
