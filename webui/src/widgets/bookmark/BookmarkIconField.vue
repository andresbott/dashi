<script setup lang="ts">
// The bookmark icon picker: source (Tabler, selfh.st, dashboard asset) and
// the matching input. Shared by the bookmark and bookmarkgroup forms.
import { ref, inject, computed, useId } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import IconSelector from '@/components/dashboards/IconSelector.vue'
import { parseIcon, getSelfhstIconUrl, getDashboardIconUrl } from '@/lib/iconUtils'
import type { IconType } from '@/lib/iconUtils'
import { useDashboardAssets } from '@/composables/useDashboards'

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    'update:modelValue': [icon: string]
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))
const iconLabelId = useId()

const parsed = computed(() => parseIcon(props.modelValue))

// Dashboard assets need a dashboard to list them from: outside one (the
// admin bookmark-file editor) the option is only kept for a value that
// already uses it.
const iconTypeOptions = computed(() => [
    { label: 'Tabler', value: 'tabler' },
    { label: 'selfh.st', value: 'selfhst' },
    ...(dashboardId.value || parsed.value.type === 'dashboard' ? [{ label: 'Dashboard asset', value: 'dashboard' }] : []),
])

const iconType = computed({
    get: () => parsed.value.type,
    set: (v: IconType) => {
        if (v === 'tabler') emit('update:modelValue', 'ti-bookmark')
        else if (v === 'selfhst') emit('update:modelValue', 'selfhst:')
        else if (v === 'dashboard') emit('update:modelValue', 'dashboard:')
    },
})

const selfhstFilename = computed({
    get: () => parsed.value.type === 'selfhst' ? parsed.value.value : '',
    set: (v: string) => emit('update:modelValue', 'selfhst:' + v),
})

const dashboardFilename = computed({
    get: () => parsed.value.type === 'dashboard' ? parsed.value.value : '',
    set: (v: string) => emit('update:modelValue', 'dashboard:' + v),
})

const { data: dashboardAssetsData } = useDashboardAssets(() => dashboardId.value)
const imageExtensions = /\.(png|webp|svg|jpg|jpeg|ico|gif)$/i

const dashboardAssetOptions = computed(() =>
    (dashboardAssetsData.value ?? []).filter(a => imageExtensions.test(a)).map(a => ({ label: a, value: a }))
)

const iconPreviewSrc = computed(() => {
    if (parsed.value.type === 'selfhst' && parsed.value.value) {
        return getSelfhstIconUrl(parsed.value.value)
    }
    if (parsed.value.type === 'dashboard' && parsed.value.value && dashboardId.value) {
        return getDashboardIconUrl(dashboardId.value, parsed.value.value)
    }
    return null
})
</script>

<template>
    <div class="flex flex-column gap-1">
        <label class="text-sm font-semibold">Icon source</label>
        <Select
            :modelValue="iconType"
            @update:modelValue="(v: IconType | undefined) => { if (v) iconType = v }"
            :options="iconTypeOptions"
            optionLabel="label"
            optionValue="value"
            class="w-full"
        />
    </div>
    <div v-if="iconType === 'tabler'" class="flex flex-column gap-1">
        <label :id="iconLabelId" class="text-sm font-semibold">Icon</label>
        <IconSelector :modelValue="modelValue" :labelledBy="iconLabelId" @update:modelValue="(v: string) => emit('update:modelValue', v)" />
    </div>
    <div v-if="iconType === 'selfhst'" class="flex flex-column gap-1">
        <label class="text-sm font-semibold">Filename</label>
        <div class="text-xs text-color-secondary mb-1">
            Enter a filename from <a href="https://selfh.st/icons/" target="_blank">selfh.st/icons</a>,
            e.g. <code>2fauth-light.webp</code> or <code>sonarr.png</code>
        </div>
        <InputText v-model="selfhstFilename" placeholder="filename.webp" />
        <img
            v-if="iconPreviewSrc"
            :src="iconPreviewSrc"
            class="icon-preview"
        />
    </div>
    <div v-if="iconType === 'dashboard'" class="flex flex-column gap-1">
        <label class="text-sm font-semibold">Asset</label>
        <Select
            v-if="dashboardAssetOptions.length > 0"
            :modelValue="dashboardFilename"
            @update:modelValue="(v: string | undefined) => { if (v !== undefined) dashboardFilename = v }"
            :options="dashboardAssetOptions"
            optionLabel="label"
            optionValue="value"
            placeholder="Select an image..."
            class="w-full"
        />
        <div v-else class="text-xs text-color-secondary">
            No image files found in this dashboard's assets folder.
        </div>
        <img
            v-if="iconPreviewSrc"
            :src="iconPreviewSrc"
            class="icon-preview"
        />
    </div>
</template>

<style scoped>
.icon-preview {
    width: 2rem;
    height: 2rem;
    object-fit: contain;
    margin-top: 0.25rem;
}
</style>
