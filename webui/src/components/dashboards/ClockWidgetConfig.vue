<script setup lang="ts">
import { ref, watch } from 'vue'
import Checkbox from 'primevue/checkbox'
import type { ClockWidgetConfig } from '@/types/clock'

const props = defineProps<{
    config: ClockWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: ClockWidgetConfig]
}>()

const hour12 = ref(props.config?.hour12 ?? false)
const showSeconds = ref(props.config?.showSeconds ?? true)
const showDate = ref(props.config?.showDate ?? true)

watch(() => props.config, (val) => {
    if (val) {
        hour12.value = val.hour12
        showSeconds.value = val.showSeconds
        showDate.value = val.showDate
    }
})

const emitUpdate = () => {
    emit('update:config', {
        hour12: hour12.value,
        showSeconds: showSeconds.value,
        showDate: showDate.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="hour12" :binary="true" inputId="hour12" @update:modelValue="emitUpdate" />
            <label for="hour12" class="text-sm font-semibold">12-hour format</label>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="showSeconds" :binary="true" inputId="showSeconds" @update:modelValue="emitUpdate" />
            <label for="showSeconds" class="text-sm font-semibold">Show seconds</label>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="showDate" :binary="true" inputId="showDate" @update:modelValue="emitUpdate" />
            <label for="showDate" class="text-sm font-semibold">Show date</label>
        </div>
    </div>
</template>
