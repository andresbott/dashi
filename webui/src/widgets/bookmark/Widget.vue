<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types/dashboard'
import type { BookmarkWidgetConfig } from './types'
import BookmarkLink from './BookmarkLink.vue'

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

        <BookmarkLink
            v-else
            :url="config.url"
            :icon="config.icon"
            :title="config.title"
            :subtitle="config.subtitle"
            :textBelow="config.textBelow"
            :newTab="config.newTab"
        />
    </div>
</template>

<style scoped>
.bookmark-widget {
    padding: 0.5rem;
    min-height: 60px;
    overflow: hidden;
}

.bookmark-widget:has(.bookmark-link--below) {
    padding: 0;
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
</style>
