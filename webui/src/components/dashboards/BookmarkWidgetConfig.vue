<script setup lang="ts">
import { ref, watch, inject, computed, onMounted } from 'vue'
import type { Ref } from 'vue'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import IconSelector from '@/components/dashboards/IconSelector.vue'
import type { BookmarkWidgetConfig } from '@/types/bookmark'
import { parseIcon, getSelfhstIconUrl, getDashboardIconUrl } from '@/lib/iconUtils'
import { getDashboardAssets } from '@/lib/api/dashboard'

const props = defineProps<{
    config: BookmarkWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: BookmarkWidgetConfig]
}>()

const dashboardId = inject<Ref<string>>('dashboardId', ref(''))

const editUrl = ref(props.config?.url ?? '')
const editIcon = ref(props.config?.icon ?? 'ti-bookmark')
const editTitle = ref(props.config?.title ?? '')
const editSubtitle = ref(props.config?.subtitle ?? '')

const iconTypeOptions = [
    { label: 'Tabler', value: 'tabler' },
    { label: 'selfh.st', value: 'selfhst' },
    { label: 'Dashboard asset', value: 'dashboard' },
]

const parsed = computed(() => parseIcon(editIcon.value))

const iconType = computed({
    get: () => parsed.value.type,
    set: (v: string) => {
        if (v === 'tabler') editIcon.value = 'ti-bookmark'
        else if (v === 'selfhst') editIcon.value = 'selfhst:'
        else if (v === 'dashboard') {
            editIcon.value = 'dashboard:'
            fetchDashboardAssets()
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

const dashboardAssets = ref<string[]>([])
const imageExtensions = /\.(png|webp|svg|jpg|jpeg|ico|gif)$/i

const dashboardAssetOptions = computed(() =>
    dashboardAssets.value.filter(a => imageExtensions.test(a)).map(a => ({ label: a, value: a }))
)

const fetchDashboardAssets = async () => {
    if (!dashboardId.value) return
    try {
        dashboardAssets.value = await getDashboardAssets(dashboardId.value)
    } catch {
        dashboardAssets.value = []
    }
}

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
    }
})

const emitUpdate = () => {
    emit('update:config', {
        url: editUrl.value,
        icon: editIcon.value,
        title: editTitle.value,
        subtitle: editSubtitle.value,
    })
}

const onTablerSelect = (icon: string) => {
    editIcon.value = icon
    emitUpdate()
}

onMounted(() => {
    if (parsed.value.type === 'dashboard') {
        fetchDashboardAssets()
    }
})
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Icon</label>
            <Select
                :modelValue="iconType"
                @update:modelValue="(v: string) => { iconType = v }"
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
                @update:modelValue="(v: string) => { dashboardFilename = v }"
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
            <InputText v-model="editUrl" placeholder="https://example.com" @input="emitUpdate" />
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
