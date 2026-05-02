import { defineAsyncComponent, type Component } from 'vue'
import xkcdModule from '@/widgets/xkcd'
import weatherModule from '@/widgets/weather'
import weatherCompactModule from '@/widgets/weather-compact'
import bookmarkModule from '@/widgets/bookmark'
import clockModule from '@/widgets/clock'
import batteryModule from '@/widgets/battery'
import searchModule from '@/widgets/search'
import pageIndicatorModule from '@/widgets/page-indicator'
import marketModule from '@/widgets/market'
import transportModule from '@/widgets/transport'
import stackModule from '@/widgets/stack'
import sysinfoModule from '@/widgets/sysinfo'
import markdownModule from '@/widgets/markdown'

export interface WidgetRegistryEntry {
    component: Component
    configComponent: Component | null
    label: string
    icon: string
    description: string
    noWidgetProp?: boolean
}

const registry: Record<string, WidgetRegistryEntry> = {
    weather: weatherModule,
    'weather-compact': weatherCompactModule,
    bookmark: bookmarkModule,
    clock: clockModule,
    battery: batteryModule,
    search: searchModule,
    'page-indicator': pageIndicatorModule,
    market: marketModule,
    xkcd: xkcdModule,
    transport: transportModule,
    stack: stackModule,
    sysinfo: sysinfoModule,
    markdown: markdownModule,
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
