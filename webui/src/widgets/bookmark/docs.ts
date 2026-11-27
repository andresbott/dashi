import type { WidgetDocs } from '@/lib/widgetDocs'

const docs: WidgetDocs = {
    summary: 'A clickable link to an external website, displayed with an icon and title. Useful for building a personal start page or quick-access panel.',
    params: [
        { name: 'url', type: 'string', description: 'Destination URL. Only `http://`, `https://`, and relative paths are allowed.' },
        { name: 'icon', type: 'string', description: 'Icon identifier (see icon types below).' },
        { name: 'title', type: 'string', description: 'Primary display text.' },
        { name: 'subtitle', type: 'string', description: 'Secondary text below the title (optional).' },
        { name: 'textBelow', type: 'boolean', default: 'false', description: 'Layout mode. `false`: icon left, text right. `true`: icon top, text below.' },
        { name: 'newTab', type: 'boolean', default: 'false', description: 'Where the link opens. `false`: in place, replacing the dashboard. `true`: in a new browser tab.' },
    ],
    sections: [
        {
            title: 'Icon types',
            text: ['The icon string determines where the icon is loaded from based on its prefix.'],
            table: {
                head: ['Prefix', 'Source', 'Example'],
                rows: [
                    ['`ti-*`', "Tabler icon from the theme's icon font.", '`ti-home`'],
                    ['`selfhst:`', 'Icon from the selfh.st CDN.', '`selfhst:sonarr.png`'],
                    ['`dashboard:`', "Image file from the dashboard's asset folder.", '`dashboard:logo.svg`'],
                ],
            },
        },
    ],
    preview: {
        title: 'Bookmark',
        config: {
            url: 'https://github.com',
            icon: 'ti-brand-github',
            title: 'GitHub',
            subtitle: 'Code hosting',
            textBelow: false,
            newTab: false,
        },
        settingsNote: 'Dashboard asset icons need a dashboard, so the docs list none.',
    },
}

export default docs
