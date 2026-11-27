<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types/dashboard'
import BookmarkLink from '../bookmark/BookmarkLink.vue'
import { useBookmarkFile } from '@/composables/useBookmarkFiles'
import { MAX_GROUP_COLUMNS } from './types'
import type { BookmarkGroupWidgetConfig } from './types'

const props = defineProps<{
    widget: Widget
}>()

const config = computed(() => (props.widget.config ?? {}) as unknown as Partial<BookmarkGroupWidgetConfig>)
// Source "file" reads the bookmarks from a content file (the server does
// the same through its cache); anything else uses the dashboard's own list.
const fromFile = computed(() => config.value.source === 'file')
const fileName = computed(() => (fromFile.value ? config.value.file ?? '' : ''))
const { bookmarks: fileBookmarks, isLoading: fileLoading, isError: fileError } = useBookmarkFile(fileName)
const bookmarks = computed(() => (fromFile.value ? fileBookmarks.value : config.value.bookmarks ?? []))
const fileMessage = computed(() => {
    if (!fromFile.value) return ''
    if (!fileName.value) return 'No bookmarks file selected'
    if (fileError.value) return `Bookmarks file "${fileName.value}" could not be loaded`
    return ''
})
// Anything but "vertical" lays out as a wrapping row, as the server does.
const direction = computed(() => (config.value.direction === 'vertical' ? 'vertical' : 'horizontal'))

// Same columns the server renders (internal/widgets/bookmark/browser.html):
// one per item when vertical, a fixed count when set, else as many as fit.
const columns = computed(() => {
    const n = config.value.columns
    if (direction.value === 'vertical' || typeof n !== 'number' || !Number.isInteger(n) || n <= 0) return 0
    return Math.min(n, MAX_GROUP_COLUMNS)
})
const gridTemplateColumns = computed(() => {
    if (direction.value === 'vertical') return 'minmax(0, 1fr)'
    if (columns.value > 0) return `repeat(${columns.value}, minmax(0, 1fr))`
    return 'repeat(auto-fit, minmax(min(100%, 10rem), 1fr))'
})
</script>

<template>
    <div class="bookmarkgroup-widget">
        <div v-if="fileMessage" class="bookmarkgroup-empty bookmarkgroup-file-error">
            <i class="ti ti-file-alert" />
            <span>{{ fileMessage }}</span>
        </div>

        <div v-else-if="bookmarks.length === 0" class="bookmarkgroup-empty">
            <i class="ti ti-bookmarks" />
            <span v-if="fromFile">{{ fileLoading ? 'Loading bookmarks…' : 'The bookmarks file is empty' }}</span>
            <span v-else>Add bookmarks in edit mode</span>
        </div>

        <div
            v-else
            class="bookmarkgroup-list"
            :class="`bookmarkgroup-list--${direction}`"
            :data-direction="direction"
            :data-columns="columns || undefined"
            :style="{ gridTemplateColumns }"
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

/* Spacing as the viewer renders it: no column gap (each link's own padding
   separates them), 1rem between rows. */
.bookmarkgroup-list {
    display: grid;
    column-gap: 0;
    row-gap: 1rem;
    min-width: 0;
}

/* A text-below link is a small square on its own; in a group it fills its
   column, as in the viewer, so titles are not cut short. */
.bookmarkgroup-list .bookmark-link--below {
    max-width: none;
    aspect-ratio: auto;
    width: 100%;
}
</style>
