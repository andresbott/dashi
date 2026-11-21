import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'A search bar that opens results in a new browser tab. Purely client-side — no data is fetched from the server.',
    params: [
        { name: 'engine', type: 'string', default: 'google', description: 'Search engine: `google`, `duckduckgo`, or `bing`.' },
        { name: 'placeholder', type: 'string', default: 'Search...', description: 'Placeholder text for the input field.' },
    ],
    preview: {
        title: 'Search',
        config: { engine: 'duckduckgo', placeholder: 'Search the web...' },
    },
}

export default docs
