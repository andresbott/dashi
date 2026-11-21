<script setup lang="ts">
import { computed, ref, inject, onBeforeUnmount } from 'vue'
import draggable from 'vuedraggable'
import Button from 'primevue/button'
import WidgetRenderer from '@/components/dashboards/WidgetRenderer.vue'
import AddWidgetDialog from '@/components/dashboards/AddWidgetDialog.vue'
import WidgetSettingsDialog from '@/components/dashboards/WidgetSettingsDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import { beginDragGuard, endDragGuard } from '@/lib/dragGuard'
import type { Widget } from '@/types/dashboard'
import { EDITING_MODE } from '@/lib/injectionKeys'
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

// The children live in the stack's config and the prop is the single source
// of truth: every change is emitted up and comes back through the prop. The
// draggable is bound one-way for the same reason (its own cross-list moves
// are disabled, so there is nothing that needs a local copy).
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

// --- Reorder (drag) ---
const onReorderStart = () => beginDragGuard('grabbing')
const onReorderEnd = () => endDragGuard()
onBeforeUnmount(endDragGuard)
</script>

<template>
    <!-- View mode -->
    <div v-if="!isEditing" class="widget-stack">
        <div v-for="child in children" :key="child.id" class="stack-child">
            <WidgetRenderer :widget="childView(child)" />
        </div>
    </div>

    <!-- Edit mode -->
    <div v-else class="widget-stack-editor">
        <draggable
            :model-value="children"
            :group="{ name: 'stack-children', pull: false, put: false }"
            item-key="id"
            handle=".stack-drag-handle"
            class="stack-children"
            :force-fallback="true"
            :fallback-on-body="true"
            :fallback-tolerance="3"
            ghost-class="stack-child-editor--source"
            @update:model-value="setChildren"
            @start="onReorderStart"
            @end="onReorderEnd"
        >
            <template #item="{ element, index }">
                <div class="stack-child-editor">
                    <div class="stack-child-controls flex align-items-center gap-1">
                        <span
                            class="stack-drag-handle"
                            role="button"
                            tabindex="0"
                            aria-label="Reorder widget within the stack"
                            v-tooltip.top="'Drag to reorder within the stack'"
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
            </template>
        </draggable>

        <div class="flex align-items-center gap-2 mt-1">
            <Button
                icon="ti ti-plus"
                label="Add"
                text
                size="small"
                aria-label="Add widget to stack"
                @click="addDialogVisible = true"
            />
            <span class="stack-hint">Children stay inside the stack.</span>
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
        :deletable="false"
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

.stack-children {
    min-height: 2rem;
}

.stack-child-editor {
    background: var(--p-surface-50);
    border: 1px solid var(--p-surface-200);
    border-radius: 6px;
    padding: 0.25rem 0.5rem;
    margin-bottom: 0.25rem;
}

/* The child being dragged: dimmed in place while its clone follows the pointer. */
.stack-child-editor--source {
    opacity: 0.35;
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
