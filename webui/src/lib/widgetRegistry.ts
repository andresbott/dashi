import { defineAsyncComponent, type Component } from 'vue'

export interface WidgetRegistryEntry {
    component: Component
    configComponent: Component | null
    label: string
    icon: string
    description: string
    noWidgetProp?: boolean
}

const registry: Record<string, WidgetRegistryEntry> = {
    weather: {
        component: defineAsyncComponent(() => import('@/components/dashboards/WeatherWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/WeatherWidgetConfig.vue')),
        label: 'Weather',
        icon: 'ti-sun',
        description: 'Current conditions and forecast',
    },
    'weather-compact': {
        component: defineAsyncComponent(() => import('@/components/dashboards/WeatherCompactWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/WeatherCompactWidgetConfig.vue')),
        label: 'Weather (Compact)',
        icon: 'ti-cloud',
        description: 'Compact weather display',
    },
    bookmark: {
        component: defineAsyncComponent(() => import('@/components/dashboards/BookmarkWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/BookmarkWidgetConfig.vue')),
        label: 'Bookmark',
        icon: 'ti-bookmark',
        description: 'Link to an external website',
    },
    clock: {
        component: defineAsyncComponent(() => import('@/components/dashboards/ClockWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/ClockWidgetConfig.vue')),
        label: 'Clock',
        icon: 'ti-clock',
        description: 'Digital clock with date',
    },
    battery: {
        component: defineAsyncComponent(() => import('@/components/dashboards/BatteryWidget.vue')),
        configComponent: null,
        label: 'Battery',
        icon: 'ti-battery-2',
        description: 'Battery status from query parameter',
    },
    search: {
        component: defineAsyncComponent(() => import('@/components/dashboards/SearchWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/SearchWidgetConfig.vue')),
        label: 'Search',
        icon: 'ti-search',
        description: 'Search engine input',
    },
    'page-indicator': {
        component: defineAsyncComponent(() => import('@/components/dashboards/PageIndicatorWidget.vue')),
        configComponent: null,
        label: 'Page Indicator',
        icon: 'ti-circles',
        description: 'Shows dots for each page',
        noWidgetProp: true,
    },
    market: {
        component: defineAsyncComponent(() => import('@/components/dashboards/MarketWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/MarketWidgetConfig.vue')),
        label: 'Market',
        icon: 'ti-chart-line',
        description: 'Stock / crypto market ticker',
    },
    xkcd: {
        component: defineAsyncComponent(() => import('@/components/dashboards/XkcdWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/XkcdWidgetConfig.vue')),
        label: 'XKCD',
        icon: 'ti-pencil',
        description: 'XKCD comic strip',
    },
    transport: {
        component: defineAsyncComponent(() => import('@/components/dashboards/TransportWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/TransportWidgetConfig.vue')),
        label: 'Departures',
        icon: 'ti-bus',
        description: 'Public transport departures',
    },
    stack: {
        component: defineAsyncComponent(() => import('@/components/dashboards/StackWidget.vue')),
        configComponent: null,
        label: 'Stack',
        icon: 'ti-layout-rows',
        description: 'Stack widgets vertically in a column',
    },
    sysinfo: {
        component: defineAsyncComponent(() => import('@/components/dashboards/SysinfoWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/SysinfoWidgetConfig.vue')),
        label: 'System Info',
        icon: 'ti-server-cog',
        description: 'Host disk, memory, and uptime',
    },
    markdown: {
        component: defineAsyncComponent(() => import('@/components/dashboards/MarkdownWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/MarkdownWidgetConfig.vue')),
        label: 'Markdown',
        icon: 'ti-markdown',
        description: 'Render markdown content from a file',
    },
    image: {
        component: defineAsyncComponent(() => import('@/components/dashboards/ImageWidget.vue')),
        configComponent: defineAsyncComponent(() => import('@/components/dashboards/ImageWidgetConfig.vue')),
        label: 'Image',
        icon: 'ti-photo',
        description: 'Display an uploaded image',
    },
}

export function getWidgetEntry(type: string): WidgetRegistryEntry | undefined {
    return registry[type]
}

export function getWidgetTypes(): string[] {
    return Object.keys(registry)
}

export function getWidgetTypeOptions(): { value: string; label: string; icon: string; description: string }[] {
    const opts: { value: string; label: string; icon: string; description: string }[] = [
        { value: 'placeholder', label: 'Placeholder', icon: 'ti-layout-grid', description: 'Empty placeholder widget' },
    ]
    for (const [key, entry] of Object.entries(registry)) {
        opts.push({ value: key, label: entry.label, icon: entry.icon, description: entry.description })
    }
    return opts
}

export default registry
