<script setup lang="ts">
import { computed, provide, useId } from 'vue'
import WidgetRenderer from '@/components/dashboards/WidgetRenderer.vue'
import { getWidgetDocs } from '@/lib/widgetDocs'
import type { Widget } from '@/types/dashboard'

// Renders a widget's real editor component for the docs. Data comes from the
// mock QueryClient the docs page provides (see lib/docQueryClient.ts); this
// component adds the injections the widget's docs ask for. Mount it with a
// :key per type, since the injections are provided once.
const props = withDefaults(defineProps<{
    type: string
    config?: Record<string, unknown>
    // No frame: the host (an index card) is the frame.
    bare?: boolean
}>(), {
    config: undefined,
    bare: false,
})

const docs = getWidgetDocs(props.type)
for (const [key, value] of docs?.preview.provide ?? []) provide(key, value)

// Unique per instance: some widgets derive SVG gradient ids from it.
const id = `doc-preview-${useId()}`

const widget = computed<Widget>(() => ({
    id,
    type: props.type,
    title: docs?.preview.title ?? props.type,
    width: 4,
    config: props.config,
}))
</script>

<template>
    <div
        class="doc-preview"
        :class="{ 'doc-preview--bare': bare, 'doc-preview--fixed': !!docs?.preview.height }"
        :style="docs?.preview.height ? { height: docs.preview.height } : undefined"
        data-testid="widget-preview"
    >
        <WidgetRenderer :widget="widget" />
    </div>
</template>

<style scoped>
.doc-preview {
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
    padding: 0.5rem 0.75rem;
    overflow-x: clip;
}

.doc-preview--bare {
    background: none;
    border: 0;
    border-radius: 0;
    padding: 0;
}

/* A widget that fills a fixed height (image, xkcd) shrinks to the card;
   the others keep their natural height and the card clips them. */
.doc-preview--bare.doc-preview--fixed {
    max-height: 100%;
}
</style>
