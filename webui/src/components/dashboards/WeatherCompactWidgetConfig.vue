<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'
import { useGeocode } from '@/composables/useWeather'
import type { WeatherWidgetConfig } from '@/types/weather'

const props = defineProps<{
    config: WeatherWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: WeatherWidgetConfig]
}>()

const compactCity = ref(props.config?.compactCity ?? true)
const compactFeelsLike = ref(props.config?.compactFeelsLike ?? false)
const compactDescription = ref(props.config?.compactDescription ?? true)
const compactAlign = ref(props.config?.compactAlign ?? 'left')
const alignOptions = [
    { label: 'Left', value: 'left' },
    { label: 'Center', value: 'center' },
    { label: 'Right', value: 'right' },
]
watch(() => props.config, (val) => {
    if (val) {
        compactCity.value = val.compactCity ?? true
        compactFeelsLike.value = val.compactFeelsLike ?? false
        compactDescription.value = val.compactDescription ?? true
        compactAlign.value = val.compactAlign ?? 'left'
    }
})

const emitUpdate = () => {
    if (!props.config) return
    emit('update:config', {
        ...props.config,
        compact: true,
        compactCity: compactCity.value,
        compactFeelsLike: compactFeelsLike.value,
        compactDescription: compactDescription.value,
        compactAlign: compactAlign.value,
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
        compact: true,
        compactCity: compactCity.value,
        compactFeelsLike: compactFeelsLike.value,
        compactDescription: compactDescription.value,
        compactAlign: compactAlign.value,
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

        <div v-if="config && config.latitude != null" class="weather-config-selected mt-3">
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
                <Checkbox v-model="compactCity" :binary="true" inputId="compactCity" @update:modelValue="emitUpdate" />
                <label for="compactCity" class="text-sm">Show location</label>
            </div>
            <div class="flex align-items-center gap-2">
                <Checkbox v-model="compactFeelsLike" :binary="true" inputId="compactFeelsLike" @update:modelValue="emitUpdate" />
                <label for="compactFeelsLike" class="text-sm">Show feels like</label>
            </div>
            <div class="flex align-items-center gap-2">
                <Checkbox v-model="compactDescription" :binary="true" inputId="compactDescription" @update:modelValue="emitUpdate" />
                <label for="compactDescription" class="text-sm">Show description</label>
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Alignment</label>
                <Select v-model="compactAlign" :options="alignOptions" optionLabel="label" optionValue="value" @update:modelValue="emitUpdate" />
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

</style>
