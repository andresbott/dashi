import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { getWidgetEntry, getWidgetTypes, getWidgetTypeOptions } from './widgetRegistry'

describe('widgetRegistry', () => {
    // app/router/widget_contract_test.go checks the same file for the backend,
    // so a widget added on one side only fails.
    it('matches the Go widget contract', () => {
        const contract: { types: string[] } = JSON.parse(
            readFileSync(
                resolve(__dirname, '../../../internal/widgets/testdata/widget_contract.json'),
                'utf-8',
            ),
        )
        expect([...getWidgetTypes()].sort()).toEqual([...contract.types].sort())
    })

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
