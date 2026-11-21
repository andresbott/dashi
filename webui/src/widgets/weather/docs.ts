import type { WidgetDocs } from '@/lib/widgetDocs'
import { weatherMocks } from './docFixture'

const location = 'Location'
const display = 'Display toggles'
const extra = 'Extra info toggles'
const forecast = 'Hourly & forecast'
const graph = 'Graph options'

const docs: WidgetDocs = {
    summary: 'Comprehensive weather widget showing current conditions, hourly and daily forecasts, and optional temperature/rain graphs. Data is fetched from the Open-Meteo API based on geographic coordinates. See also Weather (Compact) for a minimal single-line variant.',
    params: [
        { group: location, name: 'city', type: 'string', description: 'Display name of the location.' },
        { group: location, name: 'latitude', type: 'number', description: 'Latitude coordinate (required).' },
        { group: location, name: 'longitude', type: 'number', description: 'Longitude coordinate (required).' },

        { group: display, name: 'showCurrent', type: 'boolean', default: 'true', description: 'Show current conditions (temperature, icon, description).' },
        { group: display, name: 'showDetails', type: 'boolean', default: 'true', description: 'Show feels-like temperature, humidity, and wind in the current section.' },
        { group: display, name: 'showHourly', type: 'boolean', default: 'false', description: 'Show hourly forecast slots.' },
        { group: display, name: 'showForecast', type: 'boolean', default: 'true', description: 'Show multi-day forecast.' },
        { group: display, name: 'showGraph', type: 'boolean', default: 'false', description: 'Show temperature and rain graph.' },

        { group: extra, name: 'showSunrise', type: 'boolean', default: 'false', description: 'Sunrise time.' },
        { group: extra, name: 'showSunset', type: 'boolean', default: 'false', description: 'Sunset time.' },
        { group: extra, name: 'showWind', type: 'boolean', default: 'false', description: 'Wind speed.' },
        { group: extra, name: 'showHumidity', type: 'boolean', default: 'false', description: 'Humidity percentage.' },
        { group: extra, name: 'showPressure', type: 'boolean', default: 'false', description: 'Atmospheric pressure (hPa).' },
        { group: extra, name: 'showUV', type: 'boolean', default: 'false', description: 'UV index.' },
        { group: extra, name: 'showVisibility', type: 'boolean', default: 'false', description: 'Visibility distance (km).' },
        { group: extra, name: 'showAirQuality', type: 'boolean', default: 'false', description: 'Air quality index (AQI).' },

        { group: forecast, name: 'hourlyCount', type: 'number', default: '12', description: 'Hours ahead to fetch (1–24).' },
        { group: forecast, name: 'hourlySlots', type: 'number', default: '6', description: 'Number of time slots to display (1–12).' },
        { group: forecast, name: 'forecastDays', type: 'number', default: '7', description: 'Days to show in the daily forecast (1–7).' },

        { group: graph, name: 'graphHours', type: 'number', default: '24', description: 'Hours of data shown in the graph (3–24).' },
        { group: graph, name: 'graphHeight', type: 'number', default: '250', description: 'Graph height in pixels (100–600).' },
        { group: graph, name: 'graphTempColor', type: 'hex color', default: '#FF8C42', description: 'Color of the temperature line.' },
        { group: graph, name: 'graphRainColor', type: 'hex color', default: '#4A90D9', description: 'Color of the rain bars.' },
        { group: graph, name: 'graphBgColor', type: 'hex color', default: 'transparent', description: 'Graph background color.' },
        { group: graph, name: 'graphShowTemp', type: 'boolean', default: 'true', description: 'Show temperature labels on the graph.' },
        { group: graph, name: 'graphShowRain', type: 'boolean', default: 'true', description: 'Show rain percentage labels.' },
    ],
    notes: ['The extra info toggles add their data points alongside the current conditions when enabled.'],
    preview: {
        title: 'Weather',
        config: {
            city: 'Zürich, Switzerland',
            latitude: 47.3769,
            longitude: 8.5417,
            showCurrent: true,
            showDetails: true,
            showForecast: true,
            forecastDays: 5,
        },
        settingsNote: 'The preview uses sample weather, and the city search offers a few sample cities.',
    },
    mocks: weatherMocks,
}

export default docs
