import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AdminLayout from './AdminLayout.vue'

const route = { name: 'admin-background-images' }
const push = vi.fn()

vi.mock('vue-router', () => ({
    useRoute: () => route,
    useRouter: () => ({ push }),
}))

function sidebar() {
    const wrapper = mount(AdminLayout, { global: { stubs: { RouterView: true } } })
    const groups = wrapper.findAll('.admin-nav-group').map(g => ({
        label: g.find('.admin-nav-group-label').text(),
        items: g.findAll('.admin-nav-item').map(i => i.text()),
    }))
    return { wrapper, groups }
}

describe('AdminLayout sidebar', () => {
    it('groups theming apart from user content', () => {
        const { groups } = sidebar()
        expect(groups).toEqual([
            { label: 'Dashboards', items: ['Dashboards'] },
            { label: 'Theming', items: ['Themes', 'Background images', 'Widget themes'] },
            { label: 'Content', items: ['Notes', 'Images'] },
            { label: 'Documentation', items: ['Dashboards', 'Widgets', 'Theming'] },
        ])
    })

    it('highlights the background images entry and navigates by route name', async () => {
        const { wrapper } = sidebar()
        const active = wrapper.find('.admin-nav-item.active')
        expect(active.text()).toBe('Background images')
        await active.trigger('click')
        expect(push).toHaveBeenCalledWith({ name: 'admin-background-images' })
    })
})
