import { apiClient } from '@/lib/api/client'
import type { TransportDeparture, TransportStation } from '@/types/transport'

const TRANSPORT_PATH = '/widgets/transport'

export const getDepartures = async (stationId: string, limit: number = 5): Promise<TransportDeparture[]> => {
    const { data } = await apiClient.get<TransportDeparture[]>(`${TRANSPORT_PATH}/stationboard`, {
        params: { id: stationId, limit }
    })
    return data
}

export const searchStations = async (query: string): Promise<TransportStation[]> => {
    const { data } = await apiClient.get<TransportStation[]>(`${TRANSPORT_PATH}/stations`, {
        params: { query }
    })
    return data
}
