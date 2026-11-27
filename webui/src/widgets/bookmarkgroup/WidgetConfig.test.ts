import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import Slider from 'primevue/slider'
import Select from 'primevue/select'
import WidgetConfig from './WidgetConfig.vue'
import type { BookmarkGroupWidgetConfig } from './types'

vi.mock('@/composables/useBookmarkFiles', () => ({
    useBookmarkFiles: () => ({ files: ref(['links.yaml', 'tools.yml']) }),
}))

vi.mock('@/composables/useDashboards', () => ({
    useDashboardAssets: () => ({ data: ref([]) }),
}))

const base: BookmarkGroupWidgetConfig = {
    bookmarks: [
        { url: 'https://a.example', icon: 'ti-bookmark', title: 'A', subtitle: '' },
        { url: 'https://b.example', icon: 'ti-bookmark', title: 'B', subtitle: '' },
        { url: 'https://c.example', icon: 'ti-bookmark', title: 'C', subtitle: '' },
    ],
}

function mountForm(config: BookmarkGroupWidgetConfig | null = base) {
    return mount(WidgetConfig, {
        props: { config },
        global: { plugins: [PrimeVue], stubs: { IconSelector: true } },
    })
}

type Wrapper = ReturnType<typeof mountForm>

const lastEmit = (wrapper: Wrapper) => {
    const updates = wrapper.emitted('update:config')
    expect(updates).toBeDefined()
    return updates!.at(-1)![0] as BookmarkGroupWidgetConfig
}
const titles = (cfg: BookmarkGroupWidgetConfig) => cfg.bookmarks.map(b => b.title)
const items = (wrapper: Wrapper) => wrapper.findAll('.bookmarkgroup-item')

describe('bookmarkgroup WidgetConfig', () => {
    // PrimeVue's Select reads matchMedia on mount; jsdom has none.
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
    })
    afterEach(() => vi.unstubAllGlobals())

    it('lists every stored bookmark by title, collapsed', () => {
        const wrapper = mountForm()
        expect(items(wrapper).map(i => i.find('.bookmarkgroup-item-title').text())).toEqual(['A', 'B', 'C'])
        expect(wrapper.find('input[placeholder="https://example.com"]').exists()).toBe(false)
    })

    it('starts a new group with no bookmarks', () => {
        expect(items(mountForm(null))).toHaveLength(0)
    })

    it('adds a bookmark at the end and opens it for editing', async () => {
        const wrapper = mountForm()
        await wrapper.find('.bookmarkgroup-add').trigger('click')

        expect(items(wrapper)).toHaveLength(4)
        const added = lastEmit(wrapper).bookmarks.at(-1)!
        expect(added).toEqual({ url: '', icon: 'ti-bookmark', title: '', subtitle: '' })
        expect(items(wrapper)[3].find('input[placeholder="https://example.com"]').exists()).toBe(true)
    })

    it('adds any number of bookmarks', async () => {
        const wrapper = mountForm(null)
        for (let i = 0; i < 5; i++) await wrapper.find('.bookmarkgroup-add').trigger('click')
        expect(lastEmit(wrapper).bookmarks).toHaveLength(5)
    })

    it('removes a bookmark', async () => {
        const wrapper = mountForm()
        await items(wrapper)[1].find('.bookmarkgroup-remove').trigger('click')
        expect(titles(lastEmit(wrapper))).toEqual(['A', 'C'])
        expect(items(wrapper)).toHaveLength(2)
    })

    it('moves a bookmark up and down', async () => {
        const wrapper = mountForm()
        await items(wrapper)[2].find('.bookmarkgroup-up').trigger('click')
        expect(titles(lastEmit(wrapper))).toEqual(['A', 'C', 'B'])

        await items(wrapper)[0].find('.bookmarkgroup-down').trigger('click')
        expect(titles(lastEmit(wrapper))).toEqual(['C', 'A', 'B'])
    })

    it('disables moving past either end', () => {
        const wrapper = mountForm()
        expect(items(wrapper)[0].find('.bookmarkgroup-up').attributes('disabled')).toBeDefined()
        expect(items(wrapper)[2].find('.bookmarkgroup-down').attributes('disabled')).toBeDefined()
        expect(items(wrapper)[1].find('.bookmarkgroup-up').attributes('disabled')).toBeUndefined()
    })

    it('edits the fields of an opened bookmark', async () => {
        const wrapper = mountForm()
        await items(wrapper)[1].find('.bookmarkgroup-toggle').trigger('click')
        const item = items(wrapper)[1]
        await item.find('input[placeholder="Bookmark title"]').setValue('Beta')
        await item.find('input[placeholder="Short description"]').setValue('second')
        await item.find('input[placeholder="https://example.com"]').setValue('/beta')

        expect(lastEmit(wrapper).bookmarks[1]).toEqual({ url: '/beta', icon: 'ti-bookmark', title: 'Beta', subtitle: 'second' })
        expect(items(wrapper)[1].find('.bookmarkgroup-item-title').text()).toBe('Beta')
    })

    it('blocks saving while any bookmark URL is unsafe', async () => {
        const wrapper = mountForm()
        await items(wrapper)[0].find('.bookmarkgroup-toggle').trigger('click')
        await items(wrapper)[0].find('input[placeholder="https://example.com"]').setValue('javascript:alert(1)')

        expect(wrapper.find('.bookmark-url-error').text()).toContain('http://, https:// or /')
        expect(wrapper.emitted('update:config')).toBeUndefined()

        // Other edits are held back too until the URL is fixed.
        await wrapper.find('#bookmarkgroupNewTab').setValue(true)
        expect(wrapper.emitted('update:config')).toBeUndefined()
    })

    it('marks a collapsed bookmark whose stored URL is unsafe', () => {
        const wrapper = mountForm({ bookmarks: [{ url: '//evil.example', icon: '', title: 'X', subtitle: '' }] })
        expect(items(wrapper)[0].classes()).toContain('bookmarkgroup-item--invalid')
    })

    it('emits the shared options', async () => {
        const wrapper = mountForm()
        await wrapper.find('#bookmarkgroupTextBelow').setValue(true)
        expect(lastEmit(wrapper).textBelow).toBe(true)

        await wrapper.find('#bookmarkgroupNewTab').setValue(true)
        expect(lastEmit(wrapper).newTab).toBe(true)

        const vertical = wrapper.findAll('.bookmarkgroup-direction button').find(b => b.text() === 'Vertical')!
        await vertical.trigger('click')
        expect(lastEmit(wrapper).direction).toBe('vertical')
        expect(titles(lastEmit(wrapper))).toEqual(['A', 'B', 'C'])
    })

    it('shows stored shared options, defaulting to horizontal', () => {
        const stored = mountForm({ ...base, textBelow: true, newTab: true, direction: 'vertical' })
        expect((stored.find('#bookmarkgroupTextBelow').element as HTMLInputElement).checked).toBe(true)
        expect((stored.find('#bookmarkgroupNewTab').element as HTMLInputElement).checked).toBe(true)
        expect(stored.find('.bookmarkgroup-direction [aria-pressed="true"]').text()).toBe('Vertical')

        expect(mountForm().find('.bookmarkgroup-direction [aria-pressed="true"]').text()).toBe('Horizontal')
    })

    it('defaults to bookmarks kept in the dashboard', () => {
        const wrapper = mountForm()
        expect(wrapper.find('.bookmarkgroup-source [aria-pressed="true"]').text()).toBe('In dashboard')
        expect(items(wrapper)).toHaveLength(3)
        expect(wrapper.find('#bookmarkgroupFile').exists()).toBe(false)
    })

    it('switches to a file: hides the list, offers the files, keeps the own bookmarks', async () => {
        const wrapper = mountForm()
        const fromFile = wrapper.findAll('.bookmarkgroup-source button').find(b => b.text() === 'From file')!
        await fromFile.trigger('click')
        expect(lastEmit(wrapper).source).toBe('file')
        expect(items(wrapper)).toHaveLength(0)

        const select = wrapper.findAllComponents(Select).find(c => c.classes().includes('bookmarkgroup-file'))!
        expect(select.props('options')).toEqual(['links.yaml', 'tools.yml'])
        select.vm.$emit('update:modelValue', 'tools.yml')
        await wrapper.vm.$nextTick()
        const emitted = lastEmit(wrapper)
        expect(emitted.file).toBe('tools.yml')
        expect(titles(emitted)).toEqual(['A', 'B', 'C'])
    })

    it('still offers a stored file that no longer exists', () => {
        const wrapper = mountForm({ ...base, source: 'file', file: 'gone.yaml' })
        const select = wrapper.findAllComponents(Select).find(c => c.classes().includes('bookmarkgroup-file'))!
        expect(select.props('modelValue')).toBe('gone.yaml')
        expect(select.props('options')).toEqual(['gone.yaml', 'links.yaml', 'tools.yml'])
    })

    it('offers bookmarks per row only for a horizontal group', async () => {
        const wrapper = mountForm()
        const slider = wrapper.findComponent(Slider)
        expect(slider.exists()).toBe(true)
        expect(slider.props('modelValue')).toBe(0)
        expect(wrapper.find('.bookmarkgroup-columns-value').text()).toBe('Auto')

        slider.vm.$emit('update:modelValue', 3)
        slider.vm.$emit('slideend', {})
        await wrapper.vm.$nextTick()
        expect(lastEmit(wrapper).columns).toBe(3)
        expect(wrapper.find('.bookmarkgroup-columns-value').text()).toBe('3')

        const vertical = wrapper.findAll('.bookmarkgroup-direction button').find(b => b.text() === 'Vertical')!
        await vertical.trigger('click')
        expect(wrapper.findComponent(Slider).exists()).toBe(false)
    })

    it('shows a stored row length and treats an invalid one as auto', () => {
        expect(mountForm({ ...base, columns: 4 }).findComponent(Slider).props('modelValue')).toBe(4)
        expect(mountForm({ ...base, columns: 99 }).findComponent(Slider).props('modelValue')).toBe(0)
        expect(mountForm({ ...base, columns: -1 }).findComponent(Slider).props('modelValue')).toBe(0)
    })

    it('keeps the open bookmark open when its own update comes back as config', async () => {
        const wrapper = mountForm()
        await items(wrapper)[1].find('.bookmarkgroup-toggle').trigger('click')
        await items(wrapper)[1].find('input[placeholder="Bookmark title"]').setValue('Beta')
        await wrapper.setProps({ config: lastEmit(wrapper) })
        expect(items(wrapper)[1].find('input[placeholder="Bookmark title"]').exists()).toBe(true)
    })

    it('loads a different config handed in from outside', async () => {
        const wrapper = mountForm()
        await wrapper.setProps({ config: { bookmarks: [{ url: '/x', icon: '', title: 'X', subtitle: '' }] } })
        expect(items(wrapper).map(i => i.find('.bookmarkgroup-item-title').text())).toEqual(['X'])
    })
})
