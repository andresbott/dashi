<script setup lang="ts">
import { ref, computed } from 'vue'
import draggable from 'vuedraggable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DashboardWidget from '@/components/dashboards/DashboardWidget.vue'
import type { Row, Widget } from '@/types/dashboard'
import { v4 as uuidv4 } from 'uuid'

const props = defineProps<{
    row: Row
    isFirst: boolean
    isLast: boolean
}>()

const emit = defineEmits<{
    update: [row: Row]
    delete: []
    'move-up': []
    'move-down': []
}>()

const widgets = computed({
    get: () => props.row.widgets,
    set: (val: Widget[]) => emit('update', { ...props.row, widgets: val })
})

const settingsVisible = ref(false)
const editTitle = ref('')
const editHeight = ref('')
const editWidth = ref('')

const openSettings = () => {
    editTitle.value = props.row.title ?? ''
    editHeight.value = props.row.height
    editWidth.value = props.row.width
    settingsVisible.value = true
}

const saveSettings = () => {
    emit('update', { ...props.row, title: editTitle.value || undefined, height: editHeight.value, width: editWidth.value })
    settingsVisible.value = false
}

const updateWidget = (index: number, widget: Widget) => {
    const updated = [...props.row.widgets]
    updated[index] = widget
    emit('update', { ...props.row, widgets: updated })
}

const deleteWidget = (index: number) => {
    const updated = props.row.widgets.filter((_, i) => i !== index)
    emit('update', { ...props.row, widgets: updated })
}

const addWidgetDialogVisible = ref(false)

const widgetTypeOptions = [
    { value: 'placeholder', label: 'Placeholder', icon: 'ti-layout-grid', description: 'Empty placeholder widget' },
    { value: 'weather', label: 'Weather', icon: 'ti-sun', description: 'Current conditions and forecast' },
    { value: 'weather-compact', label: 'Weather (Compact)', icon: 'ti-cloud', description: 'Compact weather display' },
    { value: 'bookmark', label: 'Bookmark', icon: 'ti-bookmark', description: 'Link to an external website' },
    { value: 'clock', label: 'Clock', icon: 'ti-clock', description: 'Digital clock with date' },
    { value: 'battery', label: 'Battery', icon: 'ti-battery-2', description: 'Battery status from query parameter' },
]

const addWidget = (type: string) => {
    const option = widgetTypeOptions.find(o => o.value === type)
    const newWidget: Widget = {
        id: uuidv4(),
        type,
        title: option?.label ?? 'New Widget',
        width: 6
    }
    emit('update', { ...props.row, widgets: [...props.row.widgets, newWidget] })
    addWidgetDialogVisible.value = false
}

const gridRef = ref<InstanceType<typeof draggable> | null>(null)
const resizingIndex = ref<number | null>(null)
const resizePreviewWidth = ref<number | null>(null)

const startResize = (index: number, event: MouseEvent) => {
    event.preventDefault()
    resizingIndex.value = index
    resizePreviewWidth.value = null

    const gridEl = gridRef.value?.$el as HTMLElement | undefined

    const onMouseMove = (e: MouseEvent) => {
        if (resizingIndex.value === null || !gridEl) return
        const gridRect = gridEl.getBoundingClientRect()
        const colWidth = gridRect.width / 12
        const widgetEl = gridEl.children[resizingIndex.value] as HTMLElement
        const widgetRect = widgetEl.getBoundingClientRect()
        const newCols = Math.round((e.clientX - widgetRect.left) / colWidth)
        const clamped = Math.max(1, Math.min(12, newCols))
        resizePreviewWidth.value = clamped
    }

    const onMouseUp = () => {
        if (resizingIndex.value !== null && resizePreviewWidth.value !== null) {
            updateWidget(resizingIndex.value, {
                ...props.row.widgets[resizingIndex.value],
                width: resizePreviewWidth.value
            })
        }
        resizingIndex.value = null
        resizePreviewWidth.value = null
        document.removeEventListener('mousemove', onMouseMove)
        document.removeEventListener('mouseup', onMouseUp)
    }

    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
}

const getWidgetClass = (element: Widget, index: number) => {
    if (resizingIndex.value === index && resizePreviewWidth.value !== null) {
        return 'col-' + resizePreviewWidth.value
    }
    return 'col-' + element.width
}
</script>

<template>
    <div class="dashboard-row-editor">
        <div class="row-controls flex align-items-center gap-2 mb-2">
            <span v-if="row.title" class="row-title flex-grow-1">{{ row.title }}</span>
            <span v-else class="flex-grow-1"></span>
            <Button
                icon="ti ti-arrow-up"
                text
                rounded
                class="p-1"
                :disabled="isFirst"
                @click="emit('move-up')"
            />
            <Button
                icon="ti ti-arrow-down"
                text
                rounded
                class="p-1"
                :disabled="isLast"
                @click="emit('move-down')"
            />
            <Button
                icon="ti ti-plus"
                text
                rounded
                class="p-1"
                v-tooltip.top="'Add widget'"
                @click="addWidgetDialogVisible = true"
            />
            <Button
                icon="ti ti-settings"
                text
                rounded
                class="p-1"
                @click="openSettings"
                v-tooltip.top="'Row settings'"
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
        <draggable
            v-model="widgets"
            group="widgets"
            item-key="id"
            class="grid"
            handle=".widget-drag-handle"
            ref="gridRef"
        >
            <template #item="{ element, index }">
                <div :class="getWidgetClass(element, index)" class="widget-col">
                    <div class="widget-drag-handle" style="cursor: grab; text-align: center">
                        <i class="ti ti-grip-horizontal" style="color: var(--p-text-muted-color)" />
                    </div>
                    <DashboardWidget
                        :widget="element"
                        @update="updateWidget(index, $event)"
                        @delete="deleteWidget(index)"
                    />
                    <div
                        class="resize-handle"
                        @mousedown="startResize(index, $event)"
                    />
                </div>
            </template>
        </draggable>
        <div v-if="!row.widgets.length" class="empty-row" @click="addWidgetDialogVisible = true">
            Click to add a widget
        </div>
    </div>

    <Dialog
        v-model:visible="settingsVisible"
        header="Row Settings"
        modal
        :closable="true"
        :draggable="false"
        style="width: 24rem"
    >
        <div class="flex flex-column gap-3">
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Title</label>
                <InputText v-model="editTitle" placeholder="Optional row title" />
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Height</label>
                <InputText v-model="editHeight" placeholder="e.g. auto, 300px, 50%" />
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Width</label>
                <InputText v-model="editWidth" placeholder="e.g. 100%, 1200px, 80%" />
            </div>
        </div>
        <div class="flex justify-content-end gap-3 mt-4">
            <Button label="Save" icon="ti ti-check" @click="saveSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="settingsVisible = false" />
        </div>
    </Dialog>

    <Dialog
        v-model:visible="addWidgetDialogVisible"
        header="Add Widget"
        modal
        :closable="true"
        :draggable="false"
        style="width: 24rem"
    >
        <div class="flex flex-column gap-1">
            <div
                v-for="opt in widgetTypeOptions"
                :key="opt.value"
                class="widget-type-option"
                @click="addWidget(opt.value)"
            >
                <i :class="'ti ' + opt.icon" class="widget-type-icon" />
                <div>
                    <div class="font-semibold">{{ opt.label }}</div>
                    <div class="text-xs" style="color: var(--p-text-muted-color)">{{ opt.description }}</div>
                </div>
            </div>
        </div>
    </Dialog>
</template>

<style scoped>
.row-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--p-text-color);
}

.dashboard-row-editor {
    background: var(--p-surface-50);
    border: 1px solid var(--p-surface-200);
    border-radius: 8px;
    padding: 0.75rem;
    margin-bottom: 0.75rem;
}

.widget-col {
    position: relative;
}

.resize-handle {
    position: absolute;
    top: 0;
    right: -4px;
    width: 8px;
    height: 100%;
    cursor: col-resize;
    z-index: 10;
}

.resize-handle:hover {
    background: var(--p-primary-color);
    opacity: 0.3;
    border-radius: 4px;
}

.empty-row {
    padding: 2rem;
    text-align: center;
    color: var(--p-text-muted-color);
    border: 1px dashed var(--p-surface-300);
    border-radius: 8px;
    cursor: pointer;
}

.empty-row:hover {
    background: var(--p-surface-50);
}

.widget-type-option {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 6px;
    cursor: pointer;
}

.widget-type-option:hover {
    background: var(--p-surface-100);
}

.widget-type-icon {
    font-size: 1.5rem;
    width: 2rem;
    text-align: center;
    flex-shrink: 0;
}
</style>
