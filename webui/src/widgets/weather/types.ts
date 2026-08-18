export interface AirQuality {
    europeanAqi: number
}

export interface WeatherData {
    current: CurrentWeather
    hourly: HourlyForecast[]
    forecast: DailyForecast[]
    airQuality?: AirQuality
}

export interface CurrentWeather {
    temperature: number
    feelsLike: number
    humidity: number
    windSpeed: number
    pressure: number
    visibility: number
    weatherCode: number
    description: string
    icon: string
}

export interface HourlyForecast {
    time: string
    temperature: number
    precipitationProbability: number
    weatherCode: number
    description: string
    icon: string
}

export interface DailyForecast {
    date: string
    tempMin: number
    tempMax: number
    sunrise: string
    sunset: string
    uvIndex: number
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
    compactCity?: boolean
    compactFeelsLike?: boolean
    compactDescription?: boolean
    compactAlign?: 'left' | 'center' | 'right'
    showCurrent?: boolean
    showDetails?: boolean
    showHourly?: boolean
    hourlyCount?: number
    hourlySlots?: number
    showForecast?: boolean
    forecastDays?: number
    showSunrise?: boolean
    showSunset?: boolean
    showWind?: boolean
    showHumidity?: boolean
    showPressure?: boolean
    showUV?: boolean
    showVisibility?: boolean
    showAirQuality?: boolean
    showGraph?: boolean
    graphHours?: number
    graphTempColor?: string
    graphRainColor?: string
    graphBgColor?: string
    graphHeight?: number
    graphShowTemp?: boolean
    graphShowRain?: boolean
}
