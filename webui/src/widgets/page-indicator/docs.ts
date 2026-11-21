import { ref } from 'vue'
import type { WidgetDocs } from '@/lib/widgetDocs'
import { ACTIVE_PAGE, TOTAL_PAGES } from '@/lib/injectionKeys'

const docs: WidgetDocs = {
    summary: 'Renders navigation dots for multi-page dashboards. One dot per page; the active page is highlighted. Useful for dashboards displayed on screens where swipe or keyboard navigation is available.',
    params: [],
    paramsNote: 'This widget has no configuration properties. It automatically reflects the current page state.',
    preview: {
        title: 'Page Indicator',
        // Page 2 of 4.
        provide: [
            [ACTIVE_PAGE, ref(1)],
            [TOTAL_PAGES, ref(4)],
        ],
        settingsNote: 'Nothing to configure. The preview shows page 2 of a 4-page dashboard.',
    },
}

export default docs
