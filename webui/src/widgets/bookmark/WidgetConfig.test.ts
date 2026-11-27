import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import WidgetConfig from './WidgetConfig.vue'
import type { BookmarkWidgetConfig } from './types'

vi.mock('@/composables/useDashboards', () => ({
    useDashboardAssets: () => ({ data: ref([]) }),
}))

const base: BookmarkWidgetConfig = {
    url: 'https://example.com',
    icon: 'ti-bookmark',
    title: 'Example',
    subtitle: '',
}

function mountForm(config: BookmarkWidgetConfig | null = base) {
    return mount(WidgetConfig, {
        props: { config },
        global: { plugins: [PrimeVue], stubs: { IconSelector: true } },
    })
}

// The URL input is the one with the URL placeholder.
const urlInput = (wrapper: ReturnType<typeof mountForm>) => wrapper.find('input[placeholder="https://example.com"]')

describe('bookmark WidgetConfig URL validation', () => {
    // PrimeVue's Select reads matchMedia on mount; jsdom has none.
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
    })
    afterEach(() => vi.unstubAllGlobals())

    it.each([
        'javascript:alert(1)',
        'data:text/html,x',
        '//evil.com',
        'example.com',
    ])('shows an inline error and emits nothing for %s', async (url) => {
        const wrapper = mountForm()
        await urlInput(wrapper).setValue(url)

        expect(wrapper.find('.bookmark-url-error').exists()).toBe(true)
        expect(wrapper.find('.bookmark-url-error').text()).toContain('http://, https:// or /')
        expect(wrapper.emitted('update:config')).toBeUndefined()
    })

    it.each(['https://example.org', 'http://example.org', '/abc123'])('accepts and emits %s', async (url) => {
        const wrapper = mountForm()
        await urlInput(wrapper).setValue(url)

        expect(wrapper.find('.bookmark-url-error').exists()).toBe(false)
        const emitted = wrapper.emitted('update:config')
        expect(emitted).toBeDefined()
        expect((emitted!.at(-1)![0] as BookmarkWidgetConfig).url).toBe(url)
    })

    it('flags an unsafe URL already stored in the dashboard', () => {
        const wrapper = mountForm({ ...base, url: 'javascript:alert(1)' })
        expect(wrapper.find('.bookmark-url-error').exists()).toBe(true)
    })

    it('does not flag an empty URL on a new bookmark', () => {
        const wrapper = mountForm(null)
        expect(wrapper.find('.bookmark-url-error').exists()).toBe(false)
    })
})

describe('bookmark WidgetConfig tabler icon', () => {
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
    })
    afterEach(() => vi.unstubAllGlobals())

    // The icon selector's trigger shows the icon name itself.
    it('hands the icon to the selector without repeating its class name beside it', () => {
        const wrapper = mountForm()
        expect(wrapper.findComponent({ name: 'IconSelector' }).props('modelValue')).toBe('ti-bookmark')
        expect(wrapper.text()).not.toContain('ti-bookmark')
    })

    it('labels each field uniquely and names the selector by its visible label', () => {
        const wrapper = mountForm()
        const labels = wrapper.findAll('label').map(l => l.text())
        expect(new Set(labels).size).toBe(labels.length)
        expect(labels).toContain('Icon source')

        const iconLabel = wrapper.findAll('label').find(l => l.text() === 'Icon')!
        expect(iconLabel.attributes('id')).toBeTruthy()
        expect(wrapper.findComponent({ name: 'IconSelector' }).props('labelledBy')).toBe(iconLabel.attributes('id'))
    })
})

describe('bookmark WidgetConfig new tab option', () => {
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
    })
    afterEach(() => vi.unstubAllGlobals())

    it('starts off for a bookmark saved without it and emits it when ticked', async () => {
        const wrapper = mountForm()
        const box = wrapper.find('#bookmarkNewTab')
        expect((box.element as HTMLInputElement).checked).toBe(false)

        await box.setValue(true)
        const updates = wrapper.emitted('update:config')!
        expect((updates[updates.length - 1][0] as BookmarkWidgetConfig).newTab).toBe(true)
    })

    it('shows a stored newTab', () => {
        const wrapper = mountForm({ ...base, newTab: true })
        expect((wrapper.find('#bookmarkNewTab').element as HTMLInputElement).checked).toBe(true)
    })
})
