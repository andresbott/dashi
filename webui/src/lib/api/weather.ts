import { apiClient } from '@/lib/api/client'
import type { WeatherData, Location } from '@/types/weather'

const WEATHER_PATH = '/widgets/weather'

export const getWeather = async (lat: number, lon: number): Promise<WeatherData> => {
    const { data } = await apiClient.get<WeatherData>(WEATHER_PATH, {
        params: { lat, lon }
    })
    return data
}

export const geocodeCity = async (city: string): Promise<Location[]> => {
    const { data } = await apiClient.get<Location[]>(`${WEATHER_PATH}/geocode`, {
        params: { city }
    })
    return data
}
