<script setup lang="ts">
import { ref, watch, inject, computed } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import { useDashboardAssets } from '@/composables/useDashboards'
import Select from 'primevue/select'
import type { ImageWidgetConfig } from '@/types/image'

const props = defineProps<{
    config: ImageWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: ImageWidgetConfig]
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))
const { data: assets, isLoading } = useDashboardAssets(() => dashboardId.value)

const imageFiles = computed(() => {
    if (!assets.value) return []
    const imageExts = ['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.bmp']
    return assets.value
        .filter(a => imageExts.some(ext => a.toLowerCase().endsWith(ext)))
        .map(a => ({ label: a, value: a }))
})

const fitOptions = [
    { label: 'Cover (fill, crop if needed)', value: 'cover' },
    { label: 'Contain (fit entirely)', value: 'contain' },
    { label: 'Fill (stretch)', value: 'fill' },
]

const image = ref(props.config?.image ?? '')
const fit = ref(props.config?.fit ?? 'cover')

watch(() => props.config, (val) => {
    if (val) {
        image.value = val.image ?? ''
        fit.value = val.fit ?? 'cover'
    }
})

const emitUpdate = () => {
    emit('update:config', {
        image: image.value,
        fit: fit.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Image</label>
            <Select
                v-model="image"
                :options="imageFiles"
                optionLabel="label"
                optionValue="value"
                placeholder="Select an image"
                :loading="isLoading"
                class="w-full"
                @update:modelValue="emitUpdate"
            />
            <small class="text-color-secondary">
                Upload images in the dashboard assets panel
            </small>
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Fit</label>
            <Select
                v-model="fit"
                :options="fitOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                @update:modelValue="emitUpdate"
            />
        </div>
    </div>
</template>
