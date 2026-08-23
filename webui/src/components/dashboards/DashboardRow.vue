<script setup lang="ts">
import { ref, computed } from 'vue'
import draggable from 'vuedraggable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DashboardWidget from '@/components/dashboards/DashboardWidget.vue'
import type { Row, Widget } from '@/types/dashboard'
import { getWidgetTypeOptions } from '@/lib/widgetRegistry'
import { placeRow, firstFreeSpan, insertByColumn, columnFromDrag, GRID_COLUMNS } from '@/lib/rowLayout'
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
    // vuedraggable only mutates this on cross-row moves (within-row sorting is
    // off). A widget dragged in from another row carries a stale column, so we
    // reset it to flow (0) at its drop position; existing widgets are untouched
    // and placeRow resolves any overlap defensively.
    set: (val: Widget[]) => {
        const known = new Set(props.row.widgets.map(w => w.id))
        const next = val.map(w => (known.has(w.id) ? w : { ...w, column: 0 }))
        emit('update', { ...props.row, widgets: next })
    }
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

const widgetTypeOptions = getWidgetTypeOptions()

const addWidget = (type: string) => {
    const option = widgetTypeOptions.find(o => o.value === type)
    // Drop the widget into the first free span so it never overlaps an
    // existing one, then splice it in at the matching column position.
    const span = firstFreeSpan(props.row.widgets, 6)
    const newWidget: Widget = {
        id: uuidv4(),
        type,
        title: option?.label ?? 'New Widget',
        width: span.width,
        column: span.column
    }
    emit('update', { ...props.row, widgets: insertByColumn(props.row.widgets, newWidget) })
    addWidgetDialogVisible.value = false
}

const gridRef = ref<InstanceType<typeof draggable> | null>(null)
const resizingIndex = ref<number | null>(null)
const resizePreviewWidth = ref<number | null>(null)

// Grip-drag ("move") state. Positioning lands on drop, so the widget doesn't
// move live; instead we show a 12-column guide overlay and highlight the target
// span (guideColumn .. guideColumn + guideWidth - 1) while the pointer moves.
const dragging = ref(false)
const guideColumn = ref<number | null>(null)
const guideWidth = ref(0)
let dragMin = 1
let dragMax = GRID_COLUMNS
let dragEl: HTMLElement | null = null
// The widget's column when the drag began and the pointer x where it was
// grabbed. Positioning is resolved from how far the pointer has moved since,
// so releasing without moving leaves the widget exactly where it was.
let originColumn = 1
let grabStartX: number | null = null

// Capture the true grab point on pointer-down, before the native drag threshold
// nudges it; fires before SortableJS starts the drag.
const onGripDown = (e: PointerEvent) => {
    grabStartX = e.clientX
}

// Resolved 12-column placement (leading gap + span) for every widget, with the
// in-progress resize preview folded in so the editor grid updates live.
const placements = computed(() => {
    const list = props.row.widgets.map((widget, i) => {
        if (i === resizingIndex.value && resizePreviewWidth.value !== null) {
            return { width: resizePreviewWidth.value, column: widget.column }
        }
        return { width: widget.width, column: widget.column }
    })
    return placeRow(list)
})

const colWidthPx = (gridEl: HTMLElement) => gridEl.getBoundingClientRect().width / GRID_COLUMNS

// Right edge → column span, clamped so it never overlaps the next widget or
// spills past the grid.
const startResize = (index: number, event: MouseEvent) => {
    event.preventDefault()
    resizingIndex.value = index
    resizePreviewWidth.value = null

    const gridEl = gridRef.value?.$el as HTMLElement | undefined
    const base = placeRow(props.row.widgets)
    const col = base[index].column
    const nextStart = index < base.length - 1 ? base[index + 1].column : GRID_COLUMNS + 1
    const maxWidth = Math.min(GRID_COLUMNS - col + 1, nextStart - col)

    const onMouseMove = (e: MouseEvent) => {
        if (resizingIndex.value === null || !gridEl) return
        const widgetEl = gridEl.children[resizingIndex.value] as HTMLElement
        const widgetRect = widgetEl.getBoundingClientRect()
        const newCols = Math.round((e.clientX - widgetRect.left) / colWidthPx(gridEl))
        resizePreviewWidth.value = Math.max(1, Math.min(maxWidth, newCols))
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

// Grip drag ("move" handle). vuedraggable owns the drag: dropping onto another
// row moves the widget there (reset to flow in the widgets setter). Dropping
// within the same row sets the widget's start column from where it was
// released, clamped between its neighbours so it never overlaps. The guides
// track the pointer during the drag; the widget lands on release.
const onDragStart = (evt: { oldIndex: number }) => {
    const index = evt.oldIndex
    const base = placeRow(props.row.widgets)
    const p = base[index]
    guideWidth.value = p.width
    dragMin = index > 0 ? base[index - 1].column + base[index - 1].width : 1
    const nextStart = index < base.length - 1 ? base[index + 1].column : GRID_COLUMNS + 1
    dragMax = Math.max(dragMin, Math.min(GRID_COLUMNS - p.width + 1, nextStart - p.width))
    guideColumn.value = p.column
    originColumn = p.column
    dragEl = (gridRef.value?.$el as HTMLElement | undefined) ?? null
    dragging.value = true
    // The draggable runs in force-fallback (pointer) mode — native HTML5
    // drag-and-drop reported unreliable dragover coordinates — so we track the
    // pointer via pointermove (mousemove kept as a belt-and-suspenders).
    document.addEventListener('pointermove', onDragMove)
    document.addEventListener('mousemove', onDragMove)
}

const onDragMove = (e: MouseEvent) => {
    if (!dragEl || e.clientX === 0) return
    // Fallback if pointer-down was never seen (e.g. a non-pointer input path):
    // anchor on the first move so there is still no jump.
    if (grabStartX === null) grabStartX = e.clientX
    const rect = dragEl.getBoundingClientRect()
    guideColumn.value = columnFromDrag({
        originColumn,
        pointerX: e.clientX,
        grabStartX,
        colWidth: rect.width / GRID_COLUMNS,
        min: dragMin,
        max: dragMax,
    })
}

const onDragEnd = (evt: { oldIndex: number; from: HTMLElement; to: HTMLElement }) => {
    document.removeEventListener('pointermove', onDragMove)
    document.removeEventListener('mousemove', onDragMove)
    dragging.value = false
    const index = evt.oldIndex
    const sameRow = evt.from === evt.to
    if (sameRow && guideColumn.value !== null && index >= 0 && index < props.row.widgets.length) {
        updateWidget(index, { ...props.row.widgets[index], column: guideColumn.value })
    }
    guideColumn.value = null
    grabStartX = null
    dragEl = null
}

const getWidgetClass = (index: number) => {
    const p = placements.value[index]
    return `col-${p.width} col-offset-${p.gap}`
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
        <div class="grid-wrap">
            <div v-if="dragging" class="column-guides" aria-hidden="true">
                <div
                    v-for="c in 12"
                    :key="c"
                    class="guide-col"
                    :class="{ 'guide-col--active': guideColumn !== null && c >= guideColumn && c < guideColumn + guideWidth }"
                />
            </div>
            <draggable
                v-model="widgets"
                group="widgets"
                item-key="id"
                class="grid"
                handle=".widget-drag-handle"
                :sort="false"
                :force-fallback="true"
                ref="gridRef"
                @start="onDragStart"
                @end="onDragEnd"
            >
                <template #item="{ element, index }">
                    <div :class="getWidgetClass(index)" class="widget-col">
                        <div
                            class="widget-drag-handle"
                            style="cursor: grab; text-align: center"
                            v-tooltip.top="'Drag to move / position'"
                            @pointerdown="onGripDown"
                        >
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
        </div>
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

.grid-wrap {
    position: relative;
}

/* 12-column guide overlay shown only while dragging a widget by its move
   handle. Matches the PrimeFlex .grid negative gutter so the dividers line up
   with the columns. pointer-events: none keeps it clear of the drag. */
.column-guides {
    position: absolute;
    top: 0;
    bottom: 0;
    left: -0.5rem;
    right: -0.5rem;
    display: flex;
    pointer-events: none;
    z-index: 3;
}

.guide-col {
    flex: 1 1 0;
    border-right: 1px dashed var(--p-surface-400);
}

.guide-col:last-child {
    border-right: none;
}

.guide-col--active {
    background: color-mix(in srgb, var(--p-primary-color) 16%, transparent);
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

<style>
/* SortableJS clones the dragged cell — including its col-offset-N margin — as
   the floating fallback drag preview. On that fixed-position clone the
   percentage margin is applied on top of the computed position and shoves it
   far to the right (proportional to the start column), so neutralise it. Not
   scoped: SortableJS creates the clone outside the component's scoped subtree. */
.sortable-fallback {
    margin-left: 0 !important;
}
</style>
