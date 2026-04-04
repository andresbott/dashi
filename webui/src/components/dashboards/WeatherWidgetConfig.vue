<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'
import Slider from 'primevue/slider'
import { useGeocode } from '@/composables/useWeather'
import { useThemes } from '@/composables/useThemes'
import type { WeatherWidgetConfig } from '@/types/weather'

const props = defineProps<{
    config: WeatherWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: WeatherWidgetConfig]
}>()

const { data: themes } = useThemes()

function initFromConfig(cfg: WeatherWidgetConfig | null) {
    return {
        selectedLayout: cfg?.compact ? 'compact' : 'normal',
        showCurrent: cfg?.showCurrent ?? true,
        showDetails: cfg?.showDetails ?? true,
        showForecast: cfg?.showForecast ?? true,
        forecastDays: cfg?.forecastDays ?? 7,
        showHourly: cfg?.showHourly ?? false,
        hourlyCount: cfg?.hourlyCount ?? 12,
        hourlySlots: cfg?.hourlySlots ?? 6,
        iconTheme: cfg?.iconTheme ?? 'default',
    }
}

const init = initFromConfig(props.config)
const selectedLayout = ref(init.selectedLayout)
const showCurrent = ref(init.showCurrent)
const showDetails = ref(init.showDetails)
const showForecast = ref(init.showForecast)
const forecastDays = ref(init.forecastDays)
const showHourly = ref(init.showHourly)
const hourlyCount = ref(init.hourlyCount)
const hourlySlots = ref(init.hourlySlots)
const iconTheme = ref(init.iconTheme)
const initialized = ref(props.config !== null)

watch(() => props.config, (val, oldVal) => {
    if (!val) return
    if (initialized.value && oldVal &&
        val.city === oldVal.city && val.latitude === oldVal.latitude && val.longitude === oldVal.longitude) {
        return
    }
    const s = initFromConfig(val)
    selectedLayout.value = s.selectedLayout
    showCurrent.value = s.showCurrent
    showDetails.value = s.showDetails
    showForecast.value = s.showForecast
    forecastDays.value = s.forecastDays
    showHourly.value = s.showHourly
    hourlyCount.value = s.hourlyCount
    hourlySlots.value = s.hourlySlots
    iconTheme.value = s.iconTheme
    initialized.value = true
})

const themeOptions = ref<{ label: string; value: string }[]>([])
watch(themes, (val) => {
    if (val) {
        themeOptions.value = val.map(t => ({ label: t.name, value: t.name }))
    }
}, { immediate: true })

const emitUpdate = () => {
    if (!props.config) return
    emit('update:config', {
        ...props.config,
        compact: selectedLayout.value === 'compact',
        showCurrent: showCurrent.value,
        showDetails: showDetails.value,
        showForecast: showForecast.value,
        forecastDays: forecastDays.value,
        showHourly: showHourly.value,
        hourlyCount: hourlyCount.value,
        hourlySlots: hourlySlots.value,
        iconTheme: iconTheme.value,
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
        compact: selectedLayout.value === 'compact',
        showCurrent: showCurrent.value,
        showDetails: showDetails.value,
        showForecast: showForecast.value,
        forecastDays: forecastDays.value,
        showHourly: showHourly.value,
        hourlyCount: hourlyCount.value,
        hourlySlots: hourlySlots.value,
        iconTheme: iconTheme.value,
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
            <div class="flex flex-column gap-1">
                <label class="text-sm">Layout</label>
                <Select
                    v-model="selectedLayout"
                    :options="[
                        { label: 'Normal', value: 'normal' },
                        { label: 'Compact', value: 'compact' },
                    ]"
                    optionLabel="label"
                    optionValue="value"
                    class="w-full"
                    @update:modelValue="emitUpdate"
                />
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm">Icon theme</label>
                <Select
                    v-model="iconTheme"
                    :options="themeOptions"
                    optionLabel="label"
                    optionValue="value"
                    class="w-full"
                    @update:modelValue="emitUpdate"
                />
            </div>
            <template v-if="selectedLayout === 'normal'">
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showCurrent" :binary="true" inputId="weatherShowCurrent" @update:modelValue="emitUpdate" />
                    <label for="weatherShowCurrent" class="text-sm">Show current weather</label>
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="showDetails" :binary="true" inputId="weatherDetails" @update:modelValue="emitUpdate" />
                    <label for="weatherDetails" class="text-sm">Show details (feels like, humidity, wind)</label>
                </div>
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
            </template>
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
</style>
