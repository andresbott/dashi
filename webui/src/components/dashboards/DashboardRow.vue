<script lang="ts">
import { reactive } from 'vue'

// Module-scoped, shared across all DashboardRow instances for the duration of a
// drag. Lets any row the pointer moves over show the 12-column guides and
// receive the drop at the pointer's column — so one drag sets row + column.
const dragState = reactive({
    active: false,
    span: 1,
    leftX: null as number | null,
    pointerY: 0,
    // The row + column the pointer currently hovers, published by whichever row
    // shows the drop guide and read by the source row on release to commit a
    // cross-row move (SortableJS can't place a drop into a wide col-offset gap).
    targetRowId: null as string | null,
    targetColumn: null as number | null,
})
</script>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import draggable from 'vuedraggable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DashboardWidget from '@/components/dashboards/DashboardWidget.vue'
import type { Row, Widget } from '@/types/dashboard'
import { getWidgetTypeOptions } from '@/lib/widgetRegistry'
import { placeRow, firstFreeSpan, insertByColumn, columnFromDrag, moveWidgetToColumn, GRID_COLUMNS } from '@/lib/rowLayout'
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
    'move-widget': [payload: { widgetId: string; toRowId: string; column: number }]
}>()

// Vue owns the widget list. SortableJS only detects the drag — its list
// mutation is disabled (:move returns false) and the binding is one-way; every
// reorder and cross-row move is committed manually on drop.
const widgets = computed(() => props.row.widgets)

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
let originLeftX = 0

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

// Guides for a row that is a potential drop target while another row is being
// dragged from: where the widget would land if dropped here now (pointer over
// this row). Null unless a drag is active and this row is not the source.
const targetGuide = computed(() => {
    if (!dragState.active || dragging.value || dragState.leftX === null) return null
    const gridEl = gridRef.value?.$el as HTMLElement | undefined
    if (!gridEl) return null
    const rect = gridEl.getBoundingClientRect()
    if (dragState.pointerY < rect.top || dragState.pointerY > rect.bottom) return null
    const colWidth = rect.width / GRID_COLUMNS
    const raw = Math.round((dragState.leftX - rect.left) / colWidth) + 1
    return Math.max(1, Math.min(GRID_COLUMNS - dragState.span + 1, raw))
})

// While this row is a drop target (pointer hovering it, and it is not the drag
// source), publish itself + the guide column to the shared drag state so the
// source row can commit the cross-row move on release. Sync flush keeps the
// value current the instant the drop fires.
watch(targetGuide, (col) => {
    if (col !== null) {
        dragState.targetRowId = props.row.id
        dragState.targetColumn = col
    } else if (dragState.targetRowId === props.row.id) {
        dragState.targetRowId = null
        dragState.targetColumn = null
    }
}, { flush: 'sync' })

// The overlay renders this row's own in-row guide while it is the drag source,
// otherwise the cross-row target guide — so the guides follow the pointer from
// row to row during a single drag.
const activeGuideColumn = computed(() => (dragging.value ? guideColumn.value : targetGuide.value))
const activeGuideWidth = computed(() => (dragging.value ? guideWidth.value : dragState.span))

const colWidthPx = (gridEl: HTMLElement) => gridEl.getBoundingClientRect().width / GRID_COLUMNS

// Right edge → column span. Growing past a right-hand neighbour pushes it (and
// the rest) along via placeRow rather than blocking; the resized widget's own
// span is capped at the grid's right edge.
const startResize = (index: number, event: MouseEvent) => {
    event.preventDefault()
    resizingIndex.value = index
    resizePreviewWidth.value = null

    const gridEl = gridRef.value?.$el as HTMLElement | undefined
    const base = placeRow(props.row.widgets)
    const col = base[index].column
    // Cap the span at the grid's right edge for the widget itself; growing past
    // a right-hand neighbour is allowed — placeRow pushes the neighbours along
    // (they wrap onto the next line if the row overflows), mirroring a move.
    const maxWidth = GRID_COLUMNS - col + 1

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
// row moves the widget there (placed by column in the widgets setter). Dropping
// within the same row reorders the row via moveWidgetToColumn — the widget can
// pass its neighbours and land in an earlier or later slot at the drop column.
// The guides track the pointer during the drag; the widget lands on release.
const onDragStart = (evt: { oldIndex: number }) => {
    const index = evt.oldIndex
    const base = placeRow(props.row.widgets)
    const p = base[index]
    guideWidth.value = p.width
    // The whole row is reachable: a within-row drop reorders the widgets
    // (moveWidgetToColumn) rather than nudging within the neighbour gap, so the
    // guide is only clamped to keep the dragged span on the 12-column grid.
    dragMin = 1
    dragMax = Math.max(1, GRID_COLUMNS - p.width + 1)
    guideColumn.value = p.column
    originColumn = p.column
    dragEl = (gridRef.value?.$el as HTMLElement | undefined) ?? null
    // Widget's left edge in viewport px; shared so the target row of a
    // cross-row drop can place it at the drop column.
    const startRect = dragEl?.getBoundingClientRect()
    originLeftX = startRect ? startRect.left + (p.column - 1) * (startRect.width / GRID_COLUMNS) : 0
    dragState.active = true
    dragState.span = p.width
    dragState.leftX = originLeftX
    dragState.targetRowId = null
    dragState.targetColumn = null
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
    // Share the widget's live left edge + pointer Y so any row can show guides
    // and place it on drop.
    dragState.leftX = originLeftX + (e.clientX - grabStartX)
    dragState.pointerY = e.clientY
    // This row's own (source) guides only while the pointer is over it; once it
    // moves to another row, that row shows its own target guides instead.
    const overThisRow = e.clientY >= rect.top && e.clientY <= rect.bottom
    guideColumn.value = overThisRow
        ? columnFromDrag({
              originColumn,
              pointerX: e.clientX,
              grabStartX,
              colWidth: rect.width / GRID_COLUMNS,
              min: dragMin,
              max: dragMax,
          })
        : null
}

const onDragEnd = (evt: { oldIndex: number }) => {
    document.removeEventListener('pointermove', onDragMove)
    document.removeEventListener('mousemove', onDragMove)
    dragging.value = false
    const index = evt.oldIndex
    const { targetRowId, targetColumn } = dragState
    if (index >= 0 && index < props.row.widgets.length) {
        if (targetRowId !== null && targetRowId !== props.row.id && targetColumn !== null) {
            // Dropped over another row: hand the widget to the parent, which
            // removes it here and inserts it there at the guide column.
            emit('move-widget', { widgetId: props.row.widgets[index].id, toRowId: targetRowId, column: targetColumn })
        } else if (guideColumn.value !== null) {
            // Dropped within this row: reposition/reorder at the guide column.
            emit('update', { ...props.row, widgets: moveWidgetToColumn(props.row.widgets, index, guideColumn.value) })
        }
    }
    guideColumn.value = null
    grabStartX = null
    dragEl = null
    dragState.active = false
    dragState.leftX = null
    dragState.targetRowId = null
    dragState.targetColumn = null
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
        <div
            class="grid-wrap"
            :class="{ 'grid-wrap--empty': !row.widgets.length }"
            @click="!row.widgets.length && (addWidgetDialogVisible = true)"
        >
            <div v-if="activeGuideColumn !== null" class="column-guides" aria-hidden="true">
                <div
                    v-for="c in 12"
                    :key="c"
                    class="guide-col"
                    :class="{ 'guide-col--active': c >= activeGuideColumn && c < activeGuideColumn + activeGuideWidth }"
                />
            </div>
            <draggable
                :model-value="widgets"
                group="widgets"
                item-key="id"
                class="grid"
                handle=".widget-drag-handle"
                :sort="false"
                :move="() => false"
                :force-fallback="true"
                :style="!row.widgets.length ? { minHeight: '72px' } : undefined"
                ref="gridRef"
                @start="onDragStart"
                @end="onDragEnd"
            >
                <template #item="{ element, index }">
                    <div :class="getWidgetClass(index)" class="widget-col">
                        <DashboardWidget
                            :widget="element"
                            @update="updateWidget(index, $event)"
                            @delete="deleteWidget(index)"
                        >
                            <template #handle>
                                <span
                                    class="widget-drag-handle"
                                    v-tooltip.top="'Drag to move / position'"
                                    @pointerdown="onGripDown"
                                >
                                    <i class="ti ti-grip-vertical" />
                                </span>
                            </template>
                            <template #resize>
                                <span
                                    class="resize-handle"
                                    v-tooltip.top="'Drag to resize width'"
                                    @mousedown="startResize(index, $event)"
                                >
                                    <i class="ti ti-arrows-horizontal" />
                                </span>
                            </template>
                        </DashboardWidget>
                    </div>
                </template>
            </draggable>
            <div v-if="!row.widgets.length" class="empty-hint">Click to add a widget</div>
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

/* Drag/move grip, slotted into the widget card header. Slot content is compiled
   in this (parent) scope, so these styles apply even though it renders inside
   DashboardWidget. */
.widget-drag-handle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 6px;
    cursor: grab;
    color: var(--p-text-muted-color);
}

.widget-drag-handle:hover {
    background: var(--p-surface-100);
    color: var(--p-text-color);
}

.widget-drag-handle:active {
    cursor: grabbing;
}

/* Width-resize indicator, slotted into the card's bottom-right corner
   (positioned against .dashboard-widget, which is position: relative). */
.resize-handle {
    position: absolute;
    right: 0;
    bottom: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    border-bottom-right-radius: 8px;
    cursor: col-resize;
    color: var(--p-text-muted-color);
    z-index: 10;
}

.resize-handle:hover {
    color: var(--p-primary-color);
    background: color-mix(in srgb, var(--p-primary-color) 12%, transparent);
}

.resize-handle i {
    font-size: 0.875rem;
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

.grid-wrap--empty {
    cursor: pointer;
}

/* Placeholder for an empty row. pointer-events: none so it never intercepts a
   cross-row drop onto the (now min-height) empty draggable behind it. */
.empty-hint {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--p-text-muted-color);
    border: 1px dashed var(--p-surface-300);
    border-radius: 8px;
    pointer-events: none;
}

.grid-wrap--empty:hover .empty-hint {
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
