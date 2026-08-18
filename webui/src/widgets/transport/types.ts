export interface TransportStation {
    id: string
    name: string
    latitude: number
    longitude: number
}

export interface TransportDeparture {
    category: string
    number: string
    destination: string
    scheduled: string
    expected: string
    delay: number
    platform: string
}

export interface TransportWidgetConfig {
    stationId?: string
    stationName?: string
    limit?: number
}
