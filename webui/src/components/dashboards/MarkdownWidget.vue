<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import { useMarkdown } from '@/composables/useMarkdown'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))

const filename = computed(() => {
    const cfg = props.widget.config as { filename?: string } | null | undefined
    return cfg?.filename ?? ''
})

const { data: html, isLoading, isError } = useMarkdown(dashboardId, filename)
</script>

<template>
    <div class="markdown-widget">
        <div v-if="!filename" class="markdown-empty">
            <i class="ti ti-markdown" />
            <span>Configure in edit mode</span>
        </div>

        <div v-else-if="isLoading" class="markdown-empty">
            <i class="ti ti-loader-2 markdown-spinner" />
        </div>

        <div v-else-if="isError" class="markdown-empty">
            <i class="ti ti-file-off" />
            <span>Failed to load markdown</span>
        </div>

        <div v-else-if="html" class="markdown-content" v-html="html" />
    </div>
</template>

<style scoped>
.markdown-widget {
    padding: 0.75rem;
    min-height: 60px;
}

.markdown-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.markdown-empty .ti {
    font-size: 1.5rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.markdown-spinner {
    animation: spin 1s linear infinite;
}

.markdown-content :deep(h1) {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0 0 0.5rem 0;
}

.markdown-content :deep(h2) {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0.75rem 0 0.375rem 0;
}

.markdown-content :deep(h3) {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0.5rem 0 0.25rem 0;
}

.markdown-content :deep(p) {
    margin: 0 0 0.5rem 0;
    line-height: 1.5;
}

.markdown-content :deep(a) {
    color: var(--p-primary-color);
    text-decoration: none;
}

.markdown-content :deep(a:hover) {
    text-decoration: underline;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
    margin: 0 0 0.5rem 0;
    padding-left: 1.5rem;
}

.markdown-content :deep(li) {
    margin-bottom: 0.25rem;
    line-height: 1.5;
}

.markdown-content :deep(code) {
    background: var(--p-surface-100);
    padding: 0.125rem 0.375rem;
    border-radius: 4px;
    font-size: 0.875em;
}

.markdown-content :deep(pre) {
    background: var(--p-surface-100);
    padding: 0.75rem;
    border-radius: 6px;
    overflow-x: auto;
    margin: 0 0 0.5rem 0;
}

.markdown-content :deep(pre code) {
    background: none;
    padding: 0;
}

.markdown-content :deep(blockquote) {
    border-left: 3px solid var(--p-primary-color);
    margin: 0 0 0.5rem 0;
    padding: 0.25rem 0 0.25rem 0.75rem;
    color: var(--p-text-muted-color);
}
</style>
