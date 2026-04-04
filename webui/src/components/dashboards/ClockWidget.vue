<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Widget } from '@/types/dashboard'
import type { ClockWidgetConfig } from '@/types/clock'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<ClockWidgetConfig>(() => {
    if (!props.widget.config) return { hour12: false, showSeconds: true, showDate: true }
    const c = props.widget.config as unknown as ClockWidgetConfig
    return {
        hour12: c.hour12 ?? false,
        showSeconds: c.showSeconds ?? true,
        showDate: c.showDate ?? true,
        font: c.font,
    }
})

const clockFont = computed(() => config.value.font || undefined)

const time = ref('')
const date = ref('')
let timer: ReturnType<typeof setInterval> | undefined

const update = () => {
    const now = new Date()
    const timeOpts: Intl.DateTimeFormatOptions = {
        hour: '2-digit',
        minute: '2-digit',
        hour12: config.value.hour12,
    }
    if (config.value.showSeconds) {
        timeOpts.second = '2-digit'
    }
    time.value = now.toLocaleTimeString(undefined, timeOpts)
    date.value = now.toLocaleDateString(undefined, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

onMounted(() => {
    update()
    timer = setInterval(update, 1000)
})

onUnmounted(() => {
    clearInterval(timer)
})
</script>

<template>
    <div class="clock-widget">
        <div class="clock-time" :style="{ fontFamily: clockFont }">{{ time }}</div>
        <div v-if="config.showDate" class="clock-date">{{ date }}</div>
    </div>
</template>

<style scoped>
.clock-widget {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0.75rem 0.5rem;
    min-height: 60px;
}

.clock-time {
    font-size: 2rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
}

.clock-date {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
}
</style>
