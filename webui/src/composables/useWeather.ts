import { useQuery } from '@tanstack/vue-query'
import { getWeather, geocodeCity } from '@/lib/api/weather'
import type { Ref } from 'vue'

const FIFTEEN_MINUTES = 15 * 60 * 1000

export function useWeather(lat: Ref<number | undefined>, lon: Ref<number | undefined>) {
    return useQuery({
        queryKey: ['weather', lat, lon],
        queryFn: () => getWeather(lat.value!, lon.value!),
        enabled: () => lat.value !== undefined && lon.value !== undefined,
        refetchInterval: FIFTEEN_MINUTES,
    })
}

export function useGeocode(city: Ref<string>) {
    return useQuery({
        queryKey: ['geocode', city],
        queryFn: () => geocodeCity(city.value),
        enabled: () => city.value.length >= 2,
    })
}
