<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'
import type { ClockWidgetConfig } from '@/types/clock'
import { useThemes } from '@/composables/useThemes'

const props = defineProps<{
    config: ClockWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: ClockWidgetConfig]
}>()

const hour12 = ref(props.config?.hour12 ?? false)
const showSeconds = ref(props.config?.showSeconds ?? true)
const showDate = ref(props.config?.showDate ?? true)
const font = ref(props.config?.font ?? '')
const { data: themesData } = useThemes()

const fontOptions = computed(() => {
    if (!themesData.value) return [{ label: 'Default', value: '' }]
    const opts: { label: string; value: string }[] = [{ label: 'Default', value: '' }]
    for (const theme of themesData.value) {
        for (const f of theme.fonts ?? []) {
            opts.push({ label: f.name, value: f.name })
        }
    }
    return opts
})

watch(() => props.config, (val) => {
    if (val) {
        hour12.value = val.hour12
        showSeconds.value = val.showSeconds
        showDate.value = val.showDate
        font.value = val.font ?? ''
    }
})

const emitUpdate = () => {
    emit('update:config', {
        hour12: hour12.value,
        showSeconds: showSeconds.value,
        showDate: showDate.value,
        font: font.value || undefined,
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
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Font</label>
            <Select
                v-model="font"
                :options="fontOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                @update:modelValue="emitUpdate"
            />
        </div>
    </div>
</template>
