<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, provide } from 'vue'
import { DASHBOARD_THEME, DASHBOARD_ID, ACTIVE_PAGE, TOTAL_PAGES } from '@/lib/injectionKeys'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import { useGetDashboard, useListDashboards } from '@/composables/useDashboards'
import { useThemes } from '@/composables/useThemes'
import { getFontUrl, getThemeBackgroundUrl } from '@/lib/api/themes'

const route = useRoute()
const router = useRouter()

// Resolve default dashboard when no :id param (i.e. the "/" route)
const { dashboards: dashboardList } = useListDashboards()
const resolvedDefaultId = computed(() => {
    if (route.params.id) return null
    const list = dashboardList.value
    if (!list?.length) return null
    const target = list.find(d => d.default) ?? list[0]
    return target.id
})

const id = computed(() => (route.params.id as string) || resolvedDefaultId.value || '')
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
provide(DASHBOARD_THEME, dashboardTheme)
provide(DASHBOARD_ID, id)

const themeFontFamily = computed(() => {
    if (!dashboard.value || !themesData.value) return undefined
    const themeName = dashboard.value.theme || 'default'
    const themeInfo = themesData.value.find(t => t.name === themeName)
    if (!themeInfo?.fonts?.length) return undefined
    return themeInfo.fonts[0].name
})

const colorMode = computed(() => dashboard.value?.colorMode || 'auto')

const colorSchemeStyle = computed(() => {
    switch (colorMode.value) {
        case 'light':
            return { colorScheme: 'light' }
        case 'dark':
            return { colorScheme: 'dark' }
        default:
            return { colorScheme: 'light dark' }
    }
})

const accentColorStyle = computed(() => {
    const color = dashboard.value?.accentColor
    if (!color) return {}
    return { '--p-primary-color': color }
})

const backgroundStyle = computed(() => {
    const bg = dashboard.value?.background
    if (!bg || bg.type === 'none' || !bg.value) return {}

    switch (bg.type) {
        case 'color':
            return { background: bg.value }
        case 'gradient':
            return { background: bg.value }
        case 'image': {
            let url: string
            if (bg.value.startsWith('theme:')) {
                // "theme:themename/filename.jpg"
                const rest = bg.value.slice(6)
                const slashIdx = rest.indexOf('/')
                const themeName = rest.slice(0, slashIdx)
                const fileName = rest.slice(slashIdx + 1)
                url = getThemeBackgroundUrl(themeName, fileName)
            } else if (bg.value.startsWith('dashboard:')) {
                // "dashboard:filename.jpg"
                // For preview dashboards (id ending in -prev), use the base dashboard's assets
                const assetDashId = id.value.endsWith('-prev') ? id.value.slice(0, -5) : id.value
                const fileName = bg.value.slice(10)
                url = `/api/v0/dashboards/${assetDashId}/assets/${encodeURIComponent(fileName)}`
            } else {
                return {}
            }
            return {
                background: `url('${url}') center/cover no-repeat`,
            }
        }
        default:
            return {}
    }
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

const totalPages = computed(() => dashboard.value ? dashboard.value.pages.length : 1)
provide(ACTIVE_PAGE, activePage)
provide(TOTAL_PAGES, totalPages)

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

const onKeyDown = (e: KeyboardEvent) => {
    if (!showTabs.value) return
    if (e.key === 'ArrowLeft' && activePage.value > 0) {
        switchPage(activePage.value - 1)
    } else if (e.key === 'ArrowRight' && activePage.value < totalPages.value - 1) {
        switchPage(activePage.value + 1)
    }
}

onMounted(() => window.addEventListener('keydown', onKeyDown))
onUnmounted(() => window.removeEventListener('keydown', onKeyDown))

const isImageDashboard = computed(() => dashboard.value?.type === 'image')

const imageUrl = computed(() => {
    if (!isImageDashboard.value) return ''
    const base = `/${id.value}`
    const params = new URLSearchParams()
    if (activePage.value > 0) {
        params.set('page', String(activePage.value))
    }
    // Cache-buster so the browser refetches on page change
    params.set('_t', String(Date.now()))
    return `${base}?${params.toString()}`
})

// Debug mode: ?debug=1 shows random background colors on containers, rows, widgets
const isDebug = computed(() => route.query.debug === '1')

const debugColors = ['#ffcccc', '#ccffcc', '#ccccff', '#ffffcc', '#ffccff', '#ccffff']

function debugColor(index: number): string | undefined {
    if (!isDebug.value) return undefined
    return debugColors[index % debugColors.length]
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
    <div v-else-if="dashboard && isImageDashboard" class="dashboard-image-view" :class="{ 'dark-mode': colorMode === 'dark' }" :style="colorSchemeStyle">
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
        <div class="dashboard-image-container">
            <img :src="imageUrl" :alt="dashboard.name" class="dashboard-rendered-image" />
        </div>
    </div>
    <div v-else-if="dashboard" class="dashboard-view" :class="{ 'show-boxes': dashboard.container.showBoxes || isDebug, 'dark-mode': colorMode === 'dark' }" :style="{ fontFamily: themeFontFamily, ...backgroundStyle, ...colorSchemeStyle, ...accentColorStyle }">
        <div
            class="dashboard-container"
            :style="{
                maxWidth: dashboard.container.maxWidth,
                marginLeft: dashboard.container.horizontalAlign === 'right' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
                marginRight: dashboard.container.horizontalAlign === 'left' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
            }"
        >
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
                            v-for="(widget, widgetIdx) in row.widgets"
                            :key="widget.id"
                            :class="'col-' + widget.width"
                            :style="{ background: debugColor(widgetIdx), overflow: 'hidden' }"
                        >
                            <component
                                v-if="getWidgetEntry(widget.type)"
                                :is="getWidgetEntry(widget.type)!.component"
                                v-bind="getWidgetEntry(widget.type)!.noWidgetProp ? {} : { widget }"
                            />
                            <WidgetPlaceholder v-else :title="widget.title" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.dashboard-image-view {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
}

.dashboard-image-container {
    display: flex;
    justify-content: center;
    padding: 1rem;
}

.dashboard-rendered-image {
    max-width: 100%;
    height: auto;
}

.dashboard-view {
    min-height: 100vh;
}

.dashboard-tabs {
    display: flex;
    gap: 0.5rem;
    padding: 1rem 1rem 0 1rem;
    border-bottom: none;
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

.dashboard-row[style*="height"] > .grid {
    height: 100%;
}

.dashboard-row[style*="height"] > .grid > div {
    height: 100%;
}

.row-title {
    margin: 0 0 1rem 0;
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

.dark-mode {
    --p-text-color: var(--p-surface-100);
    --p-text-muted-color: var(--p-surface-300);
    --p-surface-border: var(--p-surface-600);
    background-color: var(--p-surface-950);
    color: var(--p-text-color);
}
</style>
