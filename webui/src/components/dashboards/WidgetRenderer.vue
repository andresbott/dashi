<script setup lang="ts">
import { computed } from 'vue'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import type { Widget } from '@/types/dashboard'

// Renders a widget's registered component. The one place that knows about the
// `noWidgetProp` registry flag, so hosts (row cards, stack children) never
// repeat the lookup-and-bind dance.
const props = defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    'update:widget': [widget: Widget]
}>()

const entry = computed(() => getWidgetEntry(props.widget.type))
</script>

<template>
    <component
        v-if="entry"
        :is="entry.component"
        v-bind="entry.noWidgetProp ? {} : { widget }"
        @update:widget="emit('update:widget', $event)"
    />
    <div v-else class="widget-unknown">Unknown widget type: {{ widget.type }}</div>
</template>

<style scoped>
.widget-unknown {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
}
</style>
