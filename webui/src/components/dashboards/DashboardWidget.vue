<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import WidgetRenderer from '@/components/dashboards/WidgetRenderer.vue'
import WidgetSettingsDialog from '@/components/dashboards/WidgetSettingsDialog.vue'
import type { Widget } from '@/types/dashboard'

defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    update: [widget: Widget]
    delete: []
}>()

const settingsVisible = ref(false)
</script>

<template>
    <div class="dashboard-widget">
        <div class="widget-header">
            <slot name="handle" />
            <span class="widget-title" :title="widget.title">{{ widget.title }}</span>
            <Button
                icon="ti ti-pencil"
                text
                rounded
                class="p-1"
                aria-label="Widget settings"
                v-tooltip.top="'Widget settings'"
                @click="settingsVisible = true"
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
        @delete="emit('delete')"
    />
</template>

<style scoped>
.dashboard-widget {
    position: relative;
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
}

/* Grip (drag handle) sits at the left, the edit button at the right, so both
   stay reachable even when the card is a single column wide. The title in
   between identifies the card; it is clipped before the controls are. Delete
   lives in the settings dialog rather than the header. */
.widget-header {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    min-height: 2rem;
    padding: 0.125rem 0.375rem;
    border-bottom: 1px solid var(--p-surface-200);
}

.widget-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.8rem;
    color: var(--p-text-muted-color);
}

.widget-body {
    padding: 0.5rem 0.75rem 1.5rem;
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
