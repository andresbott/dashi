<script setup lang="ts">
import { ref, computed, inject, onMounted, onBeforeUnmount } from 'vue'
import draggable from 'vuedraggable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DashboardWidget from '@/components/dashboards/DashboardWidget.vue'
import AddWidgetDialog from '@/components/dashboards/AddWidgetDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { Row, Widget } from '@/types/dashboard'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import {
    GRID_COLUMNS,
    placeRow,
    firstFreeSpan,
    insertByColumn,
    moveWidgetToColumn,
    clampRowToLine,
    maxResizeWidth,
    totalSpan,
} from '@/lib/rowLayout'
import { duplicateInRow } from '@/lib/duplicateWidget'
import { beginDragGuard, endDragGuard } from '@/lib/dragGuard'
import { abortIfSource, dragState, endWidgetDrag, registerRow, startWidgetDrag } from '@/lib/widgetDrag'
import { WIDGET_MOVE } from '@/lib/injectionKeys'
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

// Commits a drop; provided by the editor page (absent elsewhere: drops do nothing).
const moveWidget = inject(WIDGET_MOVE, null)

// Vue owns the widget list. SortableJS only detects the drag and paints the
// floating clone: its list mutation is disabled (:sort false, :move false) and
// the binding is one-way; every drop is resolved by the shared drag session
// (lib/widgetDrag.ts) and committed through WIDGET_MOVE on release.
const widgets = computed(() => props.row.widgets)

const rootRef = ref<HTMLElement | null>(null)
const gridRef = ref<InstanceType<typeof draggable> | null>(null)
const gridEl = () => (gridRef.value?.$el as HTMLElement | undefined) ?? null

let unregisterRow: (() => void) | null = null

onMounted(() => {
    if (rootRef.value) {
        unregisterRow = registerRow(props.row.id, { root: rootRef.value, grid: gridEl, widgets: () => props.row.widgets })
    }
})

onBeforeUnmount(() => {
    unregisterRow?.()
    // A row removed mid-gesture (page switch, row delete) must not leave
    // document listeners or a stuck global drag behind.
    abortIfSource(props.row.id)
    stopResize(false)
})

// --- Row settings ---
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

// --- Delete row (confirmed when it still holds widgets) ---
const confirmDeleteVisible = ref(false)

const requestDelete = () => {
    if (props.row.widgets.length) {
        confirmDeleteVisible.value = true
    } else {
        emit('delete')
    }
}

const confirmDelete = () => {
    confirmDeleteVisible.value = false
    emit('delete')
}

// --- Widgets ---
const updateWidget = (index: number, widget: Widget) => {
    const updated = [...props.row.widgets]
    updated[index] = widget
    emit('update', { ...props.row, widgets: updated })
}

const deleteWidget = (index: number) => {
    emit('update', { ...props.row, widgets: props.row.widgets.filter((_, i) => i !== index) })
}

const addWidgetDialogVisible = ref(false)

// A row is one 12-column line: once the widths add up to 12 nothing else fits.
const rowFull = computed(() => totalSpan(props.row.widgets) >= GRID_COLUMNS)

// The copy lands right after the original (the duplicate button is disabled
// on a full row; duplicateInRow returns the row unchanged anyway).
const duplicateWidget = (index: number) => {
    emit('update', { ...props.row, widgets: duplicateInRow(props.row.widgets, index, uuidv4) })
}

const addWidget = (type: string) => {
    // Drop the widget into the first free span so it never overlaps an
    // existing one, then splice it in at the matching column position.
    const span = firstFreeSpan(props.row.widgets, 6)
    if (span.column < 1) return // full: the add button is disabled, but guard anyway
    const newWidget: Widget = {
        id: uuidv4(),
        type,
        title: getWidgetEntry(type)?.label ?? 'New Widget',
        width: span.width,
        column: span.column,
    }
    emit('update', { ...props.row, widgets: insertByColumn(props.row.widgets, newWidget) })
}

// Resolved 12-column placement (leading gap + span) for every widget, with the
// in-progress resize preview folded in so the editor grid updates live.
const resizingIndex = ref<number | null>(null)
const resizePreviewWidth = ref<number | null>(null)

const placements = computed(() => {
    const list = props.row.widgets.map((widget, i) => {
        if (i === resizingIndex.value && resizePreviewWidth.value !== null) {
            return { width: resizePreviewWidth.value, column: widget.column }
        }
        return { width: widget.width, column: widget.column }
    })
    return placeRow(list)
})

const getWidgetClass = (index: number) => {
    const p = placements.value[index]
    return `col-${Math.min(GRID_COLUMNS, p.width)} col-offset-${Math.min(GRID_COLUMNS, p.gap)}`
}

// --- Drop guides (rendered from the shared session) ---
// A stack zone being targeted wins: then no row shows a guide.
const isDropTarget = computed(
    () => dragState.active && dragState.targetKind === 'row' && dragState.targetRowId === props.row.id,
)
const guideColumn = computed(() => (isDropTarget.value ? dragState.targetColumn : null))
const guideBlocked = computed(() => isDropTarget.value && !dragState.targetFits)

// --- Resize (pointer events, so touch and pen work like the mouse) ---
let resizeMax = GRID_COLUMNS
let resizeLeft = 0
let resizeColWidth = 1

const onResizeMove = (e: PointerEvent) => {
    if (resizingIndex.value === null) return
    const cols = Math.round((e.clientX - resizeLeft) / resizeColWidth)
    resizePreviewWidth.value = Math.max(1, Math.min(resizeMax, cols))
}

const onResizeUp = () => stopResize(true)
const onResizeCancel = () => stopResize(false)

function stopResize(commit: boolean) {
    document.removeEventListener('pointermove', onResizeMove)
    document.removeEventListener('pointerup', onResizeUp)
    document.removeEventListener('pointercancel', onResizeCancel)
    if (resizingIndex.value === null) return
    const index = resizingIndex.value
    const width = resizePreviewWidth.value
    resizingIndex.value = null
    resizePreviewWidth.value = null
    endDragGuard()
    if (commit && width !== null && width !== props.row.widgets[index]?.width) {
        updateWidget(index, { ...props.row.widgets[index], width })
    }
}

// Right edge → column span. Growing pushes right-hand neighbours along (they
// pack tighter first), but never past the grid: maxResizeWidth caps the span
// so the row stays on one line.
const startResize = (index: number, event: PointerEvent) => {
    event.preventDefault()
    const grid = gridEl()
    const widgetEl = grid?.children[index] as HTMLElement | undefined
    if (!grid || !widgetEl) return
    resizingIndex.value = index
    resizePreviewWidth.value = null
    resizeMax = maxResizeWidth(props.row.widgets, index)
    resizeLeft = widgetEl.getBoundingClientRect().left
    resizeColWidth = grid.getBoundingClientRect().width / GRID_COLUMNS
    beginDragGuard('col-resize')
    document.addEventListener('pointermove', onResizeMove)
    document.addEventListener('pointerup', onResizeUp)
    document.addEventListener('pointercancel', onResizeCancel)
}

// Keyboard alternative on the resize handle: one column per arrow press.
const nudgeWidth = (index: number, delta: number) => {
    const widget = props.row.widgets[index]
    if (!widget) return
    const width = Math.max(1, Math.min(maxResizeWidth(props.row.widgets, index), widget.width + delta))
    if (width !== widget.width) updateWidget(index, { ...widget, width })
}

// --- Move (grip drag) ---
// The true grab point, captured on pointer-down before SortableJS's drag
// threshold nudges it; the session anchors column deltas on it.
let grabX: number | null = null
let grabY: number | null = null

const onGripDown = (e: PointerEvent) => {
    grabX = e.clientX
    grabY = e.clientY
}

const resetGrab = () => {
    grabX = null
    grabY = null
}

const onDragStart = (evt: { oldIndex: number }) => {
    const widget = props.row.widgets[evt.oldIndex]
    const p = placeRow(props.row.widgets)[evt.oldIndex]
    const grid = gridEl()
    if (!widget || !p || !grid) {
        resetGrab()
        return
    }
    const rect = grid.getBoundingClientRect()
    const colWidth = rect.width / GRID_COLUMNS
    // A legacy row that overflowed the grid may report a column past 12; keep
    // the origin on the grid so a grab-and-release never jumps the widget.
    const originColumn = Math.min(p.column, Math.max(1, GRID_COLUMNS - p.width + 1))
    const originLeftX = rect.left + (originColumn - 1) * colWidth
    const startX = grabX ?? originLeftX + colWidth / 2
    const startY = grabY ?? rect.top
    resetGrab()
    startWidgetDrag(
        { kind: 'row', rowId: props.row.id, widgetId: widget.id },
        { originColumn, originLeftX, grabX: startX, grabY: startY, span: p.width, draggedType: widget.type },
    )
}

// The session resolved the drop (this row, another row or a stack); the
// editor applies it. Nothing is returned for Escape or a blocked target.
const onDragEnd = () => {
    resetGrab()
    const move = endWidgetDrag()
    if (move) moveWidget?.(move)
}

// Keyboard alternative on the grip: one column per arrow press.
const nudgeColumn = (index: number, delta: number) => {
    const p = placeRow(props.row.widgets)[index]
    if (!p) return
    const column = Math.max(1, Math.min(GRID_COLUMNS - p.width + 1, p.column + delta))
    if (column === p.column) return
    emit('update', { ...props.row, widgets: clampRowToLine(moveWidgetToColumn(props.row.widgets, index, column)) })
}
</script>

<template>
    <div
        ref="rootRef"
        class="dashboard-row-editor"
        :class="{
            'dashboard-row-editor--target': isDropTarget && !guideBlocked,
            'dashboard-row-editor--blocked': guideBlocked,
        }"
    >
        <div class="row-controls flex align-items-center gap-2 mb-2">
            <span v-if="row.title" class="row-title flex-grow-1">{{ row.title }}</span>
            <span v-else class="flex-grow-1"></span>
            <Button
                icon="ti ti-arrow-up"
                text
                rounded
                class="p-1"
                aria-label="Move row up"
                v-tooltip.top="'Move row up'"
                :disabled="isFirst"
                @click="emit('move-up')"
            />
            <Button
                icon="ti ti-arrow-down"
                text
                rounded
                class="p-1"
                aria-label="Move row down"
                v-tooltip.top="'Move row down'"
                :disabled="isLast"
                @click="emit('move-down')"
            />
            <Button
                icon="ti ti-plus"
                text
                rounded
                class="p-1"
                aria-label="Add widget"
                v-tooltip.top="rowFull ? 'Row is full (12 columns)' : 'Add widget'"
                :disabled="rowFull"
                @click="addWidgetDialogVisible = true"
            />
            <Button
                icon="ti ti-settings"
                text
                rounded
                class="p-1"
                aria-label="Row settings"
                v-tooltip.top="'Row settings'"
                @click="openSettings"
            />
            <Button
                icon="ti ti-trash"
                text
                rounded
                severity="danger"
                class="p-1"
                aria-label="Delete row"
                v-tooltip.top="'Delete row'"
                @click="requestDelete"
            />
        </div>
        <div class="grid-scroll">
        <div class="grid-wrap" :class="{ 'grid-wrap--empty': !row.widgets.length }">
            <div
                v-if="guideColumn !== null"
                class="column-guides"
                :class="{ 'column-guides--blocked': guideBlocked }"
                aria-hidden="true"
            >
                <div
                    v-for="c in GRID_COLUMNS"
                    :key="c"
                    class="guide-col"
                    :class="{ 'guide-col--active': c >= guideColumn && c < guideColumn + dragState.targetSpan }"
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
                :fallback-on-body="true"
                :fallback-tolerance="3"
                ghost-class="widget-col--source"
                :style="!row.widgets.length ? { minHeight: '72px' } : undefined"
                ref="gridRef"
                @start="onDragStart"
                @end="onDragEnd"
            >
                <template #item="{ element, index }">
                    <div :class="getWidgetClass(index)" class="widget-col">
                        <DashboardWidget
                            :widget="element"
                            :row-full="rowFull"
                            @update="updateWidget(index, $event)"
                            @delete="deleteWidget(index)"
                            @duplicate="duplicateWidget(index)"
                        >
                            <template #handle>
                                <span
                                    class="widget-drag-handle"
                                    role="button"
                                    tabindex="0"
                                    aria-label="Move widget (drag, or use the arrow keys)"
                                    v-tooltip.top="'Drag to move. Arrow keys nudge by one column, Esc cancels.'"
                                    @pointerdown="onGripDown"
                                    @mousedown.prevent
                                    @keydown.left.prevent="nudgeColumn(index, -1)"
                                    @keydown.right.prevent="nudgeColumn(index, 1)"
                                >
                                    <i class="ti ti-grip-vertical" aria-hidden="true" />
                                </span>
                            </template>
                            <template #resize>
                                <span
                                    class="resize-handle"
                                    role="button"
                                    tabindex="0"
                                    aria-label="Resize widget width (drag, or use the arrow keys)"
                                    v-tooltip.top="'Drag to resize width. Arrow keys resize by one column.'"
                                    @pointerdown="startResize(index, $event)"
                                    @keydown.left.prevent="nudgeWidth(index, -1)"
                                    @keydown.right.prevent="nudgeWidth(index, 1)"
                                >
                                    <i class="ti ti-arrows-horizontal" aria-hidden="true" />
                                </span>
                            </template>
                        </DashboardWidget>
                    </div>
                </template>
            </draggable>
            <button
                v-if="!row.widgets.length"
                type="button"
                class="empty-hint"
                @click="addWidgetDialogVisible = true"
            >
                <i class="ti ti-plus" aria-hidden="true" />
                Add a widget, or drop one here
            </button>
        </div>
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
                <label class="text-sm font-semibold" for="row-title">Title</label>
                <InputText id="row-title" v-model="editTitle" placeholder="Optional row title" />
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold" for="row-height">Height</label>
                <InputText id="row-height" v-model="editHeight" placeholder="e.g. auto, 300px, 50%" />
            </div>
            <div class="flex flex-column gap-1">
                <label class="text-sm font-semibold" for="row-width">Width</label>
                <InputText id="row-width" v-model="editWidth" placeholder="e.g. 100%, 1200px, 80%" />
            </div>
        </div>
        <div class="flex justify-content-end gap-3 mt-4">
            <Button label="Save" icon="ti ti-check" @click="saveSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="settingsVisible = false" />
        </div>
    </Dialog>

    <AddWidgetDialog v-model:visible="addWidgetDialogVisible" @select="addWidget" />

    <ConfirmDialog
        v-model:visible="confirmDeleteVisible"
        title="Delete row"
        :message="`Delete this row and its ${row.widgets.length} widget${row.widgets.length === 1 ? '' : 's'}? This cannot be undone.`"
        :name="row.title ?? ''"
        @confirm="confirmDelete"
    />
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.row-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--p-text-color);
    /* Wraps between words on a phone; no min-width: 0, so a word is never
       split just to make room for the row buttons. */
    overflow-wrap: break-word;
}

.dashboard-row-editor {
    background: var(--p-surface-50);
    border: 1px solid var(--p-surface-200);
    border-radius: 8px;
    padding: 0.75rem;
    margin-bottom: 0.75rem;
    transition: border-color 0.15s, box-shadow 0.15s;

    @include bp.below-md {
        padding: 0.5rem;
    }
}

/* Phones: the editor keeps the real 12-column grid (widths and offsets are
   what is being edited, so it does not reflow like the viewer does) but at a
   minimum width where a one-column card still fits its grip and edit button;
   the row scrolls sideways instead. The padding/negative margin pair leaves
   room for the PrimeFlex .grid gutter, which an overflow container would clip.
   Drag and resize read viewport rects and listen to scroll in the capture
   phase, so they keep working inside the scrolled row. */
.grid-scroll {
    @include bp.below-md {
        overflow-x: auto;
        margin: -0.5rem;
        padding: 0.5rem;
    }
}

/* The row the dragged widget would land in. */
.dashboard-row-editor--target {
    border-color: var(--p-primary-color);
    box-shadow: 0 0 0 1px var(--p-primary-color);
}

/* Hovered, but the widget does not fit on this row's line. */
.dashboard-row-editor--blocked {
    border-color: var(--p-red-400);
    box-shadow: 0 0 0 1px var(--p-red-400);
}

/* Drag/move grip, slotted into the widget card header. Slot content is compiled
   in this (parent) scope, so these styles apply even though it renders inside
   DashboardWidget. */
.widget-drag-handle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 6px;
    cursor: grab;
    color: var(--p-text-muted-color);
    user-select: none;
    touch-action: none;
}

.widget-drag-handle:hover,
.widget-drag-handle:focus-visible {
    background: var(--p-surface-100);
    color: var(--p-text-color);
    outline: none;
}

.widget-drag-handle:focus-visible {
    box-shadow: 0 0 0 2px var(--p-primary-color);
}

.widget-drag-handle:active {
    cursor: grabbing;
}

/* The card being dragged: dimmed in place while its clone follows the pointer. */
.widget-col--source {
    opacity: 0.35;
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
    touch-action: none;
    z-index: 10;
}

.resize-handle:hover,
.resize-handle:focus-visible {
    color: var(--p-primary-color);
    background: color-mix(in srgb, var(--p-primary-color) 12%, transparent);
    outline: none;
}

.resize-handle i {
    font-size: 0.875rem;
}

.grid-wrap {
    position: relative;

    @include bp.below-md {
        min-width: 36rem;
    }
}

/* 12-column guide overlay shown on the row the dragged widget targets.
   Matches the PrimeFlex .grid negative gutter so the dividers line up with
   the columns. pointer-events: none keeps it clear of the drag. */
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

.column-guides--blocked .guide-col--active {
    background: repeating-linear-gradient(
        -45deg,
        color-mix(in srgb, var(--p-red-500) 18%, transparent) 0 6px,
        transparent 6px 12px
    );
}

/* Placeholder for an empty row: a real button, so it is keyboard reachable. */
.empty-hint {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    border: 1px dashed var(--p-surface-300);
    border-radius: 8px;
    background: none;
    color: var(--p-text-muted-color);
    font: inherit;
    cursor: pointer;
}

.empty-hint:hover,
.empty-hint:focus-visible {
    background: var(--p-surface-100);
    color: var(--p-text-color);
    outline: none;
}
</style>

<style>
/* SortableJS clones the dragged cell — including its col-offset-N margin — as
   the floating fallback drag preview. PrimeFlex sets that margin with
   !important, which beats the inline reset SortableJS applies, and on the
   fixed-position clone it would shove the preview far to the right. Not
   scoped: the clone is appended to <body>, outside the component subtree. */
.sortable-fallback {
    margin-left: 0 !important;
    /* The cloned cell is stretched to the row's full height by the flex grid;
       let the preview shrink to the card it carries. */
    height: auto !important;
    box-shadow: 0 10px 28px rgba(0, 0, 0, 0.2);
    border-radius: 8px;
}
</style>
