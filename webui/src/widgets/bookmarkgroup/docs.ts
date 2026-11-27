import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'A set of bookmarks in one widget, laid out as a row or a column. Layout and link behaviour are set once for the whole group; each bookmark has its own icon, title, subtitle and URL. The bookmarks can live in the dashboard or in a shared YAML file.',
    params: [
        { name: 'bookmarks', type: 'array', default: '[]', description: 'The bookmarks, in display order. Each has `url`, `icon`, `title` and `subtitle`, with the same rules as the Bookmark widget: only `http://`, `https://` and relative-path URLs are allowed.' },
        { name: 'source', type: 'string', default: '"dashboard"', description: 'Where the bookmarks come from. `dashboard`: the `bookmarks` list above. `file`: the YAML file named by `file`; `bookmarks` is kept but not shown.' },
        { name: 'file', type: 'string', default: '""', description: 'With `source: file`: a file in `{DataDir}/content/bookmarks`, e.g. `links.yaml`. Manage these under Admin → Content → Bookmarks, or edit them on disk.' },
        { name: 'direction', type: 'string', default: '"horizontal"', description: 'Layout of the bookmarks. `horizontal`: rows of equal-width columns (see `columns`). `vertical`: a column.' },
        { name: 'columns', type: 'number', default: '0', description: 'Horizontal only: how many bookmarks go in one row before the next row starts (1–12). `0`: as many as fit the width, at least 10rem each.' },
        { name: 'textBelow', type: 'boolean', default: 'false', description: 'Applies to every bookmark. `false`: icon left, text right. `true`: icon top, text below.' },
        { name: 'newTab', type: 'boolean', default: 'false', description: 'Applies to every bookmark. `false`: links open in place. `true`: links open in a new browser tab.' },
    ],
    notes: [
        'Icons use the same prefixes as the Bookmark widget (`ti-*`, `selfhst:`, `dashboard:`).',
        'A bookmarks file holds a `bookmarks:` list whose entries have `title`, `url`, `icon` and `subtitle`; any other key is rejected on save. The server reads a file at most once every 10 minutes, so changes made on disk show within that time; saves from the admin page show at once.',
    ],
    preview: {
        title: 'Bookmark group',
        config: {
            bookmarks: [
                { url: 'https://github.com', icon: 'ti-brand-github', title: 'GitHub', subtitle: 'Code hosting' },
                { url: 'https://wikipedia.org', icon: 'ti-world', title: 'Wikipedia', subtitle: 'Encyclopedia' },
                { url: 'https://news.ycombinator.com', icon: 'ti-news', title: 'News', subtitle: '' },
            ],
            direction: 'horizontal',
            textBelow: true,
            newTab: false,
        },
        settingsNote: 'Dashboard asset icons need a dashboard, so the docs list none.',
    },
}

export default docs
