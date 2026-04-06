<script setup lang="ts">
import { computed } from 'vue'
import { useXkcd } from '@/composables/useXkcd'
import type { XkcdWidgetConfig } from '@/types/xkcd'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<XkcdWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as XkcdWidgetConfig
})

const mode = computed(() => config.value?.mode ?? 'latest')

const { data: comic, isLoading, isError } = useXkcd(mode)

const comicUrl = computed(() => comic.value ? `https://xkcd.com/${comic.value.num}/` : '#')
</script>

<template>
    <div class="xkcd-widget">
        <div v-if="isLoading" class="xkcd-empty">
            <i class="ti ti-loader-2 xkcd-spinner" />
        </div>

        <div v-else-if="isError" class="xkcd-empty">
            <i class="ti ti-photo-off" />
            <span>Failed to load comic</span>
        </div>

        <div v-else-if="comic" class="xkcd-content">
            <div class="xkcd-header">
                <span class="xkcd-title">{{ comic.title }}</span>
                <span class="xkcd-number">#{{ comic.num }}</span>
            </div>
            <div class="xkcd-image-container">
                <a :href="comicUrl" target="_blank" rel="noopener">
                    <img :src="comic.img" :alt="comic.alt" :title="comic.alt" class="xkcd-image" />
                </a>
            </div>
        </div>
    </div>
</template>

<style scoped>
.xkcd-widget {
    padding: 0.5rem;
    min-height: 60px;
    height: 100%;
    box-sizing: border-box;
    overflow: hidden;
}

.xkcd-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.xkcd-empty .ti {
    font-size: 1.5rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.xkcd-spinner {
    animation: spin 1s linear infinite;
}

.xkcd-content {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.xkcd-header {
    display: flex;
    justify-content: center;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    flex-shrink: 0;
}

.xkcd-title {
    font-weight: 700;
    font-size: 1em;
}

.xkcd-number {
    font-size: 0.75em;
    color: var(--p-text-muted-color);
}

.xkcd-image-container {
    flex: 1;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
}

.xkcd-image-container a {
    display: flex;
    max-height: 100%;
    max-width: 100%;
}

.xkcd-image {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    border-radius: 8px;
    border: 8px solid white;
    background: white;
}
</style>
