<script setup lang="ts">
import { computed } from 'vue'
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

const isCompact = computed(() => config.value?.compact ?? false)
const showCurrent = computed(() => config.value?.showCurrent ?? true)
const showDetails = computed(() => config.value?.showDetails ?? true)
const showForecast = computed(() => config.value?.showForecast ?? true)
const forecastDays = computed(() => config.value?.forecastDays ?? 7)
const showHourly = computed(() => config.value?.showHourly ?? false)
const hourlyCount = computed(() => config.value?.hourlyCount ?? 12)
const hourlySlots = computed(() => config.value?.hourlySlots ?? 6)
const iconTheme = computed(() => config.value?.iconTheme ?? 'default')

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

const cityName = computed(() => config.value?.city?.split(',')[0] ?? '')
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

        <template v-else-if="weather">
            <!-- Compact layout -->
            <div v-if="isCompact" class="weather-compact-content">
                <span class="weather-compact-icon">
                    <WeatherIcon :icon-name="weather.current.icon" :theme-name="iconTheme" :themes="themes" />
                </span>
                <div class="weather-compact-info">
                    <div class="weather-compact-top">
                        <span class="weather-compact-city">{{ cityName }}</span>
                        <span class="weather-compact-temp">{{ formatTemp(weather.current.temperature) }}</span>
                    </div>
                    <div class="weather-compact-desc">{{ weather.current.description }}</div>
                </div>
            </div>

            <!-- Full layout -->
            <template v-else>
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
            </template>
        </template>
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
    color: var(--p-primary-color);
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

/* Compact layout */
.weather-compact-content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.weather-compact-icon {
    font-size: 2.5rem;
    color: var(--p-primary-color);
    flex-shrink: 0;
}

.weather-compact-info {
    flex: 1;
    min-width: 0;
}

.weather-compact-top {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
}

.weather-compact-city {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.weather-compact-temp {
    font-size: 1.25rem;
    font-weight: 700;
    flex-shrink: 0;
}

.weather-compact-desc {
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}
</style>
