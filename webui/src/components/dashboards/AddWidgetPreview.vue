<script setup lang="ts">
import { provide } from 'vue'
import WidgetPreview from '@/components/docs/WidgetPreview.vue'
import { allDocMocks, getWidgetDocs } from '@/lib/widgetDocs'
import { provideDocQueryClient } from '@/lib/docQueryClient'
import { EDITING_MODE } from '@/lib/injectionKeys'

// The Add Widget dialog's preview: the widget rendered from its docs sample
// config, like the widget docs pages. The mock QueryClient is provided to this
// subtree only (and cleared when it unmounts), so the editor's own client is
// never touched. The editor provides editing mode to everything below it;
// switch it off so a preview renders as on a dashboard (a stack shows its
// children, not its drag editor). Mount with a :key per type: WidgetPreview
// provides its injections once.
const props = defineProps<{
    type: string
}>()

provideDocQueryClient(allDocMocks())
provide(EDITING_MODE, false)

const config = getWidgetDocs(props.type)?.preview.config
</script>

<template>
    <div class="add-widget-preview">
        <WidgetPreview :type="type" :config="config" />
        <span class="add-widget-preview-caption text-xs">Sample data</span>
    </div>
</template>

<style scoped>
.add-widget-preview {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
}

.add-widget-preview-caption {
    color: var(--p-text-muted-color);
    align-self: flex-end;
}
</style>
