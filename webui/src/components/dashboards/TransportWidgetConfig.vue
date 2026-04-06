<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Button from 'primevue/button'
import { useStationSearch } from '@/composables/useTransport'
import type { TransportWidgetConfig } from '@/types/transport'

const props = defineProps<{
    config: TransportWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: TransportWidgetConfig]
}>()

const editLimit = ref(props.config?.limit ?? 5)

watch(() => props.config, (val) => {
    if (val) {
        editLimit.value = val.limit ?? 5
    }
})

const emitUpdate = () => {
    if (!props.config) return
    emit('update:config', {
        ...props.config,
        limit: editLimit.value,
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

const { data: stations, isLoading: isSearching } = useStationSearch(debouncedQuery)

const selectStation = (station: { id: string; name: string }) => {
    emit('update:config', {
        stationId: station.id,
        stationName: station.name,
        limit: editLimit.value,
    })
    searchQuery.value = ''
    debouncedQuery.value = ''
}
</script>

<template>
    <div class="transport-config">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Search stop</label>
            <div class="flex gap-2">
                <InputText
                    v-model="searchQuery"
                    placeholder="Type a stop name..."
                    class="flex-grow-1"
                />
                <i v-if="isSearching" class="ti ti-loader-2 transport-config-spinner" style="align-self: center" />
            </div>
        </div>

        <div v-if="stations && stations.length > 0" class="transport-config-results">
            <Button
                v-for="(station, i) in stations"
                :key="i"
                class="transport-config-result"
                text
                severity="secondary"
                @click="selectStation(station)"
            >
                <span class="font-semibold">{{ station.name }}</span>
            </Button>
        </div>

        <div v-if="config && config.stationId" class="transport-config-selected mt-3">
            <label class="text-sm font-semibold">Current stop</label>
            <div class="text-sm">
                <i class="ti ti-bus-stop" />
                {{ config.stationName }}
            </div>
        </div>

        <div class="flex flex-column gap-1 mt-3">
            <label class="text-sm font-semibold">Departures to show</label>
            <InputNumber v-model="editLimit" :min="1" :max="15" showButtons @update:modelValue="emitUpdate" />
        </div>
    </div>
</template>

<style scoped>
.transport-config-results {
    display: flex;
    flex-direction: column;
    margin-top: 0.5rem;
    border: 1px solid var(--p-surface-200);
    border-radius: 6px;
    overflow: hidden;
}

.transport-config-result {
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

.transport-config-spinner {
    animation: spin 1s linear infinite;
}
</style>
