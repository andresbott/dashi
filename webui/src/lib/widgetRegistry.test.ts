import { describe, it, expect } from 'vitest'
import { getWidgetEntry, getWidgetTypes, getWidgetTypeOptions } from './widgetRegistry'

describe('widgetRegistry', () => {
    it('has a unique type for every widget', () => {
        const types = getWidgetTypes()
        expect(new Set(types).size).toBe(types.length)
    })

    it('resolves the types saved dashboards use', () => {
        expect(getWidgetEntry('transport')?.label).toBe('Departures')
        expect(getWidgetEntry('page-indicator')).toBeDefined()
        expect(getWidgetEntry('weather-compact')).toBeDefined()
        expect(getWidgetEntry('nope')).toBeUndefined()
    })

    it('orders picker options by label', () => {
        const labels = getWidgetTypeOptions().map(o => o.label)
        expect(labels).toEqual([...labels].sort((a, b) => a.localeCompare(b)))
    })
})
