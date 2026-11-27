export type SearchEngine = 'google' | 'duckduckgo' | 'bing'

export interface SearchWidgetConfig {
    engine: SearchEngine
    placeholder: string
}

// Mirrors the allowlist in internal/widgets/search/engine.go; types.test.ts
// checks both against internal/widgets/search/testdata/engines.json.
export const searchEngineUrls: Record<SearchEngine, string> = {
    google: 'https://www.google.com/search?q=',
    duckduckgo: 'https://duckduckgo.com/?q=',
    bing: 'https://www.bing.com/search?q=',
}

export const searchEngineLabels: Record<SearchEngine, string> = {
    google: 'Google',
    duckduckgo: 'DuckDuckGo',
    bing: 'Bing',
}
