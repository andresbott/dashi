<script setup lang="ts">
import { ref, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
import WeatherWidget from '@/components/dashboards/WeatherWidget.vue'
import WeatherWidgetConfig from '@/components/dashboards/WeatherWidgetConfig.vue'
import BookmarkWidget from '@/components/dashboards/BookmarkWidget.vue'
import ClockWidget from '@/components/dashboards/ClockWidget.vue'
import BookmarkWidgetConfig from '@/components/dashboards/BookmarkWidgetConfig.vue'
import ClockWidgetConfig from '@/components/dashboards/ClockWidgetConfig.vue'
import type { Widget } from '@/types/dashboard'
import type { WeatherWidgetConfig as WeatherWidgetConfigType } from '@/types/weather'
import type { BookmarkWidgetConfig as BookmarkWidgetConfigType } from '@/types/bookmark'
import type { ClockWidgetConfig as ClockWidgetConfigType } from '@/types/clock'

const props = defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    update: [widget: Widget]
    delete: []
}>()

const settingsVisible = ref(false)
const editTitle = ref('')

const weatherConfig = computed<WeatherWidgetConfigType | null>(() => {
    if (props.widget.type !== 'weather' || !props.widget.config) return null
    return props.widget.config as unknown as WeatherWidgetConfigType
})

const editWeatherConfig = ref<WeatherWidgetConfigType | null>(null)

const isWeatherType = computed(() => props.widget.type === 'weather')
const isBookmarkType = computed(() => props.widget.type === 'bookmark')
const isClockType = computed(() => props.widget.type === 'clock')

const bookmarkConfig = computed<BookmarkWidgetConfigType | null>(() => {
    if (props.widget.type !== 'bookmark' || !props.widget.config) return null
    return props.widget.config as unknown as BookmarkWidgetConfigType
})

const editBookmarkConfig = ref<BookmarkWidgetConfigType | null>(null)

const clockConfig = computed<ClockWidgetConfigType | null>(() => {
    if (props.widget.type !== 'clock' || !props.widget.config) return null
    return props.widget.config as unknown as ClockWidgetConfigType
})

const editClockConfig = ref<ClockWidgetConfigType | null>(null)

const openSettings = () => {
    editTitle.value = props.widget.title
    editWeatherConfig.value = weatherConfig.value ? { ...weatherConfig.value } : null
    editBookmarkConfig.value = bookmarkConfig.value ? { ...bookmarkConfig.value } : null
    editClockConfig.value = clockConfig.value ? { ...clockConfig.value } : { hour12: false, showSeconds: true, showDate: true }
    settingsVisible.value = true
}

const saveSettings = () => {
    const updated = { ...props.widget, title: editTitle.value }
    if (isWeatherType.value && editWeatherConfig.value) {
        updated.config = editWeatherConfig.value as unknown as Record<string, unknown>
    } else if (isBookmarkType.value && editBookmarkConfig.value) {
        updated.config = editBookmarkConfig.value as unknown as Record<string, unknown>
    } else if (isClockType.value && editClockConfig.value) {
        updated.config = editClockConfig.value as unknown as Record<string, unknown>
    }
    emit('update', updated)
    settingsVisible.value = false
}

const onUpdateWeatherConfig = (config: WeatherWidgetConfigType) => {
    editWeatherConfig.value = config
}

const onUpdateBookmarkConfig = (config: BookmarkWidgetConfigType) => {
    editBookmarkConfig.value = config
}

const onUpdateClockConfig = (config: ClockWidgetConfigType) => {
    editClockConfig.value = config
}
</script>

<template>
    <div class="dashboard-widget">
        <div class="widget-controls flex align-items-center gap-1">
            <Button
                icon="ti ti-pencil"
                text
                rounded
                class="p-1"
                @click="openSettings"
                v-tooltip.top="'Widget settings'"
            />
            <Button
                icon="ti ti-trash"
                text
                rounded
                severity="danger"
                class="p-1"
                @click="emit('delete')"
            />
        </div>
        <WeatherWidget v-if="widget.type === 'weather'" :widget="widget" />
        <BookmarkWidget v-else-if="widget.type === 'bookmark'" :widget="widget" />
        <ClockWidget v-else-if="widget.type === 'clock'" :widget="widget" />
        <WidgetPlaceholder v-else :title="widget.title" />
        <span class="widget-width-label">{{ widget.width }}/12</span>
    </div>

    <Dialog
        v-model:visible="settingsVisible"
        header="Widget Settings"
        modal
        :closable="true"
        :draggable="false"
        style="width: 28rem"
    >
        <div class="flex flex-column gap-3">
            <div v-if="!isWeatherType && !isBookmarkType && !isClockType" class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Title</label>
                <InputText v-model="editTitle" placeholder="Widget title" @keydown.enter="saveSettings" />
            </div>
            <WeatherWidgetConfig
                v-if="isWeatherType"
                :config="editWeatherConfig"
                @update:config="onUpdateWeatherConfig"
            />
            <BookmarkWidgetConfig
                v-if="isBookmarkType"
                :config="editBookmarkConfig"
                @update:config="onUpdateBookmarkConfig"
            />
            <ClockWidgetConfig
                v-if="isClockType"
                :config="editClockConfig"
                @update:config="onUpdateClockConfig"
            />
        </div>
        <div class="flex justify-content-end gap-3 mt-4">
            <Button label="Save" icon="ti ti-check" @click="saveSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="settingsVisible = false" />
        </div>
    </Dialog>
</template>

<style scoped>
.dashboard-widget {
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
    padding: 0.5rem;
}

.widget-controls {
    position: absolute;
    top: 0.25rem;
    right: 0.25rem;
    z-index: 5;
}

.dashboard-widget {
    position: relative;
}

.widget-width-label {
    position: absolute;
    bottom: 0.25rem;
    right: 0.5rem;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}
</style>
