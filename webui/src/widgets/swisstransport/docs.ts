import type { WidgetDocs } from '@/lib/widgetDocs'
import { transportMocks } from './docFixture'

const docs: WidgetDocs = {
    summary: 'Shows upcoming public transport departures for a given station using the Swiss Federal Railways (SBB) API. Displays line number, destination, departure countdown, and platform.',
    params: [
        { name: 'stationId', type: 'string', description: 'Station identifier from the transport API.' },
        { name: 'stationName', type: 'string', description: 'Display name of the station.' },
        { name: 'limit', type: 'number', default: '5', description: 'Number of departures to show (1–15).' },
    ],
    notes: ['Departures within 60 minutes show a countdown (e.g. "15\'"). Beyond that, the scheduled time is shown (HH:MM). Delayed departures are highlighted.'],
    preview: {
        title: 'Departures',
        config: { stationId: '8503000', stationName: 'Zürich HB', limit: 5 },
        settingsNote: 'The preview shows a sample board, and the stop search offers a few sample stops.',
    },
    mocks: transportMocks,
}

export default docs
