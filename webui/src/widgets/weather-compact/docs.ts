import type { WidgetDocs } from '@/lib/widgetDocs'

const location = 'Location'
const display = 'Display'

// Shares the weather widget's config; its data comes from the weather mocks.
const docs: WidgetDocs = {
    summary: 'A minimal single-line weather display showing an icon, temperature, and optional city name. Uses the same data source as the full Weather widget but designed for tight spaces.',
    configType: 'weather',
    params: [
        { group: location, name: 'city', type: 'string', description: 'Display name of the location. Only the part before the first comma is shown.' },
        { group: location, name: 'latitude', type: 'number', description: 'Latitude coordinate (required).' },
        { group: location, name: 'longitude', type: 'number', description: 'Longitude coordinate (required).' },

        { group: display, name: 'compact', type: 'boolean', default: 'true', description: 'Marks the config as the compact variant. Set by the settings form.' },
        { group: display, name: 'compactCity', type: 'boolean', default: 'true', description: 'Show the city name.' },
        { group: display, name: 'compactFeelsLike', type: 'boolean', default: 'false', description: 'Show feels-like temperature.' },
        { group: display, name: 'compactDescription', type: 'boolean', default: 'true', description: 'Show weather description text.' },
        { group: display, name: 'compactAlign', type: 'string', default: 'left', description: 'Alignment: `left`, `center`, or `right`.' },
    ],
    notes: ['Location is configured the same way as the full Weather widget (`city`, `latitude`, `longitude`).'],
    preview: {
        title: 'Weather (Compact)',
        config: {
            city: 'Zürich, Switzerland',
            latitude: 47.3769,
            longitude: 8.5417,
            compact: true,
            compactCity: true,
            compactFeelsLike: false,
            compactDescription: true,
            compactAlign: 'left',
        },
        settingsNote: 'The preview uses sample weather, and the city search offers a few sample cities.',
    },
}

export default docs
