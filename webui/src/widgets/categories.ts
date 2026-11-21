// Widget categories, in display order. Each widget's index.ts names one; the
// widget documentation index groups by them, and the add-widget picker can
// reuse the same grouping.
export const WIDGET_CATEGORIES = [
    { id: 'time', label: 'Time & date' },
    { id: 'weather', label: 'Weather' },
    { id: 'information', label: 'Information' },
    { id: 'navigation', label: 'Navigation & layout' },
] as const

export type WidgetCategory = (typeof WIDGET_CATEGORIES)[number]['id']
