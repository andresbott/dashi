<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Slider from 'primevue/slider'
import { useGeocode } from '@/composables/useWeather'
import type { WeatherWidgetConfig } from '@/types/weather'

const props = defineProps<{
    config: WeatherWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: WeatherWidgetConfig]
}>()

function initFromConfig(cfg: WeatherWidgetConfig | null) {
    return {
        showCurrent: cfg?.showCurrent ?? true,
        showDetails: cfg?.showDetails ?? true,
        showForecast: cfg?.showForecast ?? true,
        forecastDays: cfg?.forecastDays ?? 7,
        showHourly: cfg?.showHourly ?? false,
        hourlyCount: cfg?.hourlyCount ?? 12,
        hourlySlots: cfg?.hourlySlots ?? 6,
        showSunrise: cfg?.showSunrise ?? false,
        showSunset: cfg?.showSunset ?? false,
        showWind: cfg?.showWind ?? false,
        showHumidity: cfg?.showHumidity ?? false,
        showPressure: cfg?.showPressure ?? false,
        showUV: cfg?.showUV ?? false,
        showVisibility: cfg?.showVisibility ?? false,
        showAirQuality: cfg?.showAirQuality ?? false,
    }
}

const init = initFromConfig(props.config)
const showCurrent = ref(init.showCurrent)
const showDetails = ref(init.showDetails)
const showForecast = ref(init.showForecast)
const forecastDays = ref(init.forecastDays)
const showHourly = ref(init.showHourly)
const hourlyCount = ref(init.hourlyCount)
const hourlySlots = ref(init.hourlySlots)
const showSunrise = ref(init.showSunrise)
const showSunset = ref(init.showSunset)
const showWind = ref(init.showWind)
const showHumidity = ref(init.showHumidity)
const showPressure = ref(init.showPressure)
const showUV = ref(init.showUV)
const showVisibility = ref(init.showVisibility)
const showAirQuality = ref(init.showAirQuality)
const initialized = ref(props.config !== null)

watch(() => props.config, (val, oldVal) => {
    if (!val) return
    if (initialized.value && oldVal &&
        val.city === oldVal.city && val.latitude === oldVal.latitude && val.longitude === oldVal.longitude) {
        return
    }
    const s = initFromConfig(val)
    showCurrent.value = s.showCurrent
    showDetails.value = s.showDetails
    showForecast.value = s.showForecast
    forecastDays.value = s.forecastDays
    showHourly.value = s.showHourly
    hourlyCount.value = s.hourlyCount
    hourlySlots.value = s.hourlySlots
    showSunrise.value = s.showSunrise
    showSunset.value = s.showSunset
    showWind.value = s.showWind
    showHumidity.value = s.showHumidity
    showPressure.value = s.showPressure
    showUV.value = s.showUV
    showVisibility.value = s.showVisibility
    showAirQuality.value = s.showAirQuality
    initialized.value = true
})

const emitUpdate = () => {
    if (!props.config) return
    emit('update:config', {
        ...props.config,
        showCurrent: showCurrent.value,
        showDetails: showDetails.value,
        showForecast: showForecast.value,
        forecastDays: forecastDays.value,
        showHourly: showHourly.value,
        hourlyCount: hourlyCount.value,
        hourlySlots: hourlySlots.value,

        showSunrise: showSunrise.value,
        showSunset: showSunset.value,
        showWind: showWind.value,
        showHumidity: showHumidity.value,
        showPressure: showPressure.value,
        showUV: showUV.value,
        showVisibility: showVisibility.value,
        showAirQuality: showAirQuality.value,
    })
}

const searchQuery = ref('')
const debouncedQuery = ref('')

let debounceTimer: ReturnType<typeof setTimeout> | undefined
watch(searchQuery, (val) => {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
        debouncedQuery.value = val
    }, 400)
})

const { data: locations, isLoading: isSearching } = useGeocode(debouncedQuery)

const selectLocation = (loc: { name: string; country: string; latitude: number; longitude: number }) => {
    emit('update:config', {
        city: `${loc.name}, ${loc.country}`,
        latitude: loc.latitude,
        longitude: loc.longitude,
        showCurrent: showCurrent.value,
        showDetails: showDetails.value,
        showForecast: showForecast.value,
        forecastDays: forecastDays.value,
        showHourly: showHourly.value,
        hourlyCount: hourlyCount.value,
        hourlySlots: hourlySlots.value,

        showSunrise: showSunrise.value,
        showSunset: showSunset.value,
        showWind: showWind.value,
        showHumidity: showHumidity.value,
        showPressure: showPressure.value,
        showUV: showUV.value,
        showVisibility: showVisibility.value,
        showAirQuality: showAirQuality.value,
    })
    searchQuery.value = ''
    debouncedQuery.value = ''
}
</script>

<template>
    <div class="weather-config">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Search city</label>
            <div class="flex gap-2">
                <InputText
                    v-model="searchQuery"
                    placeholder="Type a city name..."
                    class="flex-grow-1"
                />
                <i v-if="isSearching" class="ti ti-loader-2 weather-config-spinner" style="align-self: center" />
            </div>
        </div>

        <div v-if="locations && locations.length > 0" class="weather-config-results">
            <Button
                v-for="(loc, i) in locations"
                :key="i"
                class="weather-config-result"
                text
                severity="secondary"
                @click="selectLocation(loc)"
            >
                <span class="font-semibold">{{ loc.name }}</span>
                <span class="text-sm ml-1" style="color: var(--p-text-muted-color)">{{ loc.country }}</span>
                <span class="text-xs ml-auto" style="color: var(--p-text-muted-color)">
                    {{ loc.latitude.toFixed(2) }}, {{ loc.longitude.toFixed(2) }}
                </span>
            </Button>
        </div>

        <div v-if="config" class="weather-config-selected mt-3">
            <label class="text-sm font-semibold">Current location</label>
            <div class="text-sm">
                <i class="ti ti-map-pin" />
                {{ config.city }}
                <span style="color: var(--p-text-muted-color)">
                    ({{ config.latitude.toFixed(4) }}, {{ config.longitude.toFixed(4) }})
                </span>
            </div>
        </div>

        <div class="flex flex-column gap-3 mt-3">
            <label class="text-sm font-semibold">Display</label>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showCurrent" :binary="true" inputId="weatherShowCurrent" @update:modelValue="emitUpdate" />
                    <label for="weatherShowCurrent" class="text-sm">Show current weather</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showDetails" :binary="true" inputId="weatherDetails" @update:modelValue="emitUpdate" />
                    <label for="weatherDetails" class="text-sm">Show details (feels like, humidity, wind)</label>
                </div>
                <div class="config-divider" />
                <label class="text-sm font-semibold">Extra Info</label>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showSunrise" :binary="true" inputId="weatherSunrise" @update:modelValue="emitUpdate" />
                    <label for="weatherSunrise" class="text-sm">Sunrise</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showSunset" :binary="true" inputId="weatherSunset" @update:modelValue="emitUpdate" />
                    <label for="weatherSunset" class="text-sm">Sunset</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showWind" :binary="true" inputId="weatherWind" @update:modelValue="emitUpdate" />
                    <label for="weatherWind" class="text-sm">Wind speed</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showHumidity" :binary="true" inputId="weatherHumidity" @update:modelValue="emitUpdate" />
                    <label for="weatherHumidity" class="text-sm">Humidity</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showPressure" :binary="true" inputId="weatherPressure" @update:modelValue="emitUpdate" />
                    <label for="weatherPressure" class="text-sm">Pressure</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showUV" :binary="true" inputId="weatherUV" @update:modelValue="emitUpdate" />
                    <label for="weatherUV" class="text-sm">UV Index</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showVisibility" :binary="true" inputId="weatherVisibility" @update:modelValue="emitUpdate" />
                    <label for="weatherVisibility" class="text-sm">Visibility</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showAirQuality" :binary="true" inputId="weatherAirQuality" @update:modelValue="emitUpdate" />
                    <label for="weatherAirQuality" class="text-sm">Air Quality (AQI)</label>
                </div>
                <div class="config-divider" />
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showHourly" :binary="true" inputId="weatherHourly" @update:modelValue="emitUpdate" />
                    <label for="weatherHourly" class="text-sm">Show hourly forecast</label>
                </div>
                <div v-if="showHourly" class="flex flex-column gap-2 ml-4">
                    <div class="flex flex-column gap-1">
                        <label class="text-sm">Hours ahead: {{ hourlyCount }}</label>
                        <Slider v-model="hourlyCount" :min="1" :max="24" :step="1" class="w-full" @slideend="emitUpdate" />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="text-sm">Slots to show: {{ hourlySlots }}</label>
                        <Slider v-model="hourlySlots" :min="1" :max="12" :step="1" class="w-full" @slideend="emitUpdate" />
                    </div>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showForecast" :binary="true" inputId="weatherForecast" @update:modelValue="emitUpdate" />
                    <label for="weatherForecast" class="text-sm">Show daily forecast</label>
                </div>
                <div v-if="showForecast" class="flex flex-column gap-1 ml-4">
                    <label class="text-sm">Days to show: {{ forecastDays }}</label>
                    <Slider v-model="forecastDays" :min="1" :max="7" :step="1" class="w-full" @slideend="emitUpdate" />
                </div>
        </div>
    </div>
</template>

<style scoped>
.weather-config-results {
    display: flex;
    flex-direction: column;
    margin-top: 0.5rem;
    border: 1px solid var(--p-surface-200);
    border-radius: 6px;
    overflow: hidden;
}

.weather-config-result {
    display: flex;
    align-items: center;
    width: 100%;
    text-align: left;
    border-radius: 0;
    padding: 0.5rem 0.75rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.weather-config-spinner {
    animation: spin 1s linear infinite;
}

.config-divider {
    border-top: 1px solid var(--p-surface-200, #e5e7eb);
    margin: 0.25rem 0;
}
</style>
