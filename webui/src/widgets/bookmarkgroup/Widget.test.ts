import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { computed, ref, type Ref } from 'vue'
import Widget from './Widget.vue'
import type { Widget as DashboardWidget } from '@/types/dashboard'
import type { BookmarkGroupWidgetConfig } from './types'

// The file query, faked: tests set the file's bookmarks or an error, and
// read which name the widget asked for.
const fileState = {
    files: {} as Record<string, BookmarkGroupWidgetConfig['bookmarks']>,
    error: false,
    asked: [] as string[],
}
vi.mock('@/composables/useBookmarkFiles', () => ({
    useBookmarkFile: (name: Ref<string>) => ({
        bookmarks: computed(() => {
            if (name.value) fileState.asked.push(name.value)
            return fileState.files[name.value] ?? []
        }),
        isLoading: ref(false),
        isError: computed(() => name.value !== '' && fileState.error),
    }),
}))

const two: BookmarkGroupWidgetConfig = {
    bookmarks: [
        { url: 'https://github.com', icon: 'ti-brand-github', title: 'GitHub', subtitle: 'Code hosting' },
        { url: '/notes', icon: '', title: 'Notes', subtitle: '' },
    ],
}

function mountWidget(config: Partial<BookmarkGroupWidgetConfig> | null) {
    const widget = { id: 'w1', type: 'bookmarkgroup', title: '', width: 6, config } as unknown as DashboardWidget
    return mount(Widget, { props: { widget } })
}

describe('bookmarkgroup Widget', () => {
    beforeEach(() => {
        fileState.files = {}
        fileState.error = false
        fileState.asked = []
    })

    it('renders the bookmarks of the configured file instead of its own', () => {
        fileState.files['links.yaml'] = [{ url: 'https://file.example', icon: '', title: 'From file', subtitle: '' }]
        const wrapper = mountWidget({ ...two, source: 'file', file: 'links.yaml' })
        const links = wrapper.findAll('a.bookmark-link')
        expect(links.map(l => l.attributes('href'))).toEqual(['https://file.example'])
        expect(fileState.asked).toContain('links.yaml')
    })

    it('ignores the file while the source is the dashboard', () => {
        fileState.files['links.yaml'] = [{ url: 'https://file.example', icon: '', title: 'From file', subtitle: '' }]
        const links = mountWidget({ ...two, file: 'links.yaml' }).findAll('a.bookmark-link')
        expect(links.map(l => l.attributes('href'))).toEqual(['https://github.com', '/notes'])
        expect(fileState.asked).toEqual([])
    })

    it.each([
        [{ source: 'file' as const }, 'No bookmarks file selected'],
        [{ source: 'file' as const, file: 'gone.yaml' }, 'Bookmarks file "gone.yaml" could not be loaded'],
    ])('explains a file that cannot be shown (%o)', (extra, message) => {
        fileState.error = true
        const wrapper = mountWidget({ ...two, ...extra })
        expect(wrapper.find('a').exists()).toBe(false)
        expect(wrapper.find('.bookmarkgroup-file-error').text()).toBe(message)
    })

    it('says when the file has no bookmarks', () => {
        fileState.files['empty.yaml'] = []
        expect(mountWidget({ source: 'file', file: 'empty.yaml', bookmarks: [] }).text()).toContain('The bookmarks file is empty')
    })

    it('renders one link per bookmark, in stored order', () => {
        const links = mountWidget(two).findAll('a.bookmark-link')
        expect(links.map(l => l.attributes('href'))).toEqual(['https://github.com', '/notes'])
        expect(links[0].text()).toContain('GitHub')
        expect(links[0].text()).toContain('Code hosting')
    })

    it('shows a hint when no bookmarks are configured', () => {
        for (const config of [null, {}, { bookmarks: [] }]) {
            const wrapper = mountWidget(config)
            expect(wrapper.find('a').exists()).toBe(false)
            expect(wrapper.text()).toContain('Add bookmarks in edit mode')
        }
    })

    it.each([
        [undefined, 'horizontal'],
        ['horizontal', 'horizontal'],
        ['vertical', 'vertical'],
        ['diagonal', 'horizontal'],
    ] as const)('lays out direction %s as %s', (direction, want) => {
        const wrapper = mountWidget({ ...two, direction: direction as BookmarkGroupWidgetConfig['direction'] })
        expect(wrapper.find('.bookmarkgroup-list').attributes('data-direction')).toBe(want)
    })

    it.each([
        [{}, 'repeat(auto-fit, minmax(min(100%, 10rem), 1fr))', undefined],
        [{ columns: 3 }, 'repeat(3, minmax(0, 1fr))', '3'],
        [{ columns: 40 }, 'repeat(12, minmax(0, 1fr))', '12'],
        [{ columns: -1 }, 'repeat(auto-fit, minmax(min(100%, 10rem), 1fr))', undefined],
        [{ columns: 3, direction: 'vertical' }, 'minmax(0, 1fr)', undefined],
    ] as const)('lays out %o in grid columns %s', (extra, want, dataColumns) => {
        const list = mountWidget({ ...two, ...extra } as Partial<BookmarkGroupWidgetConfig>).find('.bookmarkgroup-list')
        expect((list.element as HTMLElement).style.gridTemplateColumns).toBe(want)
        expect(list.attributes('data-columns')).toBe(dataColumns)
    })

    it('applies textBelow and newTab to every link', () => {
        const plain = mountWidget(two).findAll('a.bookmark-link')
        for (const link of plain) {
            expect(link.classes()).not.toContain('bookmark-link--below')
            expect(link.attributes('target')).toBeUndefined()
        }

        const shared = mountWidget({ ...two, textBelow: true, newTab: true }).findAll('a.bookmark-link')
        expect(shared).toHaveLength(2)
        for (const link of shared) {
            expect(link.classes()).toContain('bookmark-link--below')
            expect(link.attributes('target')).toBe('_blank')
            expect(link.attributes('rel')).toBe('noopener noreferrer')
        }
    })

    it('renders an unsafe stored URL without an href', () => {
        const links = mountWidget({
            bookmarks: [{ url: 'javascript:alert(1)', icon: '', title: 'x', subtitle: '' }],
        }).findAll('a.bookmark-link')
        expect(links).toHaveLength(1)
        expect(links[0].attributes('href')).toBeUndefined()
    })
})
