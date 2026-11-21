import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'Displays a battery level indicator. The value is read from the `battery` URL query parameter, making it suitable for external devices that pass their battery level via URL.',
    params: [],
    paramsNote: 'This widget has no configuration properties. It reads its value entirely from the URL.',
    sections: [
        {
            title: 'Value source',
            table: {
                head: ['Source', 'Description'],
                rows: [['`?battery=75`', 'Value from 0 to 100. Clamped to valid range.']],
            },
        },
    ],
    preview: {
        title: 'Battery',
        settingsNote: 'Nothing to configure. The editor shows this placeholder; the dashboard shows the level passed in the URL.',
    },
}

export default docs
