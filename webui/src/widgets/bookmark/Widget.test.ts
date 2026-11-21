import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Widget from './Widget.vue'
import type { Widget as DashboardWidget } from '@/types/dashboard'

function mountWidget(url: string) {
    const widget = {
        id: 'w1',
        type: 'bookmark',
        title: '',
        width: 3,
        config: { url, icon: 'ti-bookmark', title: 'Example', subtitle: '' },
    } as unknown as DashboardWidget
    return mount(Widget, { props: { widget } })
}

describe('bookmark Widget href', () => {
    it.each(['https://example.com', '/abc123'])('links to an allowed URL (%s)', (url) => {
        expect(mountWidget(url).find('a.bookmark-link').attributes('href')).toBe(url)
    })

    // Stored dashboards may predate the server-side check: an unsafe URL must
    // not reach the href, the link renders inert instead.
    it.each(['javascript:alert(1)', 'JAVASCRIPT:alert(1)', 'data:text/html,x', '//evil.com'])(
        'renders no href for %s',
        (url) => {
            const link = mountWidget(url).find('a.bookmark-link')
            expect(link.exists()).toBe(true)
            expect(link.attributes('href')).toBeUndefined()
        },
    )
})
