<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types/dashboard'
import type { BookmarkWidgetConfig } from '@/types/bookmark'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<BookmarkWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    const c = props.widget.config as unknown as BookmarkWidgetConfig
    if (!c.url || !c.title) return null
    return c
})
</script>

<template>
    <div class="bookmark-widget">
        <div v-if="!config" class="bookmark-empty">
            <i class="ti ti-bookmark-question" />
            <span>Configure bookmark in edit mode</span>
        </div>

        <a v-else :href="config.url" target="_blank" rel="noopener noreferrer" class="bookmark-link">
            <i :class="'ti ' + config.icon" class="bookmark-icon" />
            <div class="bookmark-text">
                <span class="bookmark-title">{{ config.title }}</span>
                <span v-if="config.subtitle" class="bookmark-subtitle">{{ config.subtitle }}</span>
            </div>
        </a>
    </div>
</template>

<style scoped>
.bookmark-widget {
    padding: 0.5rem;
    min-height: 60px;
}

.bookmark-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.bookmark-empty .ti {
    font-size: 1.5rem;
}

.bookmark-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem;
    border-radius: 6px;
    text-decoration: none;
    color: inherit;
    transition: background 0.15s;
}

.bookmark-link:hover {
    background: var(--p-surface-100);
}

.bookmark-icon {
    font-size: 2rem;
    color: var(--p-primary-color);
    flex-shrink: 0;
}

.bookmark-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.bookmark-title {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.bookmark-subtitle {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
