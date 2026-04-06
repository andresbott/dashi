import { useQuery } from '@tanstack/vue-query'
import { getDepartures, searchStations } from '@/lib/api/transport'
import type { Ref } from 'vue'

const ONE_MINUTE = 60 * 1000

export function useTransport(stationId: Ref<string | undefined>, limit: Ref<number>) {
    return useQuery({
        queryKey: ['transport', stationId, limit],
        queryFn: () => getDepartures(stationId.value!, limit.value),
        enabled: () => !!stationId.value,
        refetchInterval: ONE_MINUTE,
    })
}

export function useStationSearch(query: Ref<string>) {
    return useQuery({
        queryKey: ['stationSearch', query],
        queryFn: () => searchStations(query.value),
        enabled: () => query.value.length >= 2,
    })
}
