import type { WidgetDocs } from '@/lib/widgetDocs'
import { xkcdMocks } from './docFixture'

const docs: WidgetDocs = {
    summary: 'Displays an XKCD webcomic. Clicking the comic opens it on xkcd.com.',
    params: [
        { name: 'mode', type: 'string', default: 'latest', description: 'Display mode: `latest` (most recent comic), `random` (daily random, same all day), or `random-each` (new random on every load).' },
    ],
    notes: ['Comics are cached on disk by the server.'],
    preview: {
        title: 'XKCD',
        // The comic scales to fill its cell, so the frame needs a height.
        height: '15rem',
        config: { mode: 'latest' },
        settingsNote: 'The preview shows a sample panel instead of a real comic.',
    },
    mocks: xkcdMocks,
}

export default docs
