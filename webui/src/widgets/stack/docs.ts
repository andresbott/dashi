import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'A container widget that stacks multiple child widgets vertically within a single grid cell. Use it to group related widgets or fit more content into a row without splitting into separate columns.',
    params: [
        { name: 'widgets', type: 'widget[]', description: 'The child widgets, top to bottom. Each entry is a full widget (`id`, `type`, `title`, `config`); its width follows the stack.' },
    ],
    sections: [
        {
            title: 'Children',
            list: [
                'Child widgets are arranged top to bottom.',
                'Each child widget can be any type except another stack.',
                'In the editor, children can be added, removed and individually configured, and reordered by dragging their grip.',
                'Drag any widget (except a stack) from a row into the stack body to add it at that position; it takes the stack\'s width.',
                'Drag a child out onto a row to take it out of the stack; it lands half the grid wide (6 columns), or narrower when the row has less room, and a full row refuses it.',
            ],
        },
    ],
    preview: {
        title: 'Stack',
        config: {
            widgets: [
                { id: 'doc-stack-clock', type: 'clock', title: 'Clock', width: 4, config: { hour12: false, showSeconds: false, showDate: true } },
                {
                    id: 'doc-stack-weather',
                    type: 'weather-compact',
                    title: 'Weather (Compact)',
                    width: 4,
                    config: { city: 'Zürich, Switzerland', latitude: 47.3769, longitude: 8.5417, compact: true, compactCity: true, compactDescription: true, compactAlign: 'center' },
                },
            ],
        },
        settingsNote: 'Children are managed inside the stack in the dashboard editor. The preview stacks a clock and a compact weather widget.',
    },
}

export default docs
