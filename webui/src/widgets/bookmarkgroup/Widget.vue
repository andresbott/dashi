<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types/dashboard'
import BookmarkLink from '../bookmark/BookmarkLink.vue'
import type { BookmarkGroupWidgetConfig } from './types'

const props = defineProps<{
    widget: Widget
}>()

const config = computed(() => (props.widget.config ?? {}) as unknown as Partial<BookmarkGroupWidgetConfig>)
const bookmarks = computed(() => config.value.bookmarks ?? [])
// Anything but "vertical" lays out as a wrapping row, as the server does.
const direction = computed(() => (config.value.direction === 'vertical' ? 'vertical' : 'horizontal'))
</script>

<template>
    <div class="bookmarkgroup-widget">
        <div v-if="bookmarks.length === 0" class="bookmarkgroup-empty">
            <i class="ti ti-bookmarks" />
            <span>Add bookmarks in edit mode</span>
        </div>

        <div
            v-else
            class="bookmarkgroup-list"
            :class="`bookmarkgroup-list--${direction}`"
            :data-direction="direction"
        >
            <BookmarkLink
                v-for="(bookmark, i) in bookmarks"
                :key="i"
                :url="bookmark.url"
                :icon="bookmark.icon"
                :title="bookmark.title"
                :subtitle="bookmark.subtitle"
                :textBelow="config.textBelow"
                :newTab="config.newTab"
                :class="{ 'bookmarkgroup-tile': config.textBelow && direction === 'horizontal' }"
            />
        </div>
    </div>
</template>

<style scoped>
.bookmarkgroup-widget {
    padding: 0.5rem;
    min-height: 60px;
    overflow: hidden;
}

.bookmarkgroup-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.bookmarkgroup-empty .ti {
    font-size: 1.5rem;
}

.bookmarkgroup-list {
    display: flex;
    gap: 0.5rem;
    min-width: 0;
}

.bookmarkgroup-list--horizontal {
    flex-direction: row;
    flex-wrap: wrap;
}

.bookmarkgroup-list--vertical {
    flex-direction: column;
    gap: 0.25rem;
}

/* A text-below link centres itself in a single bookmark; in a row of tiles it
   must sit at a fixed width instead, so the tiles line up in columns. */
.bookmarkgroup-tile {
    width: 6rem;
    margin: 0;
}
</style>
