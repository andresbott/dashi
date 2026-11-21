import type { DocQueryMock } from '@/lib/widgetDocs'
import type { DailyForecast, HourlyForecast, Location, WeatherData } from './types'

// Canned weather for the widget docs: the same mild autumn day for any
// location. Times are built from "now" so the hourly slots and day names
// look current; the values never change.

const pad = (n: number) => String(n).padStart(2, '0')
const localDate = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
const localDateTime = (d: Date) => `${localDate(d)}T${pad(d.getHours())}:00`

const sky = {
    clear: { weatherCode: 0, description: 'Clear sky', icon: 'clear-sky' },
    partly: { weatherCode: 2, description: 'Partly cloudy', icon: 'partly-cloudy' },
    overcast: { weatherCode: 3, description: 'Overcast', icon: 'overcast' },
    rain: { weatherCode: 61, description: 'Slight rain', icon: 'rain-slight' },
    showers: { weatherCode: 80, description: 'Slight rain showers', icon: 'rain-showers-slight' },
}

const dailySky = [sky.partly, sky.clear, sky.showers, sky.rain, sky.overcast, sky.partly, sky.clear]
const dailyMax = [14, 16, 12, 10, 11, 13, 15]
const dailyMin = [6, 7, 5, 4, 4, 5, 7]
const rainChance = [0, 0, 5, 10, 20, 35, 60, 70, 55, 30, 15, 5]

export function sampleWeather(now = new Date()): WeatherData {
    const start = new Date(now)
    start.setMinutes(0, 0, 0)

    const hourly: HourlyForecast[] = Array.from({ length: 24 }, (_, i) => {
        const t = new Date(start.getTime() + i * 3600_000)
        // A gentle daily curve: warmest mid-afternoon.
        const temperature = 10 + 4 * Math.sin(((t.getHours() - 9) / 24) * 2 * Math.PI)
        const precipitationProbability = rainChance[i % rainChance.length]
        const s = precipitationProbability >= 50 ? sky.showers : precipitationProbability >= 20 ? sky.overcast : sky.partly
        return { time: localDateTime(t), temperature: Math.round(temperature * 10) / 10, precipitationProbability, ...s }
    })

    const forecast: DailyForecast[] = dailySky.map((s, i) => {
        const day = new Date(start)
        day.setDate(day.getDate() + i)
        const date = localDate(day)
        return {
            date,
            tempMin: dailyMin[i],
            tempMax: dailyMax[i],
            sunrise: `${date}T07:24`,
            sunset: `${date}T16:51`,
            uvIndex: 1.8,
            ...s,
        }
    })

    return {
        current: {
            temperature: 12.4,
            feelsLike: 10.8,
            humidity: 68,
            windSpeed: 11,
            pressure: 1017,
            visibility: 24,
            ...sky.partly,
        },
        hourly,
        forecast,
        airQuality: { europeanAqi: 21 },
    }
}

const sampleLocations: Location[] = [
    { name: 'Zürich', country: 'Switzerland', latitude: 47.3769, longitude: 8.5417 },
    { name: 'Berlin', country: 'Germany', latitude: 52.52, longitude: 13.405 },
    { name: 'Barcelona', country: 'Spain', latitude: 41.3874, longitude: 2.1686 },
    { name: 'Boston', country: 'United States', latitude: 42.3601, longitude: -71.0589 },
    { name: 'Tokyo', country: 'Japan', latitude: 35.6762, longitude: 139.6503 },
]

// Case- and accent-insensitive, so "zur" finds Zürich.
const fold = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()

export const weatherMocks: Record<string, DocQueryMock> = {
    weather: () => sampleWeather(),
    // The city search in the settings form only offers a few sample cities.
    geocode: (key) => {
        const q = fold(String(key[1] ?? ''))
        return sampleLocations.filter(l => fold(l.name).includes(q))
    },
}
