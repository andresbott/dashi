export interface WeatherData {
    current: CurrentWeather
    hourly: HourlyForecast[]
    forecast: DailyForecast[]
}

export interface CurrentWeather {
    temperature: number
    feelsLike: number
    humidity: number
    windSpeed: number
    weatherCode: number
    description: string
    icon: string
}

export interface HourlyForecast {
    time: string
    temperature: number
    weatherCode: number
    description: string
    icon: string
}

export interface DailyForecast {
    date: string
    tempMin: number
    tempMax: number
    weatherCode: number
    description: string
    icon: string
}

export interface Location {
    name: string
    country: string
    latitude: number
    longitude: number
}

export interface WeatherWidgetConfig {
    city: string
    latitude: number
    longitude: number
    compact?: boolean
    showCurrent?: boolean
    showDetails?: boolean
    showHourly?: boolean
    hourlyCount?: number
    hourlySlots?: number
    showForecast?: boolean
    forecastDays?: number
    iconTheme?: string
}
