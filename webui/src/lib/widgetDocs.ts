import type { InjectionKey } from 'vue'
import type { WidgetModule } from '@/widgets/types'

// Documentation metadata for one widget, declared in
// webui/src/widgets/{folder}/docs.ts and shown on /admin/docs/widgets/:type.
// Text fields are plain strings; `backticks` render as inline code.

export interface WidgetParamDoc {
    // Config key, exactly as saved (see widget_contract.json).
    name: string
    type: string
    // Shown as code; omit when there is no default.
    default?: string
    description: string
    // Optional sub-heading: parameters with the same group are listed together.
    group?: string
}

export interface WidgetDocSection {
    title: string
    text?: string[]
    list?: string[]
    table?: { head: string[]; rows: string[][] }
}

// Returns the canned payload for a TanStack query key (e.g. ['weather', lat,
// lon]); undefined makes the query fail.
export type DocQueryMock = (queryKey: readonly unknown[]) => unknown

export interface WidgetDocs {
    // Short description at the top of the widget's page.
    summary: string
    // Every config field. widgetDocs.test.ts checks them against the contract.
    params: WidgetParamDoc[]
    // Shown on the Details tab when params is empty.
    paramsNote?: string
    // Extra reference blocks after the parameter table.
    sections?: WidgetDocSection[]
    // Hints at the end of the Details tab.
    notes?: string[]
    // Contract key whose fields this widget documents, when it shares another
    // widget's config (weather-compact uses weather's). Defaults to the type.
    configType?: string
    preview: {
        title?: string
        config?: Record<string, unknown>
        // Preview frame height (CSS), for widgets that fill their cell.
        height?: string
        // Injections the widget reads (page indicator: active and total pages).
        provide?: [InjectionKey<unknown> | string, unknown][]
        // The real settings form writes user data (markdown edits notes):
        // the Overview tab shows `presets` instead.
        hideForm?: boolean
        presets?: { label: string; config: Record<string, unknown> }[]
        // Shown above the settings (or instead of them when there are none).
        settingsNote?: string
    }
    // Canned API payloads by query-key prefix. All widgets' mocks are merged,
    // so a stack child finds its data too.
    mocks?: Record<string, DocQueryMock>
}

const docModules = import.meta.glob<{ default: WidgetDocs }>('../widgets/*/docs.ts', { eager: true })
const indexModules = import.meta.glob<{ default: WidgetModule }>('../widgets/*/index.ts', { eager: true })

// Keyed by widget type: a folder name can differ from its type.
const docsByType: Record<string, WidgetDocs> = {}
for (const [path, mod] of Object.entries(docModules)) {
    const entry = indexModules[path.replace('docs.ts', 'index.ts')]
    if (entry) docsByType[entry.default.type] = mod.default
}

export function getWidgetDocs(type: string): WidgetDocs | undefined {
    return docsByType[type]
}

export function getDocumentedTypes(): string[] {
    return Object.keys(docsByType)
}

export function allDocMocks(): Record<string, DocQueryMock> {
    return Object.assign({}, ...Object.values(docsByType).map(d => d.mocks ?? {}))
}

// Splits `code` spans out of a doc string for rendering without v-html.
export function docSegments(text: string): { code: boolean; text: string }[] {
    return text
        .split('`')
        .map((part, i) => ({ code: i % 2 === 1, text: part }))
        .filter(s => s.text !== '')
}
