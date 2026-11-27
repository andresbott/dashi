<script setup lang="ts">
import { computed, ref, inject, onMounted, onBeforeUnmount } from 'vue'
import draggable from 'vuedraggable'
import Button from 'primevue/button'
import WidgetRenderer from '@/components/dashboards/WidgetRenderer.vue'
import AddWidgetDialog from '@/components/dashboards/AddWidgetDialog.vue'
import WidgetSettingsDialog from '@/components/dashboards/WidgetSettingsDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import { abortIfSource, dragState, endWidgetDrag, registerStackZone, startWidgetDrag } from '@/lib/widgetDrag'
import type { Widget } from '@/types/dashboard'
import { EDITING_MODE, WIDGET_MOVE } from '@/lib/injectionKeys'
import { v4 as uuidv4 } from 'uuid'

const props = defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    'update:widget': [widget: Widget]
}>()

const isEditing = inject(EDITING_MODE, false)

// A stack cannot contain another stack: the picker hides the type, so the
// children list never has to filter it out after the fact.
const EXCLUDED_CHILD_TYPES = ['stack']

// Commits a drop; provided by the editor page (absent elsewhere: drops do nothing).
const moveWidget = inject(WIDGET_MOVE, null)

// The children live in the stack's config and the prop is the single source
// of truth: add, edit and delete are emitted up and come back through the
// prop. Drags never touch the list here: the draggable is passive and every
// drop (reorder, in, out) is committed page-wide through WIDGET_MOVE.
const children = computed<Widget[]>(() => {
    const config = props.widget.config as { widgets?: Widget[] } | undefined
    return config?.widgets ?? []
})

const setChildren = (list: Widget[]) => {
    emit('update:widget', {
        ...props.widget,
        config: { ...props.widget.config, widgets: list },
    })
}

// Children always span the stack: they inherit its width in both modes.
const childView = (child: Widget): Widget => ({ ...child, width: props.widget.width })

// --- Add child ---
const addDialogVisible = ref(false)

const addChild = (type: string) => {
    const newChild: Widget = {
        id: uuidv4(),
        type,
        title: getWidgetEntry(type)?.label ?? 'New Widget',
        width: props.widget.width,
    }
    setChildren([...children.value, newChild])
}

// --- Child settings ---
const settingsVisible = ref(false)
const editingIndex = ref<number | null>(null)
const editingChild = computed(() =>
    editingIndex.value === null ? null : (children.value[editingIndex.value] ?? null),
)

const openChildSettings = (index: number) => {
    editingIndex.value = index
    settingsVisible.value = true
}

const saveChild = (updated: Widget) => {
    if (editingIndex.value === null) return
    const list = [...children.value]
    list[editingIndex.value] = updated
    setChildren(list)
}

// --- Delete child (confirmed) ---
const confirmDeleteVisible = ref(false)
const deletingIndex = ref<number | null>(null)
const deletingChild = computed(() =>
    deletingIndex.value === null ? null : (children.value[deletingIndex.value] ?? null),
)

const askDeleteChild = (index: number) => {
    deletingIndex.value = index
    confirmDeleteVisible.value = true
}

const deleteChild = () => {
    confirmDeleteVisible.value = false
    if (deletingIndex.value === null) return
    setChildren(children.value.filter((_, i) => i !== deletingIndex.value))
    deletingIndex.value = null
}

// --- Drag in, out and within (shared drag session, lib/widgetDrag.ts) ---
const rootRef = ref<HTMLElement | null>(null)
const listRef = ref<InstanceType<typeof draggable> | null>(null)

// The drop zone is the stack card's whole body (its padding included), so a
// drop just below the last child still lands in the stack; the card header
// and the gaps around the card target the row.
const zoneEl = () => rootRef.value?.closest<HTMLElement>('.widget-body') ?? rootRef.value
const childEls = () => {
    const list = listRef.value?.$el as HTMLElement | undefined
    return list ? Array.from(list.querySelectorAll<HTMLElement>(':scope > .stack-item > .stack-child-editor')) : []
}

let unregisterZone: (() => void) | null = null

onMounted(() => {
    if (isEditing) unregisterZone = registerStackZone(props.widget.id, { zone: zoneEl, children: childEls })
})

onBeforeUnmount(() => {
    unregisterZone?.()
    // A stack removed mid-gesture must not leave a stuck global drag behind.
    abortIfSource(props.widget.id)
})

// The true grab point, captured on pointer-down before SortableJS's drag
// threshold nudges it.
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

// Index of the child being dragged out of this stack, for the marker.
const dragSourceIndex = ref<number | null>(null)

const onDragStart = (evt: { oldIndex?: number; item: HTMLElement }) => {
    const index = evt.oldIndex ?? -1
    const child = children.value[index]
    if (!child) {
        resetGrab()
        return
    }
    const rect = evt.item.getBoundingClientRect()
    dragSourceIndex.value = index
    startWidgetDrag(
        { kind: 'stack', stackId: props.widget.id, widgetId: child.id },
        {
            originColumn: 1,
            originLeftX: rect.left,
            grabX: grabX ?? rect.left + rect.width / 2,
            grabY: grabY ?? rect.top + rect.height / 2,
            span: props.widget.width,
            draggedType: child.type,
            sourceIndex: index,
        },
    )
    resetGrab()
}

// The session resolved the drop (this stack, another stack or a row); the
// editor applies it. Nothing is returned for Escape or a blocked row.
const onDragEnd = () => {
    resetGrab()
    dragSourceIndex.value = null
    const move = endWidgetDrag()
    if (move) moveWidget?.(move)
}

const isDropTarget = computed(
    () => dragState.active && dragState.targetKind === 'stack' && dragState.targetStackId === props.widget.id,
)

// Where the insert marker shows, counted in the rendered list. The session's
// index is counted after the dragged child is removed, but that child is
// still on screen (dimmed) while it is dragged within this stack.
const markerIndex = computed(() => {
    if (!isDropTarget.value || dragState.targetIndex === null) return null
    const index = dragState.targetIndex
    const from = dragState.sourceStackId === props.widget.id ? dragSourceIndex.value : null
    return from !== null && index >= from ? index + 1 : index
})
</script>

<template>
    <!-- View mode -->
    <div v-if="!isEditing" class="widget-stack">
        <div v-for="child in children" :key="child.id" class="stack-child">
            <WidgetRenderer :widget="childView(child)" />
        </div>
    </div>

    <!-- Edit mode -->
    <div
        v-else
        ref="rootRef"
        class="widget-stack-editor"
        :class="{ 'widget-stack-editor--target': isDropTarget }"
    >
        <div class="stack-children-wrap">
            <draggable
                ref="listRef"
                :model-value="children"
                :group="{ name: 'stack-children', pull: false, put: false }"
                item-key="id"
                handle=".stack-drag-handle"
                class="stack-children"
                :sort="false"
                :move="() => false"
                :force-fallback="true"
                :fallback-on-body="true"
                :fallback-tolerance="3"
                ghost-class="stack-item--source"
                @start="onDragStart"
                @end="onDragEnd"
            >
                <template #item="{ element, index }">
                    <div class="stack-item">
                        <div v-if="markerIndex === index" class="stack-insert-marker" aria-hidden="true" />
                        <div class="stack-child-editor">
                            <div class="stack-child-controls flex align-items-center gap-1">
                                <span
                                    class="stack-drag-handle"
                                    role="button"
                                    tabindex="0"
                                    aria-label="Move widget (drag to reorder, or out of the stack)"
                                    v-tooltip.top="'Drag to reorder, or drag out of the stack onto a row. Esc cancels.'"
                                    @pointerdown="onGripDown"
                                    @mousedown.prevent
                                >
                                    <i class="ti ti-grip-vertical" aria-hidden="true" />
                                </span>
                                <span class="text-sm flex-grow-1 stack-child-label">
                                    {{ element.title }}
                                    <span class="stack-child-type">({{ element.type }})</span>
                                </span>
                                <Button
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1"
                                    aria-label="Widget settings"
                                    v-tooltip.top="'Widget settings'"
                                    @click="openChildSettings(index)"
                                />
                                <Button
                                    icon="ti ti-trash"
                                    text
                                    rounded
                                    severity="danger"
                                    class="p-1"
                                    aria-label="Remove widget from stack"
                                    v-tooltip.top="'Remove from stack'"
                                    @click="askDeleteChild(index)"
                                />
                            </div>
                            <div class="stack-child-preview">
                                <WidgetRenderer :widget="childView(element)" />
                            </div>
                        </div>
                    </div>
                </template>
            </draggable>
            <!-- Drop feedback: an empty stack shows a placeholder over its
                 (min-height) list, otherwise the marker sits past the last child. -->
            <div v-if="markerIndex !== null && !children.length" class="stack-drop-placeholder" aria-hidden="true">
                <div class="stack-insert-marker" />
            </div>
            <div v-else-if="markerIndex !== null && markerIndex >= children.length" class="stack-end-marker" aria-hidden="true">
                <div class="stack-insert-marker" />
            </div>
        </div>

        <div class="flex align-items-center gap-2 mt-1">
            <Button
                icon="ti ti-plus"
                label="Add"
                text
                size="small"
                aria-label="Add widget to stack"
                @click="addDialogVisible = true"
            />
            <span class="stack-hint">Drag widgets in or out of the stack.</span>
        </div>
    </div>

    <AddWidgetDialog
        v-model:visible="addDialogVisible"
        header="Add Widget to Stack"
        :exclude="EXCLUDED_CHILD_TYPES"
        @select="addChild"
    />

    <WidgetSettingsDialog
        v-if="editingChild"
        v-model:visible="settingsVisible"
        :widget="editingChild"
        @save="saveChild"
    />

    <ConfirmDialog
        v-model:visible="confirmDeleteVisible"
        title="Remove widget"
        message="Remove this widget from the stack? This cannot be undone."
        :name="deletingChild?.title ?? ''"
        @confirm="deleteChild"
    />
</template>

<style scoped>
.widget-stack-editor {
    min-height: 3rem;
    padding-top: 0.25rem;
}

/* The stack the dragged widget would land in (same colour as a target row). */
.widget-stack-editor--target {
    border-radius: 6px;
    outline: 1px solid var(--p-primary-color);
    outline-offset: 0.25rem;
}

/* Positioning context for the drop feedback, which never shifts the layout:
   the session measures the children on every pointer move. */
.stack-children-wrap {
    position: relative;
}

.stack-children {
    min-height: 2rem;
}

.stack-item {
    position: relative;
}

.stack-child-editor {
    background: var(--p-surface-50);
    border: 1px solid var(--p-surface-200);
    border-radius: 6px;
    padding: 0.25rem 0.5rem;
    margin-bottom: 0.25rem;
}

/* The child being dragged: dimmed in place while its clone follows the pointer. */
.stack-item--source {
    opacity: 0.35;
}

/* Insert marker: a line in the gap above the child it lands before. */
.stack-insert-marker {
    position: absolute;
    left: 0;
    right: 0;
    top: -0.1875rem;
    height: 2px;
    border-radius: 1px;
    background: var(--p-primary-color);
    pointer-events: none;
    z-index: 3;
}

.stack-end-marker {
    position: relative;
}

/* Empty stack under the pointer: the row guide's dashed border and tint. */
.stack-drop-placeholder {
    position: absolute;
    inset: 0;
    border: 1px dashed var(--p-surface-400);
    border-radius: 6px;
    background: color-mix(in srgb, var(--p-primary-color) 16%, transparent);
    pointer-events: none;
}

.stack-drop-placeholder .stack-insert-marker {
    top: 0.25rem;
    left: 0.25rem;
    right: 0.25rem;
}

.stack-drag-handle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 4px;
    cursor: grab;
    color: var(--p-text-muted-color);
    user-select: none;
    touch-action: none;
}

.stack-drag-handle:hover,
.stack-drag-handle:focus-visible {
    background: var(--p-surface-100);
    color: var(--p-text-color);
    outline: none;
}

.stack-child-label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.stack-child-type {
    color: var(--p-text-muted-color);
}

.stack-child-preview {
    pointer-events: none;
    opacity: 0.8;
}

.stack-hint {
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}
</style>
