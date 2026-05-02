export type SearchEngine = 'google' | 'duckduckgo' | 'bing'

export interface SearchWidgetConfig {
    engine: SearchEngine
    placeholder: string
}

export const searchEngineUrls: Record<SearchEngine, string> = {
    google: 'https://www.google.com/search?q=',
    duckduckgo: 'https://duckduckgo.com/?q=',
    bing: 'https://www.bing.com/search?q=',
}
