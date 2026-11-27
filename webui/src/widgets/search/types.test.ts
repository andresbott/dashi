import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { searchEngineLabels, searchEngineUrls } from './types'

// The engine list shared with internal/widgets/search/engine_test.go, so the
// editor preview and the server renders show the same label and search the
// same URL.
const fixture = JSON.parse(
    readFileSync(
        resolve(__dirname, '../../../../internal/widgets/search/testdata/engines.json'),
        'utf-8',
    ),
) as { default: string; engines: { engine: string; label: string; action: string }[] }

describe('search engines', () => {
    it('lists the same engines as the server', () => {
        const engines = fixture.engines.map((e) => e.engine).sort()
        expect(Object.keys(searchEngineUrls).sort()).toEqual(engines)
        expect(Object.keys(searchEngineLabels).sort()).toEqual(engines)
    })

    it.each(fixture.engines)('$engine matches the server', ({ engine, label, action }) => {
        const key = engine as keyof typeof searchEngineUrls
        expect(searchEngineLabels[key]).toBe(label)
        expect(searchEngineUrls[key]).toBe(`${action}?q=`)
    })
})
