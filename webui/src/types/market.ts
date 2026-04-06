export interface MarketQuote {
    symbol: string
    name: string
    currency: string
    price: number
    change: number
    changePercent: number
    marketState: string
}

export interface MarketPricePoint {
    time: string
    close: number
}

export interface MarketData {
    quote: MarketQuote
    points: MarketPricePoint[]
}

export interface MarketWidgetConfig {
    symbol?: string
    range?: string
    showChart?: boolean
    showChange?: boolean
}
