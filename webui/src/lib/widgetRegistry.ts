import type { WidgetModule } from '@/widgets/types'

import weather from '@/widgets/weather'
import weatherCompact from '@/widgets/weather-compact'
import bookmark from '@/widgets/bookmark'
import clock from '@/widgets/clock'
import battery from '@/widgets/battery'
import search from '@/widgets/search'
import pageIndicator from '@/widgets/page-indicator'
import market from '@/widgets/market'
import xkcd from '@/widgets/xkcd'
import transport from '@/widgets/transport'
import stack from '@/widgets/stack'
import sysinfo from '@/widgets/sysinfo'
import markdown from '@/widgets/markdown'
import image from '@/widgets/image'

const modules: WidgetModule[] = [
    weather, weatherCompact, bookmark, clock, battery, search,
    pageIndicator, market, xkcd, transport, stack, sysinfo, markdown, image,
]

const registry: Record<string, WidgetModule> =
    Object.fromEntries(modules.map(m => [m.type, m]))

export function getWidgetEntry(type: string): WidgetModule | undefined {
    return registry[type]
}

export function getWidgetTypes(): string[] {
    return modules.map(m => m.type)
}

export function getWidgetTypeOptions(): { value: string; label: string; icon: string; description: string }[] {
    return modules.map(m => ({ value: m.type, label: m.label, icon: m.icon, description: m.description }))
}

export type WidgetRegistryEntry = WidgetModule
export default registry
