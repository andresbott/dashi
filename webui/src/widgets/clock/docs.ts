import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'Displays a digital clock with optional date. The time updates every second in interactive dashboards.',
    params: [
        { name: 'hour12', type: 'boolean', default: 'false', description: 'Use 12-hour format instead of 24-hour.' },
        { name: 'showSeconds', type: 'boolean', default: 'true', description: 'Show seconds in the time display.' },
        { name: 'showDate', type: 'boolean', default: 'true', description: 'Show the date below the time.' },
        { name: 'font', type: 'string', description: "Font name from the dashboard's widget theme. Falls back to the theme's default font." },
    ],
    preview: {
        title: 'Clock',
        config: { hour12: false, showSeconds: true, showDate: true },
    },
}

export default docs
