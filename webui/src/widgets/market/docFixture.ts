import type { DocQueryMock } from '@/lib/widgetDocs'
import type { MarketData } from './types'

// Canned market data for the widget docs. Any symbol works; the price curve
// is a seeded random walk, so the same symbol and range always draw the same
// chart, and changing the range in the settings visibly changes it.

const pointsPerRange: Record<string, number> = { '1d': 78, '5d': 60, '1mo': 22, '3mo': 64, '6mo': 126, '1y': 250 }

function seededRandom(seed: string): () => number {
    let h = 2166136261
    for (const c of seed) h = Math.imul(h ^ c.charCodeAt(0), 16777619)
    return () => {
        h = Math.imul(h ^ (h >>> 15), 2246822507)
        h = Math.imul(h ^ (h >>> 13), 3266489909)
        return ((h ^= h >>> 16) >>> 0) / 4294967296
    }
}

export function sampleMarket(symbol: string, range: string): MarketData {
    const rand = seededRandom(`${symbol}:${range}`)
    const count = pointsPerRange[range] ?? 22
    const step = range === '1d' ? 5 * 60_000 : 24 * 3600_000
    const start = Date.UTC(2026, 0, 2)

    let close = 180
    const points = Array.from({ length: count }, (_, i) => {
        close = Math.max(1, close * (1 + (rand() - 0.47) * 0.03))
        return { time: new Date(start + i * step).toISOString(), close: Math.round(close * 100) / 100 }
    })

    const first = points[0].close
    const price = points[points.length - 1].close
    const change = Math.round((price - first) * 100) / 100
    return {
        quote: {
            symbol: symbol.toUpperCase(),
            name: 'Sample Holdings Inc.',
            currency: 'USD',
            price,
            change,
            changePercent: (change / first) * 100,
            marketState: 'REGULAR',
        },
        points,
    }
}

export const marketMocks: Record<string, DocQueryMock> = {
    market: (key) => sampleMarket(String(key[1] ?? 'DEMO'), String(key[2] ?? '1mo')),
}
