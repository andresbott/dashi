import type { WidgetModule } from '@/widgets/types'

// Each widget folder registers itself through its index.ts. Those files only
// declare metadata and async components, so eager loading here keeps every
// widget in its own chunk.
const discovered = import.meta.glob<{ default: WidgetModule }>('../widgets/*/index.ts', { eager: true })

const modules: WidgetModule[] = Object.values(discovered)
    .map(m => m.default)
    .sort((a, b) => a.label.localeCompare(b.label))

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
