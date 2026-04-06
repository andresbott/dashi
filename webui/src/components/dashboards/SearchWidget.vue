<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Widget } from '@/types/dashboard'
import type { SearchWidgetConfig } from '@/types/search'
import { searchEngineUrls } from '@/types/search'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<SearchWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as SearchWidgetConfig
})

const query = ref('')

const engineLabel = computed(() => {
    switch (config.value?.engine) {
        case 'duckduckgo': return 'DuckDuckGo'
        case 'bing': return 'Bing'
        default: return 'Google'
    }
})

const placeholder = computed(() => config.value?.placeholder || 'Search...')

const doSearch = () => {
    if (!query.value.trim()) return
    const engine = config.value?.engine ?? 'google'
    const baseUrl = searchEngineUrls[engine]
    window.open(baseUrl + encodeURIComponent(query.value.trim()), '_blank')
}
</script>

<template>
    <div class="search-widget">
        <form class="search-form" @submit.prevent="doSearch">
            <i class="ti ti-search search-icon" />
            <input
                v-model="query"
                type="text"
                :placeholder="placeholder"
                class="search-input"
            />
            <button type="submit" class="search-button">
                <i class="ti ti-arrow-right" />
            </button>
        </form>
        <span class="search-engine-label">{{ engineLabel }}</span>
    </div>
</template>

<style scoped>
.search-widget {
    padding: 0.5rem;
    min-height: 60px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 0.25rem;
}

.search-form {
    display: flex;
    align-items: center;
    background: color-mix(in srgb, currentColor 10%, transparent);
    border: 1px solid color-mix(in srgb, currentColor 20%, transparent);
    border-radius: 8px;
    overflow: hidden;
    transition: border-color 0.15s;
}

.search-form:focus-within {
    border-color: var(--p-primary-color);
}

.search-icon {
    padding: 0 0.5rem 0 0.75rem;
    color: var(--p-text-muted-color);
    font-size: 1.1rem;
    flex-shrink: 0;
}

.search-input {
    flex: 1;
    border: none;
    background: transparent;
    padding: 0.6rem 0.5rem;
    font-size: 0.95rem;
    color: var(--p-text-color, inherit);
    font-family: inherit;
    outline: none;
}

.search-input::placeholder {
    color: var(--p-text-muted-color);
}

.search-button {
    border: none;
    background: var(--p-primary-color);
    color: var(--p-primary-contrast-color);
    padding: 0 0.75rem;
    cursor: pointer;
    font-size: 1rem;
    display: flex;
    align-items: center;
    align-self: stretch;
    transition: background 0.15s;
}

.search-button:hover {
    opacity: 0.9;
}

.search-engine-label {
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
    text-align: right;
    padding-right: 0.25rem;
}
</style>
