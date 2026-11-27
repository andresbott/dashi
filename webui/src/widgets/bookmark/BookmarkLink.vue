<script setup lang="ts">
// One bookmark link: icon plus title/subtitle. Shared by the bookmark and
// bookmarkgroup widgets; the caller decides what to show when unconfigured.
import { computed, inject, ref } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import { parseIcon, getSelfhstIconUrl, getDashboardIconUrl } from '@/lib/iconUtils'
import { safeBookmarkHref } from './url'

const props = defineProps<{
    url: string
    icon?: string
    title: string
    subtitle?: string
    textBelow?: boolean
    newTab?: boolean
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))

// Never bind a stored URL that fails the bookmark rule: the link renders
// without an href instead.
const href = computed(() => safeBookmarkHref(props.url))

const iconInfo = computed(() => {
    if (!props.icon) return null
    const parsed = parseIcon(props.icon)
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
    <a
        :href="href"
        :target="newTab ? '_blank' : undefined"
        :rel="newTab ? 'noopener noreferrer' : undefined"
        class="bookmark-link" :class="{ 'bookmark-link--below': textBelow }">
        <img
            v-if="iconInfo?.type === 'image'"
            :src="iconInfo.src"
            class="bookmark-icon-img"
        />
        <i v-else-if="iconInfo?.type === 'tabler'" :class="iconInfo.class" class="bookmark-icon" />
        <div class="bookmark-text">
            <span class="bookmark-title">{{ title }}</span>
            <span v-if="subtitle" class="bookmark-subtitle">{{ subtitle }}</span>
        </div>
    </a>
</template>

<style scoped>
.bookmark-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem;
    border-radius: 6px;
    text-decoration: none;
    color: inherit;
    transition: background 0.15s;
    min-width: 0;
}

.bookmark-link:hover {
    background: color-mix(in srgb, currentColor 10%, transparent);
}

.bookmark-link--below {
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    gap: 0.25rem;
    max-width: 6.5rem;
    aspect-ratio: 1;
    padding: 0.5rem;
    box-sizing: border-box;
    overflow: hidden;
    margin: 0 auto;
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
    max-width: 100%;
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
