<script setup lang="ts">
import { ref, watch, inject, computed } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'
import IconSelector from '@/components/dashboards/IconSelector.vue'
import type { BookmarkWidgetConfig } from '@/types/bookmark'
import { parseIcon, getSelfhstIconUrl, getDashboardIconUrl } from '@/lib/iconUtils'
import type { IconType } from '@/lib/iconUtils'
import { useDashboardAssets } from '@/composables/useDashboards'

const props = defineProps<{
    config: BookmarkWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: BookmarkWidgetConfig]
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))

const editUrl = ref(props.config?.url ?? '')
const editIcon = ref(props.config?.icon ?? 'ti-bookmark')
const editTitle = ref(props.config?.title ?? '')
const editSubtitle = ref(props.config?.subtitle ?? '')
const editTextBelow = ref(props.config?.textBelow ?? false)

const iconTypeOptions = [
    { label: 'Tabler', value: 'tabler' },
    { label: 'selfh.st', value: 'selfhst' },
    { label: 'Dashboard asset', value: 'dashboard' },
]

const parsed = computed(() => parseIcon(editIcon.value))

const iconType = computed({
    get: () => parsed.value.type,
    set: (v: IconType) => {
        if (v === 'tabler') editIcon.value = 'ti-bookmark'
        else if (v === 'selfhst') editIcon.value = 'selfhst:'
        else if (v === 'dashboard') {
            editIcon.value = 'dashboard:'
        }
        emitUpdate()
    },
})

const selfhstFilename = computed({
    get: () => parsed.value.type === 'selfhst' ? parsed.value.value : '',
    set: (v: string) => {
        editIcon.value = 'selfhst:' + v
        emitUpdate()
    },
})

const dashboardFilename = computed({
    get: () => parsed.value.type === 'dashboard' ? parsed.value.value : '',
    set: (v: string) => {
        editIcon.value = 'dashboard:' + v
        emitUpdate()
    },
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

watch(() => props.config, (val) => {
    if (val) {
        editUrl.value = val.url
        editIcon.value = val.icon
        editTitle.value = val.title
        editSubtitle.value = val.subtitle
        editTextBelow.value = val.textBelow ?? false
    }
})

const isSafeUrl = (url: string): boolean => {
    if (!url) return true
    try {
        const parsed = new URL(url)
        return ['http:', 'https:'].includes(parsed.protocol)
    } catch {
        // Allow relative URLs and empty strings
        return !url.toLowerCase().trimStart().startsWith('javascript:')
    }
}

const urlError = computed(() => {
    if (editUrl.value && !isSafeUrl(editUrl.value)) {
        return 'Only http:// and https:// URLs are allowed'
    }
    return ''
})

const emitUpdate = () => {
    if (urlError.value) return
    emit('update:config', {
        url: editUrl.value,
        icon: editIcon.value,
        title: editTitle.value,
        subtitle: editSubtitle.value,
        textBelow: editTextBelow.value,
    })
}

const onTablerSelect = (icon: string) => {
    editIcon.value = icon
    emitUpdate()
}

// Assets are fetched reactively via useDashboardAssets query
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Icon</label>
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
            <label class="text-sm font-semibold">Icon</label>
            <div class="flex align-items-center gap-2">
                <IconSelector :modelValue="editIcon" @update:modelValue="onTablerSelect" />
                <span class="text-sm text-color-secondary">{{ editIcon }}</span>
            </div>
        </div>
        <div v-if="iconType === 'selfhst'" class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Filename</label>
            <div class="text-xs text-color-secondary mb-1">
                Enter a filename from <a href="https://selfh.st/icons/" target="_blank">selfh.st/icons</a>,
                e.g. <code>2fauth-light.webp</code> or <code>sonarr.png</code>
            </div>
            <InputText v-model="selfhstFilename" placeholder="filename.webp" @input="emitUpdate" />
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
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Title</label>
            <InputText v-model="editTitle" placeholder="Bookmark title" @input="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Subtitle</label>
            <InputText v-model="editSubtitle" placeholder="Short description" @input="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">URL</label>
            <InputText v-model="editUrl" placeholder="https://example.com" @input="emitUpdate" :invalid="!!urlError" />
            <small v-if="urlError" class="text-red-500">{{ urlError }}</small>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="editTextBelow" :binary="true" inputId="bookmarkTextBelow" @update:modelValue="emitUpdate" />
            <label for="bookmarkTextBelow" class="text-sm">Text below icon</label>
        </div>
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
