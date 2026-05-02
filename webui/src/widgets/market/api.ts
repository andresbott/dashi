import { apiClient } from '@/lib/api/client'
import type { MarketData } from '@/types/market'

const MARKET_PATH = '/widgets/market'

export const getMarketData = async (symbol: string, range: string = '1mo'): Promise<MarketData> => {
    const { data } = await apiClient.get<MarketData>(MARKET_PATH, {
        params: { symbol, range }
    })
    return data
}
