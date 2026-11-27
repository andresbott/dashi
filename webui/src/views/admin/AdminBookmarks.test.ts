import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import AdminBookmarks from './AdminBookmarks.vue'
import { getBookmarkFile } from '@/lib/api/bookmarkFiles'

const files = ref<string[]>([])
const saveFile = vi.fn()
const deleteFile = vi.fn()

vi.mock('@/composables/useBookmarkFiles', () => ({
    useBookmarkFiles: () => ({ files, isLoading: ref(false), saveFile, deleteFile }),
}))
vi.mock('@/composables/useDashboards', () => ({
    useDashboardAssets: () => ({ data: ref([]) }),
}))
vi.mock('@/lib/api/bookmarkFiles', () => ({ getBookmarkFile: vi.fn() }))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: vi.fn() }) }))

function mountView() {
    return mount(AdminBookmarks, {
        global: { plugins: [PrimeVue], stubs: { teleport: true, IconSelector: true } },
    })
}

type Wrapper = ReturnType<typeof mountView>

const button = (wrapper: Wrapper, label: string) => wrapper.findAll('button').find(b => b.text() === label)!

async function openEdit(wrapper: Wrapper) {
    await wrapper.find('.bookmarkfile-edit').trigger('click')
    await flushPromises()
}

describe('AdminBookmarks', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        files.value = []
        // PrimeVue's Select reads matchMedia on mount; jsdom has none.
        vi.stubGlobal('matchMedia', () => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
    })
    afterEach(() => vi.unstubAllGlobals())

    it('explains what bookmark files are when there are none', () => {
        expect(mountView().text()).toContain('No bookmark files yet')
    })

    it('lists the files', () => {
        files.value = ['links.yaml', 'tools.yml']
        const text = mountView().text()
        expect(text).toContain('links.yaml')
        expect(text).toContain('tools.yml')
    })

    it('creates an empty file, adding .yaml, and opens it', async () => {
        saveFile.mockResolvedValue(undefined)
        vi.mocked(getBookmarkFile).mockResolvedValue([])
        const wrapper = mountView()
        await button(wrapper, 'New').trigger('click')
        await wrapper.find('input[placeholder="links.yaml"]').setValue('work')
        await button(wrapper, 'Create').trigger('click')
        await flushPromises()
        expect(saveFile).toHaveBeenCalledWith({ name: 'work.yaml', bookmarks: [] })
        expect(getBookmarkFile).toHaveBeenCalledWith('work.yaml')
        expect(wrapper.find('.bookmarkfile-empty').exists()).toBe(true)
    })

    it('edits the bookmarks of a file with the list editor and saves them', async () => {
        files.value = ['links.yaml']
        vi.mocked(getBookmarkFile).mockResolvedValue([{ url: 'https://a.example', icon: 'ti-world', title: 'A', subtitle: '' }])
        saveFile.mockResolvedValue(undefined)
        const wrapper = mountView()
        await openEdit(wrapper)

        expect(wrapper.findAll('.bookmarkgroup-item-title').map(t => t.text())).toEqual(['A'])
        await button(wrapper, 'Add bookmark').trigger('click')
        await wrapper.find('input[placeholder="Bookmark title"]').setValue('B')
        await wrapper.find('input[placeholder="https://example.com"]').setValue('/b')

        await wrapper.find('.bookmarkfile-save').trigger('click')
        await flushPromises()
        expect(saveFile).toHaveBeenCalledWith({
            name: 'links.yaml',
            bookmarks: [
                { url: 'https://a.example', icon: 'ti-world', title: 'A', subtitle: '' },
                { url: '/b', icon: 'ti-bookmark', title: 'B', subtitle: '' },
            ],
        })
    })

    it('does not save while a URL fails the bookmark rule', async () => {
        files.value = ['links.yaml']
        vi.mocked(getBookmarkFile).mockResolvedValue([])
        const wrapper = mountView()
        await openEdit(wrapper)
        await button(wrapper, 'Add bookmark').trigger('click')
        await wrapper.find('input[placeholder="https://example.com"]').setValue('javascript:alert(1)')

        expect(wrapper.find('.bookmarkfile-save').attributes('disabled')).toBeDefined()
        await wrapper.find('.bookmarkfile-save').trigger('click')
        expect(saveFile).not.toHaveBeenCalled()
    })

    it('shows why a file cannot be edited', async () => {
        files.value = ['broken.yaml']
        vi.mocked(getBookmarkFile).mockRejectedValue({ response: { data: { error: 'bookmarks: invalid file: yaml: line 2: bad' } } })
        const wrapper = mountView()
        await openEdit(wrapper)
        expect(wrapper.find('.bookmarkfile-load-error').text()).toContain('bookmarks: invalid file: yaml: line 2: bad')
        expect(wrapper.find('.bookmarkfile-save').attributes('disabled')).toBeDefined()
    })

    it('shows the server message when a save is rejected', async () => {
        files.value = ['links.yaml']
        vi.mocked(getBookmarkFile).mockResolvedValue([])
        saveFile.mockRejectedValue({ response: { data: { error: 'data: invalid name' } } })
        const wrapper = mountView()
        await openEdit(wrapper)
        await wrapper.find('.bookmarkfile-save').trigger('click')
        await flushPromises()
        expect(wrapper.find('.bookmarkfile-save-error').text()).toBe('data: invalid name')
    })
})
