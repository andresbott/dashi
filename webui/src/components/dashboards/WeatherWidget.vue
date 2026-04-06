<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import { DASHBOARD_THEME } from '@/lib/injectionKeys'
import { useWeather } from '@/composables/useWeather'
import { useThemes } from '@/composables/useThemes'
import WeatherIcon from '@/components/dashboards/WeatherIcon.vue'
import type { WeatherWidgetConfig } from '@/types/weather'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<WeatherWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    const c = props.widget.config as unknown as WeatherWidgetConfig
    if (!c.latitude || !c.longitude) return null
    return c
})

const showCurrent = computed(() => config.value?.showCurrent ?? true)
const showDetails = computed(() => config.value?.showDetails ?? true)
const showForecast = computed(() => config.value?.showForecast ?? true)
const forecastDays = computed(() => config.value?.forecastDays ?? 7)
const showHourly = computed(() => config.value?.showHourly ?? false)
const hourlyCount = computed(() => config.value?.hourlyCount ?? 12)
const hourlySlots = computed(() => config.value?.hourlySlots ?? 6)
const iconTheme = inject(DASHBOARD_THEME, ref('default'))

const showGraph = computed(() => config.value?.showGraph ?? false)
const graphHours = computed(() => config.value?.graphHours ?? 24)
const graphTempColor = computed(() => config.value?.graphTempColor ?? '#FF8C42')
const graphRainColor = computed(() => config.value?.graphRainColor ?? '#4A90D9')
const graphHeight = computed(() => config.value?.graphHeight ?? 250)
const graphShowTemp = computed(() => config.value?.graphShowTemp ?? true)
const graphShowRain = computed(() => config.value?.graphShowRain ?? true)

const showSunrise = computed(() => config.value?.showSunrise ?? false)
const showSunset = computed(() => config.value?.showSunset ?? false)
const showWind = computed(() => config.value?.showWind ?? false)
const showHumidity = computed(() => config.value?.showHumidity ?? false)
const showPressure = computed(() => config.value?.showPressure ?? false)
const showUV = computed(() => config.value?.showUV ?? false)
const showVisibility = computed(() => config.value?.showVisibility ?? false)
const showAirQuality = computed(() => config.value?.showAirQuality ?? false)

const hasExtraInfo = computed(() =>
    showSunrise.value || showSunset.value || showWind.value || showHumidity.value ||
    showPressure.value || showUV.value || showVisibility.value || showAirQuality.value
)

const formatTime = (isoTime: string) => {
    if (!isoTime) return ''
    const parts = isoTime.split('T')
    return parts.length > 1 ? parts[1] : isoTime
}

const lat = computed(() => config.value?.latitude)
const lon = computed(() => config.value?.longitude)

const { data: weather, isLoading, isError } = useWeather(lat, lon)
const { data: themes } = useThemes()

const formatTemp = (temp: number) => `${Math.round(temp)}°`

const formatDay = (dateStr: string) => {
    const date = new Date(dateStr + 'T00:00:00')
    return date.toLocaleDateString(undefined, { weekday: 'short' })
}

const formatHour = (timeStr: string) => {
    const date = new Date(timeStr.replace('T', ' '))
    return date.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
}

// --- Graph SVG data ---
const graphData = computed(() => {
    if (!weather.value?.hourly || weather.value.hourly.length < 2) return null
    const points = weather.value.hourly.slice(0, graphHours.value)
    if (points.length < 2) return null

    const temps = points.map(p => p.temperature)
    const min = Math.min(...temps)
    const max = Math.max(...temps)
    const range = max - min || 1
    const padding = range * 0.15
    const pMin = min - padding
    const pRange = range + padding * 2

    const w = 400
    const h = 120

    // Temperature line
    const coords = points.map((p, i) => ({
        x: (i / (points.length - 1)) * w,
        y: h - ((p.temperature - pMin) / pRange) * h,
    }))

    let tempPath = `M ${coords[0].x},${coords[0].y}`
    for (let i = 1; i < coords.length; i++) {
        tempPath += ` L ${coords[i].x},${coords[i].y}`
    }
    const tempFill = `${tempPath} L ${w},${h} L 0,${h} Z`

    // Rain bars
    const barWidth = w / points.length
    const rainBars = points.map((p, i) => ({
        x: i * barWidth,
        width: barWidth,
        height: (p.precipitationProbability / 100) * h,
        y: h - (p.precipitationProbability / 100) * h,
    }))

    // Temp range labels (same as image dashboard: max top-left, min bottom-left)
    const tempMax = `${Math.round(max)}°`
    const tempMin = `${Math.round(min)}°`

    // Time labels — every ~3 hours (same spacing as image dashboard)
    const step = Math.max(1, Math.floor(points.length / 8) + 1)
    const timeLabels: string[] = []
    for (let i = 0; i < points.length; i += step) {
        timeLabels.push(formatHour(points[i].time))
    }

    return { tempPath, tempFill, rainBars, tempMax, tempMin, timeLabels }
})

const hourlySlice = computed(() => {
    if (!weather.value?.hourly) return []
    const all = weather.value.hourly.slice(0, hourlyCount.value)
    if (all.length === 0) return []
    const slots = Math.min(hourlySlots.value, all.length)
    if (slots >= all.length) return all
    const result = []
    for (let i = 0; i < slots; i++) {
        const idx = Math.round(((i + 1) * all.length) / slots) - 1
        result.push(all[idx])
    }
    return result
})

</script>

<template>
    <div class="weather-widget">
        <div v-if="!config" class="weather-empty">
            <i class="ti ti-map-pin-question" />
            <span>Set a location in edit mode</span>
        </div>

        <div v-else-if="isLoading" class="weather-empty">
            <i class="ti ti-loader-2 weather-spinner" />
        </div>

        <div v-else-if="isError" class="weather-empty">
            <i class="ti ti-cloud-off" />
            <span>Failed to load weather</span>
        </div>

        <div v-else-if="weather">
                <div v-if="showCurrent" class="weather-current">
                    <div class="weather-main">
                        <span class="weather-icon">
                            <WeatherIcon :icon-name="weather.current.icon" :theme-name="iconTheme" :themes="themes" />
                        </span>
                        <span class="weather-temp">{{ formatTemp(weather.current.temperature) }}</span>
                    </div>
                    <div class="weather-city">{{ config.city }}</div>
                    <div class="weather-desc">{{ weather.current.description }}</div>
                    <div v-if="showDetails" class="weather-details">
                        <span>Feels like {{ formatTemp(weather.current.feelsLike) }}</span>
                        <span>Humidity {{ weather.current.humidity }}%</span>
                        <span>Wind {{ weather.current.windSpeed }} km/h</span>
                    </div>
                </div>
                <div v-if="hasExtraInfo && weather" class="weather-extra-info">
                    <div v-if="showSunrise && weather.forecast.length" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="sunrise" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Sunrise</span>
                        <span class="extra-info-value">{{ formatTime(weather.forecast[0].sunrise) }}</span>
                    </div>
                    <div v-if="showSunset && weather.forecast.length" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="sunset" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Sunset</span>
                        <span class="extra-info-value">{{ formatTime(weather.forecast[0].sunset) }}</span>
                    </div>
                    <div v-if="showWind" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="wind" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Wind</span>
                        <span class="extra-info-value">{{ Math.round(weather.current.windSpeed) }} km/h</span>
                    </div>
                    <div v-if="showHumidity" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="humidity" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Humidity</span>
                        <span class="extra-info-value">{{ weather.current.humidity }}%</span>
                    </div>
                    <div v-if="showPressure" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="pressure" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Pressure</span>
                        <span class="extra-info-value">{{ Math.round(weather.current.pressure) }} hPa</span>
                    </div>
                    <div v-if="showUV && weather.forecast.length" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="uv-index" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">UV Index</span>
                        <span class="extra-info-value">{{ weather.forecast[0].uvIndex.toFixed(1) }}</span>
                    </div>
                    <div v-if="showVisibility" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="visibility" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">Visibility</span>
                        <span class="extra-info-value">{{ Math.round(weather.current.visibility) }} km</span>
                    </div>
                    <div v-if="showAirQuality && weather.airQuality" class="extra-info-item">
                        <span class="extra-info-icon"><WeatherIcon icon-name="air-quality" :theme-name="iconTheme" :themes="themes" /></span>
                        <span class="extra-info-title">AQI</span>
                        <span class="extra-info-value">{{ weather.airQuality.europeanAqi }}</span>
                    </div>
                </div>
                <div v-if="showHourly && hourlySlice.length" class="weather-hourly">
                    <div
                        v-for="hour in hourlySlice"
                        :key="hour.time"
                        class="weather-hourly-item"
                    >
                        <span class="hourly-time">{{ formatHour(hour.time) }}</span>
                        <WeatherIcon :icon-name="hour.icon" :theme-name="iconTheme" :themes="themes" />
                        <span class="hourly-temp">{{ formatTemp(hour.temperature) }}</span>
                    </div>
                </div>
                <div v-if="showGraph && graphData" class="weather-graph-container">
                    <div v-if="graphShowTemp || graphShowRain" class="weather-graph-top">
                        <span>{{ graphShowTemp ? graphData.tempMax : '' }}</span>
                        <span>{{ graphShowRain ? '100%' : '' }}</span>
                    </div>
                    <div class="weather-graph-img" :style="{ height: graphHeight + 'px' }">
                        <svg viewBox="0 0 400 120" preserveAspectRatio="none" class="weather-graph-svg">
                            <defs>
                                <linearGradient :id="'wg-temp-' + widget.id" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="0%" :stop-color="graphTempColor" stop-opacity="0.3" />
                                    <stop offset="100%" :stop-color="graphTempColor" stop-opacity="0" />
                                </linearGradient>
                            </defs>
                            <g v-if="graphShowRain">
                                <rect
                                    v-for="(bar, i) in graphData.rainBars"
                                    :key="'rain-' + i"
                                    :x="bar.x"
                                    :y="bar.y"
                                    :width="bar.width"
                                    :height="bar.height"
                                    :fill="graphRainColor"
                                    fill-opacity="0.25"
                                />
                            </g>
                            <path :d="graphData.tempFill" :fill="'url(#wg-temp-' + widget.id + ')'" />
                            <path :d="graphData.tempPath" fill="none" :stroke="graphTempColor" stroke-width="2" vector-effect="non-scaling-stroke" />
                        </svg>
                    </div>
                    <div v-if="graphShowTemp || graphShowRain" class="weather-graph-top">
                        <span>{{ graphShowTemp ? graphData.tempMin : '' }}</span>
                        <span>{{ graphShowRain ? '0%' : '' }}</span>
                    </div>
                    <div class="weather-graph-bottom">
                        <span v-for="(label, i) in graphData.timeLabels" :key="'tml-' + i">{{ label }}</span>
                    </div>
                </div>
                <div v-if="showForecast" class="weather-forecast">
                    <div
                        v-for="day in weather.forecast.slice(0, forecastDays)"
                        :key="day.date"
                        class="weather-forecast-day"
                    >
                        <span class="forecast-day-name">{{ formatDay(day.date) }}</span>
                        <WeatherIcon :icon-name="day.icon" :theme-name="iconTheme" :themes="themes" />
                        <span class="forecast-temps">
                            {{ formatTemp(day.tempMax) }}
                            <span class="forecast-temp-min">{{ formatTemp(day.tempMin) }}</span>
                        </span>
                    </div>
                </div>
        </div>
    </div>
</template>

<style scoped>
.weather-widget {
    padding: 0.5rem;
    min-height: 60px;
}

.weather-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.weather-empty .ti {
    font-size: 1.5rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.weather-spinner {
    animation: spin 1s linear infinite;
    font-size: 1.5rem;
}

/* Full layout */
.weather-current {
    text-align: center;
    margin-bottom: 0.75rem;
}

.weather-main {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
}

.weather-icon {
    font-size: 2rem;
}

.weather-temp {
    font-size: 2rem;
    font-weight: 700;
}

.weather-city {
    font-weight: 600;
    margin-top: 0.25rem;
}

.weather-desc {
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.weather-details {
    display: flex;
    justify-content: center;
    gap: 1rem;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
    margin-top: 0.5rem;
}

.weather-extra-info {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
    gap: 0.5rem;
    border-top: 1px solid var(--p-surface-200);
    padding-top: 0.5rem;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
}

.extra-info-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.125rem;
    text-align: center;
}

.extra-info-icon {
    font-size: 1.25rem;
}

.extra-info-title {
    font-size: 0.65rem;
    color: var(--p-text-muted-color);
    text-transform: uppercase;
    letter-spacing: 0.03em;
}

.extra-info-value {
    font-size: 0.8rem;
    font-weight: 600;
}

.weather-graph-container {
    border-top: 1px solid var(--p-surface-200);
    padding-top: 0.5rem;
    margin-bottom: 0.5rem;
}

.weather-graph-top {
    display: flex;
    justify-content: space-between;
    font-size: 0.8em;
    margin-bottom: 2px;
    color: var(--p-text-muted-color);
}

.weather-graph-img {
    width: 100%;
}

.weather-graph-svg {
    display: block;
    width: 100%;
    height: 100%;
}

.weather-graph-bottom {
    display: flex;
    justify-content: space-between;
    font-size: 0.75em;
    color: var(--p-text-muted-color);
    margin-top: 2px;
}

.weather-forecast {
    display: flex;
    justify-content: space-between;
    gap: 0.25rem;
    border-top: 1px solid var(--p-surface-200);
    padding-top: 0.5rem;
}

.weather-forecast-day {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
    flex: 1;
}

.forecast-day-name {
    font-weight: 600;
}

.forecast-temp-min {
    color: var(--p-text-muted-color);
}

.weather-hourly {
    display: flex;
    justify-content: space-between;
    gap: 0.25rem;
    border-top: 1px solid var(--p-surface-200);
    padding-top: 0.5rem;
    margin-bottom: 0.5rem;
}

.weather-hourly-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
    flex: 1;
}

.hourly-time {
    font-weight: 600;
}

.hourly-temp {
    color: var(--p-text-muted-color);
}
</style>
