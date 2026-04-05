<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import type { Ref } from 'vue'
import type { Widget } from '@/types/dashboard'
import type { BookmarkWidgetConfig } from '@/types/bookmark'
import { parseIcon, getSelfhstIconUrl, getDashboardIconUrl } from '@/lib/iconUtils'

const props = defineProps<{
    widget: Widget
}>()

const dashboardId = inject<Ref<string>>('dashboardId', ref(''))

const config = computed<BookmarkWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    const c = props.widget.config as unknown as BookmarkWidgetConfig
    if (!c.url || !c.title) return null
    return c
})

const iconInfo = computed(() => {
    if (!config.value?.icon) return null
    const parsed = parseIcon(config.value.icon)
    if (parsed.type === 'selfhst') {
        return { type: 'image' as const, src: getSelfhstIconUrl(parsed.value) }
    }
    if (parsed.type === 'dashboard') {
        return { type: 'image' as const, src: getDashboardIconUrl(dashboardId.value, parsed.value) }
    }
    return { type: 'tabler' as const, class: 'ti ' + parsed.value }
})
</script>

<template>
    <div class="bookmark-widget">
        <div v-if="!config" class="bookmark-empty">
            <i class="ti ti-bookmark-question" />
            <span>Configure bookmark in edit mode</span>
        </div>

        <a v-else :href="config.url" target="_blank" rel="noopener noreferrer" class="bookmark-link">
            <img
                v-if="iconInfo?.type === 'image'"
                :src="iconInfo.src"
                class="bookmark-icon-img"
            />
            <i v-else-if="iconInfo?.type === 'tabler'" :class="iconInfo.class" class="bookmark-icon" />
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
    background: color-mix(in srgb, currentColor 10%, transparent);
}

.bookmark-icon {
    font-size: 2rem;
    color: var(--p-primary-color);
    flex-shrink: 0;
}

.bookmark-icon-img {
    width: 2rem;
    height: 2rem;
    object-fit: contain;
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
