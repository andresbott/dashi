import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { getWidgetTypes } from './widgetRegistry'
import { allDocMocks, docSegments, getDocumentedTypes, getWidgetDocs } from './widgetDocs'

interface ContractConfig {
    fields: string[]
    editorOnly?: string[]
}

// The same file pins the config fields on the Go and editor sides (see
// webui/src/widgets/contract.test.ts), so checking the docs against it means a
// field added to a widget fails here until it is documented.
const contract: { configs: Record<string, ContractConfig> } = JSON.parse(
    readFileSync(resolve(__dirname, '../../../internal/widgets/testdata/widget_contract.json'), 'utf-8'),
)
const contractFields = (key: string) => {
    const cfg = contract.configs[key]
    return cfg ? [...cfg.fields, ...(cfg.editorOnly ?? [])].sort() : []
}

const configKey = (type: string) => getWidgetDocs(type)?.configType ?? type
const documented = (type: string) => (getWidgetDocs(type)?.params ?? []).map(p => p.name)

describe('widget docs', () => {
    it('documents every registered widget', () => {
        expect(getDocumentedTypes().sort()).toEqual([...getWidgetTypes()].sort())
    })

    // Widgets sharing a config (weather, weather-compact) split the fields
    // between them, so the check is per config, over all its widgets.
    it.each(Object.keys(contract.configs))('documents every %s config field', (key) => {
        const sharing = getWidgetTypes().filter(t => configKey(t) === key)
        const names = new Set(sharing.flatMap(documented))
        expect([...names].sort()).toEqual(contractFields(key))
    })

    it.each(getWidgetTypes())('%s documents only real config fields, once each', (type) => {
        const names = documented(type)
        expect(new Set(names).size).toBe(names.length)
        const allowed = contractFields(configKey(type))
        for (const name of names) expect(allowed).toContain(name)
    })

    it.each(getWidgetTypes())('%s has a summary and describes every parameter', (type) => {
        const docs = getWidgetDocs(type)!
        expect(docs.summary.trim()).not.toBe('')
        for (const p of docs.params) {
            expect(p.type.trim(), p.name).not.toBe('')
            expect(p.description.trim(), p.name).not.toBe('')
        }
        if (docs.params.length === 0) expect(docs.paramsNote).toBeTruthy()
    })

    it.each(getWidgetTypes())('%s preview config uses only real config fields', (type) => {
        const allowed = contractFields(configKey(type))
        for (const key of Object.keys(getWidgetDocs(type)!.preview.config ?? {})) {
            expect(allowed).toContain(key)
        }
    })

    it('mocks the data endpoints of the data-backed widgets', () => {
        expect(Object.keys(allDocMocks()).sort()).toEqual([
            'data-images', 'data-notes-html', 'data-notes-list', 'geocode', 'market',
            'stationSearch', 'sysinfo', 'transport', 'weather', 'xkcd',
        ])
    })
})

describe('docSegments', () => {
    it('splits backtick spans into code segments', () => {
        expect(docSegments('Use `a` or `b`.')).toEqual([
            { code: false, text: 'Use ' },
            { code: true, text: 'a' },
            { code: false, text: ' or ' },
            { code: true, text: 'b' },
            { code: false, text: '.' },
        ])
    })
})
