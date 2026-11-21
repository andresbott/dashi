import type { DocQueryMock } from '@/lib/widgetDocs'
import type { TransportDeparture, TransportStation } from './types'

// Canned departures for the widget docs: the same board for any stop, with
// times relative to now so the countdowns read naturally. One departure runs
// late and one is over an hour out (shown as HH:MM).

const board: { category: string; number: string; destination: string; inMin: number; delay: number; platform: string }[] = [
    { category: 'T', number: '4', destination: 'Tiefenbrunnen', inMin: 2, delay: 0, platform: '' },
    { category: 'S', number: 'S8', destination: 'Winterthur', inMin: 5, delay: 2, platform: '33' },
    { category: 'B', number: '31', destination: 'Hegibachplatz', inMin: 7, delay: 0, platform: '' },
    { category: 'IC', number: 'IC 5', destination: 'Genève-Aéroport', inMin: 12, delay: 0, platform: '16' },
    { category: 'T', number: '13', destination: 'Albisgütli', inMin: 15, delay: 0, platform: '' },
    { category: 'S', number: 'S3', destination: 'Aarau', inMin: 21, delay: 0, platform: '41' },
    { category: 'B', number: '46', destination: 'Rütihof', inMin: 26, delay: 1, platform: '' },
    { category: 'IR', number: 'IR 36', destination: 'Basel SBB', inMin: 34, delay: 0, platform: '9' },
    { category: 'T', number: '11', destination: 'Auzelg', inMin: 41, delay: 0, platform: '' },
    { category: 'S', number: 'S12', destination: 'Brugg AG', inMin: 48, delay: 0, platform: '34' },
    { category: 'B', number: '32', destination: 'Strassenverkehrsamt', inMin: 55, delay: 0, platform: '' },
    { category: 'IC', number: 'IC 1', destination: 'St. Gallen', inMin: 63, delay: 0, platform: '13' },
    { category: 'T', number: '7', destination: 'Stettbach', inMin: 70, delay: 0, platform: '' },
    { category: 'S', number: 'S5', destination: 'Pfäffikon SZ', inMin: 78, delay: 0, platform: '42' },
    { category: 'IR', number: 'IR 75', destination: 'Konstanz', inMin: 92, delay: 0, platform: '6' },
]

export function sampleDepartures(limit: number, now = Date.now()): TransportDeparture[] {
    return board.slice(0, Math.max(0, limit)).map(d => {
        const scheduled = new Date(now + d.inMin * 60_000)
        const expected = new Date(scheduled.getTime() + d.delay * 60_000)
        return {
            category: d.category,
            number: d.number,
            destination: d.destination,
            scheduled: scheduled.toISOString(),
            expected: expected.toISOString(),
            delay: d.delay,
            platform: d.platform,
        }
    })
}

const sampleStations: TransportStation[] = [
    { id: '8503000', name: 'Zürich HB', latitude: 47.3779, longitude: 8.5403 },
    { id: '8507000', name: 'Bern', latitude: 46.949, longitude: 7.4391 },
    { id: '8500010', name: 'Basel SBB', latitude: 47.5474, longitude: 7.5896 },
    { id: '8505000', name: 'Luzern', latitude: 47.0502, longitude: 8.3102 },
    { id: '8501008', name: 'Genève', latitude: 46.2102, longitude: 6.1425 },
]

const fold = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()

export const transportMocks: Record<string, DocQueryMock> = {
    transport: (key) => sampleDepartures(Number(key[2] ?? 5)),
    // The stop search in the settings form only offers a few sample stops.
    stationSearch: (key) => {
        const q = fold(String(key[1] ?? ''))
        return sampleStations.filter(s => fold(s.name).includes(q))
    },
}
