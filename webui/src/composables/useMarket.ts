import { useQuery } from '@tanstack/vue-query'
import { getMarketData } from '@/lib/api/market'
import type { Ref } from 'vue'

export function useMarket(symbol: Ref<string | undefined>, range: Ref<string>) {
    return useQuery({
        queryKey: ['market', symbol, range],
        queryFn: () => getMarketData(symbol.value!, range.value),
        enabled: () => !!symbol.value,
        refetchInterval: () => {
            switch (range.value) {
                case '1d': return 15 * 60 * 1000
                case '5d': return 60 * 60 * 1000
                default:   return 60 * 60 * 1000
            }
        },
    })
}
