<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import WidgetRenderer from '@/components/dashboards/WidgetRenderer.vue'
import WidgetSettingsDialog from '@/components/dashboards/WidgetSettingsDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { Widget } from '@/types/dashboard'

defineProps<{
    widget: Widget
    // The card's row has no free column left, so a copy cannot be placed.
    rowFull?: boolean
}>()

const emit = defineEmits<{
    update: [widget: Widget]
    delete: []
    duplicate: []
}>()

const settingsVisible = ref(false)

// --- Delete (confirmed) ---
const confirmDeleteVisible = ref(false)

const confirmDelete = () => {
    confirmDeleteVisible.value = false
    emit('delete')
}
</script>

<template>
    <div class="dashboard-widget">
        <div class="widget-header">
            <slot name="handle" />
            <span class="widget-title" :title="widget.title">{{ widget.title }}</span>
            <Button
                icon="ti ti-copy"
                text
                rounded
                class="p-1"
                aria-label="Duplicate widget"
                v-tooltip.top="rowFull ? 'Row is full (12 columns)' : 'Duplicate widget'"
                :disabled="rowFull"
                @click="emit('duplicate')"
            />
            <Button
                icon="ti ti-pencil"
                text
                rounded
                class="p-1"
                aria-label="Widget settings"
                v-tooltip.top="'Widget settings'"
                @click="settingsVisible = true"
            />
            <Button
                icon="ti ti-trash"
                text
                rounded
                severity="danger"
                class="p-1"
                aria-label="Delete widget"
                v-tooltip.top="'Delete widget'"
                @click="confirmDeleteVisible = true"
            />
        </div>
        <div class="widget-body">
            <WidgetRenderer :widget="widget" @update:widget="emit('update', $event)" />
        </div>
        <span class="widget-width-label">{{ widget.width }}/12</span>
        <slot name="resize" />
    </div>

    <WidgetSettingsDialog
        v-model:visible="settingsVisible"
        :widget="widget"
        @save="emit('update', $event)"
    />

    <ConfirmDialog
        v-model:visible="confirmDeleteVisible"
        title="Delete widget"
        message="Delete this widget? This cannot be undone."
        :name="widget.title"
        @confirm="confirmDelete"
    />
</template>

<style scoped>
.dashboard-widget {
    position: relative;
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
}

/* Grip (drag handle) sits at the left, the widget controls (duplicate, edit,
   delete) at the right. The title in between identifies the card; it is clipped before
   the controls are. On a card too narrow for every control the header wraps
   (the title's zero flex-basis means it never forces a wrap by itself), so
   each control stays reachable even at a single column wide. */
.widget-header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.25rem;
    min-height: 2rem;
    padding: 0.125rem 0.375rem;
    border-bottom: 1px solid var(--p-surface-200);
}

.widget-title {
    flex: 1 1 0;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.8rem;
    color: var(--p-text-muted-color);
}

.widget-body {
    padding: 0.5rem 0.75rem 1.5rem;
    /* A preview wider than a narrow card (a big clock in 3 of 12 columns) is
       clipped instead of spilling over its neighbours. clip, not hidden: no
       scroll container, so the vertical axis is left alone. */
    overflow-x: clip;
}

.widget-width-label {
    position: absolute;
    bottom: 0.375rem;
    left: 0.625rem;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
    pointer-events: none;
}
</style>
