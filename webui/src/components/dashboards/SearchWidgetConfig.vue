<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import type { SearchWidgetConfig, SearchEngine } from '@/types/search'

const props = defineProps<{
    config: SearchWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: SearchWidgetConfig]
}>()

const engineOptions = [
    { label: 'Google', value: 'google' },
    { label: 'DuckDuckGo', value: 'duckduckgo' },
    { label: 'Bing', value: 'bing' },
]

const editEngine = ref<SearchEngine>(props.config?.engine ?? 'google')
const editPlaceholder = ref(props.config?.placeholder ?? 'Search...')

watch(() => props.config, (val) => {
    if (val) {
        editEngine.value = val.engine
        editPlaceholder.value = val.placeholder
    }
})

const emitUpdate = () => {
    emit('update:config', {
        engine: editEngine.value,
        placeholder: editPlaceholder.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Search Engine</label>
            <Select
                v-model="editEngine"
                :options="engineOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                @update:modelValue="emitUpdate"
            />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Placeholder Text</label>
            <InputText v-model="editPlaceholder" placeholder="Search..." @input="emitUpdate" />
        </div>
    </div>
</template>
