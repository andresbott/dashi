<script setup lang="ts">
import { computed, ref, watch, onUnmounted, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
import WeatherWidget from '@/components/dashboards/WeatherWidget.vue'
import WeatherCompactWidget from '@/components/dashboards/WeatherCompactWidget.vue'
import BookmarkWidget from '@/components/dashboards/BookmarkWidget.vue'
import ClockWidget from '@/components/dashboards/ClockWidget.vue'
import { useGetDashboard } from '@/composables/useDashboards'
import { useThemes } from '@/composables/useThemes'
import { getFontUrl } from '@/lib/api/themes'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)
const { data: dashboard, isLoading, isError } = useGetDashboard(() => id.value)

const { data: themesData } = useThemes()

const fontStyleEl = ref<HTMLStyleElement | null>(null)

const injectFonts = () => {
    if (fontStyleEl.value) {
        fontStyleEl.value.remove()
        fontStyleEl.value = null
    }

    if (!dashboard.value || !themesData.value) return

    const themeName = dashboard.value.theme || 'default'
    const themeInfo = themesData.value.find(t => t.name === themeName)
    if (!themeInfo || !themeInfo.fonts?.length) return

    const rules = themeInfo.fonts.map(f =>
        `@font-face { font-family: '${f.name}'; src: url('${getFontUrl(themeName, f.name)}') format('truetype'); }`
    ).join('\n')

    const style = document.createElement('style')
    style.textContent = rules
    document.head.appendChild(style)
    fontStyleEl.value = style
}

watch([() => dashboard.value, () => themesData.value], injectFonts, { immediate: true })

const dashboardTheme = computed(() => dashboard.value?.theme || 'default')
provide('dashboardTheme', dashboardTheme)

const themeFontFamily = computed(() => {
    if (!dashboard.value || !themesData.value) return undefined
    const themeName = dashboard.value.theme || 'default'
    const themeInfo = themesData.value.find(t => t.name === themeName)
    if (!themeInfo?.fonts?.length) return undefined
    return themeInfo.fonts[0].name
})

onUnmounted(() => {
    if (fontStyleEl.value) {
        fontStyleEl.value.remove()
    }
})

// Page navigation
const activePage = computed(() => {
    if (!dashboard.value) return 0
    const pageParam = route.query.page
    const pageIndex = pageParam ? parseInt(String(pageParam), 10) : 0
    const maxIndex = dashboard.value.pages.length - 1
    return Math.max(0, Math.min(pageIndex, maxIndex))
})

const showTabs = computed(() => {
    return dashboard.value ? dashboard.value.pages.length > 1 : false
})

const currentRows = computed(() => {
    if (!dashboard.value) return []
    return dashboard.value.pages[activePage.value].rows
})

function pageName(index: number) {
    if (!dashboard.value) return `Page ${index + 1}`
    const page = dashboard.value.pages[index]
    return page.name || `Page ${index + 1}`
}

function switchPage(index: number) {
    router.replace({ query: { ...route.query, page: String(index) } })
}
</script>

<template>
    <div v-if="isLoading" class="p-4">Loading...</div>
    <div v-else-if="isError" class="not-found">
        <i class="ti ti-error-404" />
        <h1>Page not found</h1>
        <p>The dashboard you're looking for doesn't exist.</p>
        <Button label="Go to Dashboards" icon="ti ti-arrow-left" @click="router.push({ name: 'dashboards' })" />
    </div>
    <div v-else-if="dashboard" class="dashboard-view" :class="{ 'show-boxes': dashboard.container.showBoxes }" :style="{ fontFamily: themeFontFamily }">
        <div v-if="showTabs" class="dashboard-tabs">
            <button
                v-for="(page, index) in dashboard.pages"
                :key="index"
                class="dashboard-tab"
                :class="{ active: activePage === index }"
                @click="switchPage(index)"
            >
                {{ pageName(index) }}
            </button>
        </div>
        <div
            class="dashboard-container"
            :style="{
                maxWidth: dashboard.container.maxWidth,
                marginLeft: dashboard.container.horizontalAlign === 'right' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
                marginRight: dashboard.container.horizontalAlign === 'left' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
            }"
        >
            <div
                class="dashboard-rows"
                :style="{
                    display: 'flex',
                    flexDirection: 'column',
                    minHeight: '80vh',
                    justifyContent: dashboard.container.verticalAlign === 'center' ? 'center' : dashboard.container.verticalAlign === 'bottom' ? 'flex-end' : 'flex-start',
                }"
            >
                <div
                    v-for="row in currentRows"
                    :key="row.id"
                    class="dashboard-row"
                    :style="{ height: row.height, maxWidth: row.width, margin: '0 auto', width: '100%' }"
                >
                    <h3 v-if="row.title" class="row-title">{{ row.title }}</h3>
                    <div class="grid">
                        <div
                            v-for="widget in row.widgets"
                            :key="widget.id"
                            :class="'col-' + widget.width"
                        >
                            <WeatherWidget v-if="widget.type === 'weather'" :widget="widget" />
                            <WeatherCompactWidget v-else-if="widget.type === 'weather-compact'" :widget="widget" />
                            <BookmarkWidget v-else-if="widget.type === 'bookmark'" :widget="widget" />
                            <ClockWidget v-else-if="widget.type === 'clock'" :widget="widget" />
                            <WidgetPlaceholder v-else :title="widget.title" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.dashboard-tabs {
    display: flex;
    gap: 0.5rem;
    padding: 1rem 1rem 0 1rem;
    border-bottom: 1px solid var(--p-surface-border);
}

.dashboard-tab {
    padding: 0.5rem 1rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    color: var(--p-text-muted-color);
    font-size: 0.95rem;
    transition: all 0.2s;
}

.dashboard-tab:hover {
    color: var(--p-text-color);
}

.dashboard-tab.active {
    color: var(--p-primary-color);
    border-bottom-color: var(--p-primary-color);
    font-weight: 500;
}

.dashboard-row {
    padding: 0.5rem;
}

.row-title {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    font-weight: 600;
}

.not-found {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    gap: 0.75rem;
    color: var(--p-text-muted-color);
}

.not-found .ti-error-404 {
    font-size: 4rem;
}

.not-found h1 {
    margin: 0;
    color: var(--p-text-color);
}

.show-boxes .dashboard-container {
    border: 2px dashed var(--p-primary-color);
}

.show-boxes .dashboard-row {
    border: 2px dashed var(--p-orange-500, #f59e0b);
}

.show-boxes .grid > div {
    border: 1px dotted var(--p-text-muted-color);
}
</style>
