<script setup lang="ts">
import { ref, watch } from 'vue'
import Select from 'primevue/select'
import type { XkcdWidgetConfig } from '@/types/xkcd'

const props = defineProps<{
    config: XkcdWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: XkcdWidgetConfig]
}>()

const modeOptions = [
    { label: 'Latest', value: 'latest' },
    { label: 'Daily Random', value: 'random' },
    { label: 'Random on Each Load', value: 'random-each' },
]

const editMode = ref(props.config?.mode ?? 'latest')

watch(() => props.config, (val) => {
    if (val) {
        editMode.value = val.mode ?? 'latest'
    }
})

const emitUpdate = () => {
    emit('update:config', {
        mode: editMode.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Mode</label>
            <Select
                v-model="editMode"
                :options="modeOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                @change="emitUpdate"
            />
        </div>
    </div>
</template>
