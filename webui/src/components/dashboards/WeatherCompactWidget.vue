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

const compactCity = computed(() => config.value?.compactCity ?? true)
const compactFeelsLike = computed(() => config.value?.compactFeelsLike ?? false)
const compactDescription = computed(() => config.value?.compactDescription ?? true)
const compactAlign = computed(() => config.value?.compactAlign ?? 'left')
const iconTheme = inject(DASHBOARD_THEME, ref('default'))

const lat = computed(() => config.value?.latitude)
const lon = computed(() => config.value?.longitude)

const { data: weather, isLoading, isError } = useWeather(lat, lon)
const { data: themes } = useThemes()

const formatTemp = (temp: number) => `${Math.round(temp)}°`
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
            <div
                class="weather-compact-content"
                :style="{
                    justifyContent: compactAlign === 'center' ? 'center' : compactAlign === 'right' ? 'flex-end' : 'flex-start',
                }"
            >
                <span class="weather-compact-icon">
                    <WeatherIcon :icon-name="weather.current.icon" :theme-name="iconTheme" :themes="themes" />
                </span>
                <div class="weather-compact-info">
                    <div class="weather-compact-top">
                        <span v-if="compactCity" class="weather-compact-city">{{ cityName }}</span>
                        <span class="weather-compact-temp">{{ formatTemp(weather.current.temperature) }}</span>
                    </div>
                    <div v-if="compactFeelsLike" class="weather-compact-feels">Feels like {{ formatTemp(weather.current.feelsLike) }}</div>
                    <div v-if="compactDescription" class="weather-compact-desc">{{ weather.current.description }}</div>
                </div>
            </div>
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

.weather-compact-content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.weather-compact-icon {
    font-size: 2.5rem;
    flex-shrink: 0;
}

.weather-compact-info {
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

.weather-compact-feels {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
}

.weather-compact-desc {
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

</style>
