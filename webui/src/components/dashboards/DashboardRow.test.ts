import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import DashboardRow from './DashboardRow.vue'
import type { Row } from '@/types/dashboard'

// Only the Row Settings dialog is under test. The dialog is stubbed to render
// its slot inline (PrimeVue teleports it otherwise); widgets and the add
// dialog are not involved.
const DialogStub = {
    name: 'Dialog',
    props: ['visible'],
    template: '<div v-if="visible" class="dialog-stub"><slot /></div>',
}

function mountRow(row: Partial<Row> = {}) {
    const full: Row = { id: 'r1', height: 'auto', width: '100%', widgets: [], ...row }
    return mount(DashboardRow, {
        props: { row: full, isFirst: true, isLast: true },
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
            stubs: { Dialog: DialogStub, AddWidgetDialog: true, ConfirmDialog: true, DashboardWidget: true },
        },
    })
}

describe('DashboardRow settings', () => {
    it('edits and saves the row margins', async () => {
        const wrapper = mountRow()
        await wrapper.find('[aria-label="Row settings"]').trigger('click')

        await wrapper.find('#row-margin-top').setValue('2rem')
        await wrapper.find('#row-margin-bottom').setValue(' 40px ')
        await wrapper.findAll('button').find((b) => b.text() === 'Save')!.trigger('click')

        const updated = wrapper.emitted('update')![0][0] as Row
        expect(updated.marginTop).toBe('2rem')
        expect(updated.marginBottom).toBe('40px')
    })

    it('shows stored margins and drops cleared ones', async () => {
        const wrapper = mountRow({ marginTop: '1em', marginBottom: '3rem' })
        await wrapper.find('[aria-label="Row settings"]').trigger('click')

        const top = wrapper.find('#row-margin-top').element as HTMLInputElement
        expect(top.value).toBe('1em')
        await wrapper.find('#row-margin-bottom').setValue('')
        await wrapper.findAll('button').find((b) => b.text() === 'Save')!.trigger('click')

        const updated = wrapper.emitted('update')![0][0] as Row
        expect(updated.marginTop).toBe('1em')
        expect(updated.marginBottom).toBeUndefined()
    })
})
