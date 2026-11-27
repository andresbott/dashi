import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Widget from './Widget.vue'
import type { Widget as DashboardWidget } from '@/types/dashboard'
import type { BookmarkGroupWidgetConfig } from './types'

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
