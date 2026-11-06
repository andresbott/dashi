import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import ts from 'typescript'
import type { WidgetModule } from './types'

interface ContractConfig {
    fields: string[]
    editorOnly?: string[]
    noForm?: boolean
}

// internal/widgets/contract_test.go checks the Go config structs against the
// same file, so a field added on one side only fails. editorOnly lists fields
// the viewer ignores: known gaps, not approved design.
const contract: { configs: Record<string, ContractConfig> } = JSON.parse(
    readFileSync(
        resolve(__dirname, '../../../internal/widgets/testdata/widget_contract.json'),
        'utf-8',
    ),
)

const typeSources = import.meta.glob<string>('./*/types.ts', {
    query: '?raw',
    import: 'default',
    eager: true,
})
const indexes = import.meta.glob<{ default: WidgetModule }>('./*/index.ts', { eager: true })

function configProperties(source: string): string[] | undefined {
    const file = ts.createSourceFile('types.ts', source, ts.ScriptTarget.Latest)
    const config = file.statements.find(
        (s): s is ts.InterfaceDeclaration =>
            ts.isInterfaceDeclaration(s) && s.name.text.endsWith('WidgetConfig'),
    )
    return config?.members
        .filter(ts.isPropertySignature)
        .map(m => m.name.getText(file))
        .sort()
}

// Keyed by widget type, because a folder name can differ from its type.
const discovered: Record<string, string[]> = {}
for (const [path, source] of Object.entries(typeSources)) {
    const properties = configProperties(source)
    if (properties) {
        const widgetType = indexes[path.replace('types.ts', 'index.ts')].default.type
        discovered[widgetType] = properties
    }
}

const withForm = Object.entries(contract.configs).filter(([, cfg]) => !cfg.noForm)

describe('widget config contract', () => {
    it('covers every editor config type', () => {
        expect(Object.keys(discovered).sort()).toEqual(withForm.map(([type]) => type).sort())
    })

    for (const [type, cfg] of withForm) {
        it(`${type} fields match`, () => {
            expect(discovered[type]).toEqual([...cfg.fields, ...(cfg.editorOnly ?? [])].sort())
        })
    }
})
