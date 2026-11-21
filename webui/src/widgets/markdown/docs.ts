import type { WidgetDocs } from '@/lib/widgetDocs'
import { markdownMocks } from './docFixture'

const docs: WidgetDocs = {
    summary: 'Renders a markdown note: headings, lists, links, quotes and code. Notes are shared across all dashboards and managed under Content → Notes.',
    params: [
        { name: 'filename', type: 'string', description: 'Name of the note to show, e.g. `welcome.md`.' },
    ],
    notes: [
        'In the dashboard editor the settings can also create a note and edit its text in place.',
        'Raw HTML in a note is dropped and unsafe link schemes are removed when it is rendered.',
    ],
    preview: {
        title: 'Markdown',
        config: { filename: 'welcome.md' },
        hideForm: true,
        presets: [
            { label: 'welcome.md', config: { filename: 'welcome.md' } },
            { label: 'chores.md', config: { filename: 'chores.md' } },
        ],
        settingsNote: 'The real settings edit your notes, so here you pick between two sample notes instead.',
    },
    mocks: markdownMocks,
}

export default docs
