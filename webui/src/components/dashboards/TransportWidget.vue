<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useTransport } from '@/composables/useTransport'
import type { TransportWidgetConfig } from '@/types/transport'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<TransportWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as TransportWidgetConfig
})

const stationId = computed(() => config.value?.stationId)
const stationName = computed(() => config.value?.stationName ?? '')
const limit = computed(() => config.value?.limit ?? 5)

const { data: departures, isLoading, isError, isFetching, refetch } = useTransport(stationId, limit)

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
    timer = setInterval(() => { now.value = Date.now() }, 10000)
})
onUnmounted(() => { clearInterval(timer) })

const formatDeparture = (expected: string) => {
    const diff = new Date(expected).getTime() - now.value
    const mins = Math.round(diff / 60000)
    if (mins < 0) return "0'"
    if (mins > 60) {
        const d = new Date(now.value + diff)
        return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0')
    }
    return mins + "'"
}

const categoryIcon = (cat: string) => {
    switch (cat) {
        case 'T': return 'ti-train'
        case 'B': return 'ti-bus'
        default: return 'ti-train'
    }
}
</script>

<template>
    <div class="transport-widget">
        <div v-if="!config?.stationId" class="transport-empty">
            <i class="ti ti-bus-stop" />
            <span>Set a stop in edit mode</span>
        </div>

        <div v-else-if="isLoading" class="transport-empty">
            <i class="ti ti-loader-2 transport-spinner" />
        </div>

        <div v-else-if="isError" class="transport-empty">
            <i class="ti ti-alert-circle" />
            <span>Failed to load departures</span>
        </div>

        <div v-else-if="departures" class="transport-content">
            <div class="transport-header">
                <span>{{ stationName }}</span>
                <button class="transport-refresh" :class="{ 'transport-spinner': isFetching }" @click.prevent="refetch()" title="Refresh">
                    <i class="ti ti-refresh" />
                </button>
            </div>
            <div v-if="departures.length === 0" class="transport-empty">
                <span>No departures</span>
            </div>
            <div v-else class="transport-board">
                <div v-for="(dep, i) in departures" :key="i" class="transport-row">
                    <span class="transport-line">
                        <i :class="'ti ' + categoryIcon(dep.category)" />
                        {{ dep.number }}
                    </span>
                    <span class="transport-dest">{{ dep.destination }}</span>
                    <span class="transport-mins" :class="{ 'transport-delayed': dep.delay > 0 }">
                        {{ formatDeparture(dep.expected) }}
                    </span>
                    <span v-if="dep.platform" class="transport-platform">{{ dep.platform }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.transport-widget {
    padding: 0.5rem;
    min-height: 60px;
    height: 100%;
    box-sizing: border-box;
    overflow: hidden;
}

.transport-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.transport-empty .ti {
    font-size: 1.5rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.transport-spinner {
    animation: spin 1s linear infinite;
}

.transport-content {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.transport-header {
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 0.9em;
    margin-bottom: 0.5rem;
    flex-shrink: 0;
    gap: 0.5rem;
}

.transport-refresh {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--p-text-muted-color);
    padding: 0.15rem;
    border-radius: 4px;
    display: flex;
    align-items: center;
    font-size: 1em;
    transition: color 0.15s;
}

.transport-refresh:hover {
    color: var(--p-text-color);
}

.transport-board {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
}

.transport-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.25rem 0;
    border-bottom: 1px solid color-mix(in srgb, currentColor 10%, transparent);
    font-size: 0.875rem;
}

.transport-row:last-child {
    border-bottom: none;
}

.transport-line {
    font-weight: 700;
    min-width: 3rem;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-shrink: 0;
}

.transport-line .ti {
    font-size: 0.9em;
    color: var(--p-text-muted-color);
}

.transport-dest {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.transport-mins {
    font-weight: 700;
    min-width: 2.5rem;
    text-align: right;
    flex-shrink: 0;
}

.transport-delayed {
    color: var(--p-red-500, #ef4444);
}

.transport-platform {
    font-size: 0.75em;
    color: var(--p-text-muted-color);
    min-width: 1.5rem;
    text-align: center;
    flex-shrink: 0;
}
</style>
