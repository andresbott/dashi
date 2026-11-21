import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import AdminLayout from './AdminLayout.vue'

const { route, push } = vi.hoisted(() => ({
    route: { name: 'admin-background-images', fullPath: '/admin/background-images' },
    push: vi.fn(),
}))

// The layout watches route.fullPath, so the mocked route must be reactive for
// a navigation that does not go through the sidebar to be observed.
const reactiveRoute = reactive(route)

vi.mock('vue-router', () => ({
    useRoute: () => reactiveRoute,
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

beforeEach(() => {
    push.mockClear()
    reactiveRoute.name = 'admin-background-images'
    reactiveRoute.fullPath = '/admin/background-images'
})

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
        expect(active.attributes('aria-current')).toBe('page')
        await active.trigger('click')
        expect(push).toHaveBeenCalledWith({ name: 'admin-background-images' })
    })
})

// Below the md breakpoint CSS turns the sidebar into a drawer; these tests
// pin the state the CSS keys off (.open, aria-expanded), not the media query.
describe('AdminLayout mobile drawer', () => {
    it('toggles the drawer from the topbar menu button', async () => {
        const { wrapper } = sidebar()
        const toggle = wrapper.find('.admin-nav-toggle')
        const nav = wrapper.find('#admin-nav')

        expect(toggle.attributes('aria-controls')).toBe('admin-nav')
        expect(toggle.attributes('aria-expanded')).toBe('false')
        expect(nav.classes()).not.toContain('open')
        expect(wrapper.find('.admin-nav-backdrop').exists()).toBe(false)

        await toggle.trigger('click')
        expect(toggle.attributes('aria-expanded')).toBe('true')
        expect(toggle.attributes('aria-label')).toBe('Close menu')
        expect(nav.classes()).toContain('open')
        expect(wrapper.find('.admin-nav-backdrop').exists()).toBe(true)

        await toggle.trigger('click')
        expect(nav.classes()).not.toContain('open')
    })

    it('closes when a nav entry is picked', async () => {
        const { wrapper } = sidebar()
        await wrapper.find('.admin-nav-toggle').trigger('click')

        const notes = wrapper.findAll('.admin-nav-item').find(i => i.text() === 'Notes')!
        await notes.trigger('click')

        expect(push).toHaveBeenCalledWith({ name: 'admin-notes' })
        expect(wrapper.find('#admin-nav').classes()).not.toContain('open')
    })

    it('closes on any route change, e.g. a link inside the panel', async () => {
        const { wrapper } = sidebar()
        await wrapper.find('.admin-nav-toggle').trigger('click')
        expect(wrapper.find('#admin-nav').classes()).toContain('open')

        reactiveRoute.name = 'admin-theme-edit'
        reactiveRoute.fullPath = '/admin/themes/abc123'
        await nextTick()

        expect(wrapper.find('#admin-nav').classes()).not.toContain('open')
    })

    it('closes on the backdrop and on Escape', async () => {
        const { wrapper } = sidebar()
        const toggle = wrapper.find('.admin-nav-toggle')

        await toggle.trigger('click')
        await wrapper.find('.admin-nav-backdrop').trigger('click')
        expect(wrapper.find('#admin-nav').classes()).not.toContain('open')

        await toggle.trigger('click')
        window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
        await nextTick()
        expect(wrapper.find('#admin-nav').classes()).not.toContain('open')
        wrapper.unmount()
    })
})
