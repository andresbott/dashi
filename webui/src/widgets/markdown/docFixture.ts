import type { DocQueryMock } from '@/lib/widgetDocs'

// Sample notes for the widget docs, as the server would render them (the
// widget shows server-rendered HTML).
export const sampleNotes: Record<string, string> = {
    'welcome.md': [
        '<h3>Welcome home</h3>',
        '<p>This note lives in <strong>Content → Notes</strong> and can be shown on any dashboard.</p>',
        '<blockquote>Wi-Fi: <code>dashi-guest</code></blockquote>',
    ].join(''),
    'chores.md': [
        '<h3>This week</h3>',
        '<ul><li>Water the plants</li><li>Take out the recycling (Thursday)</li><li>Book the bike service</li></ul>',
        '<p><a href="https://example.com">Shopping list</a></p>',
    ].join(''),
}

export const markdownMocks: Record<string, DocQueryMock> = {
    'data-notes-list': () => Object.keys(sampleNotes),
    'data-notes-html': (key) => sampleNotes[String(key[1])],
}
