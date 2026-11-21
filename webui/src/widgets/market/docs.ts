import type { WidgetDocs } from '@/lib/widgetDocs'
import { marketMocks } from './docFixture'

const docs: WidgetDocs = {
    summary: 'Displays a stock or cryptocurrency ticker with current price, change percentage, and an optional price history chart. Data is fetched from Yahoo Finance.',
    params: [
        { name: 'symbol', type: 'string', description: 'Ticker symbol, e.g. `AAPL`, `BTC-USD`.' },
        { name: 'range', type: 'string', default: '1mo', description: 'Time period: `1d`, `5d`, `1mo`, `3mo`, `6mo`, `1y`.' },
        { name: 'showChart', type: 'boolean', default: 'true', description: 'Display the price history chart.' },
        { name: 'showChange', type: 'boolean', default: 'true', description: 'Show price change amount and percentage.' },
    ],
    notes: ['The chart is rendered server-side as a PNG image. Price changes are colored green (positive) or red (negative).'],
    preview: {
        title: 'Market',
        config: { symbol: 'AAPL', range: '1mo', showChart: true, showChange: true },
        settingsNote: 'The preview draws sample prices for any symbol.',
    },
    mocks: marketMocks,
}

export default docs
