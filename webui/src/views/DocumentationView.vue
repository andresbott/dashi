<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import dashiIcon from '@/assets/icon-64.png'

const router = useRouter()
const route = useRoute()

interface Section {
    id: string
    label: string
    route: string
    children?: { id: string; label: string }[]
}

const sections: Section[] = [
    {
        id: 'dashboards', label: 'Dashboards', route: 'doc-dashboards', children: [
            { id: 'dash-types', label: 'Types' },
            { id: 'dash-properties', label: 'Properties' },
            { id: 'dash-pages', label: 'Pages' },
            { id: 'dash-query-params', label: 'Query Parameters' },
            { id: 'dash-id', label: 'Dashboard ID' },
        ]
    },
    {
        id: 'widgets', label: 'Widgets', route: 'doc-widgets', children: [
            { id: 'widget-clock', label: 'Clock' },
            { id: 'widget-weather', label: 'Weather' },
            { id: 'widget-weather-compact', label: 'Weather Compact' },
            { id: 'widget-bookmark', label: 'Bookmark' },
            { id: 'widget-search', label: 'Search' },
            { id: 'widget-market', label: 'Market' },
            { id: 'widget-xkcd', label: 'XKCD' },
            { id: 'widget-departures', label: 'SBB Departures' },
            { id: 'widget-battery', label: 'Battery' },
            { id: 'widget-page-indicator', label: 'Page Indicator' },
            { id: 'widget-stack', label: 'Stack' },
        ]
    },
    {
        id: 'theming', label: 'Theming', route: 'doc-theming', children: [
            { id: 'theme-themes', label: 'Themes' },
            { id: 'theme-assets', label: 'Dashboard Assets' },
            { id: 'theme-custom-css', label: 'Custom CSS' },
            { id: 'theme-backgrounds', label: 'Backgrounds' },
        ]
    },
]

const activeRoute = computed(() => route.name as string)

function navigateTo(routeName: string) {
    router.push({ name: routeName })
}

function scrollTo(routeName: string, id: string) {
    if (activeRoute.value !== routeName) {
        router.push({ name: routeName }).then(() => {
            setTimeout(() => {
                const el = document.getElementById(id)
                if (el) el.scrollIntoView({ behavior: 'smooth' })
            }, 50)
        })
    } else {
        const el = document.getElementById(id)
        if (el) el.scrollIntoView({ behavior: 'smooth' })
    }
}
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/dashboards')">Dashi</span>
    </header>
    <div class="documentation-view">
        <div class="flex align-items-center justify-content-between mb-4">
            <h1 class="text-2xl font-bold text-color">Documentation</h1>
            <Button
                label="Back to Dashboards"
                icon="ti ti-arrow-left"
                severity="secondary"
                @click="router.push({ name: 'dashboards' })"
            />
        </div>

        <div class="doc-layout">
            <nav class="doc-sidebar">
                <template v-for="s in sections" :key="s.id">
                    <div class="doc-sidebar-group">
                        <div
                            class="doc-sidebar-heading"
                            :class="{ active: activeRoute === s.route }"
                            @click="navigateTo(s.route)"
                        >
                            {{ s.label }}
                        </div>
                        <ul v-if="s.children && activeRoute === s.route">
                            <li
                                v-for="c in s.children"
                                :key="c.id"
                                @click="scrollTo(s.route, c.id)"
                            >
                                {{ c.label }}
                            </li>
                        </ul>
                    </div>
                </template>
            </nav>

            <main class="doc-content">
                <router-view />
            </main>
        </div>
    </div>
</template>

<style>
.documentation-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}

.documentation-view .doc-layout {
    display: flex;
    gap: 2rem;
}

/* ---- Sidebar ---- */

.documentation-view .doc-sidebar {
    position: sticky;
    top: 1rem;
    align-self: flex-start;
    min-width: 190px;
    flex-shrink: 0;
}

.documentation-view .doc-sidebar-group {
    margin-bottom: 0.75rem;
}

.documentation-view .doc-sidebar-heading {
    padding: 0.45rem 0.75rem;
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--p-surface-800, #1f2937);
    cursor: pointer;
    border-radius: 6px;
    transition: background 0.15s, color 0.15s;
}

.documentation-view .doc-sidebar-heading:hover {
    background: var(--surface-hover);
}

.documentation-view .doc-sidebar-heading.active {
    background: var(--primary-color);
    color: var(--primary-color-text);
}

.documentation-view .doc-sidebar ul {
    list-style: none;
    padding: 0;
    margin: 0.15rem 0 0 0;
}

.documentation-view .doc-sidebar li {
    padding: 0.3rem 0.75rem 0.3rem 1rem;
    font-size: 0.82rem;
    font-weight: 400;
    cursor: pointer;
    border-radius: 6px;
    color: var(--text-color-secondary);
    border-left: 2px solid transparent;
    transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.documentation-view .doc-sidebar li:hover {
    background: var(--surface-hover);
    color: var(--text-color);
}

.documentation-view .doc-sidebar li.active {
    color: var(--primary-color);
    border-left-color: var(--primary-color);
    font-weight: 500;
    background: none;
}

/* ---- Content ---- */

.documentation-view .doc-content {
    flex: 1;
    min-width: 0;
    color: var(--text-color);
}

.documentation-view .doc-content h2 {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0 0 1rem 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--surface-border);
}

.documentation-view .doc-content h3 {
    font-size: 1rem;
    font-weight: 600;
    margin: 1.5rem 0 0.4rem 0;
    color: var(--text-color);
}

.documentation-view .doc-content h4 {
    font-size: 0.9rem;
    font-weight: 600;
    margin: 1.25rem 0 0.3rem 0;
    color: var(--text-color);
}

.documentation-view .doc-content p {
    margin: 0 0 0.5rem 0;
    line-height: 1.5;
    font-size: 0.875rem;
    color: var(--text-color-secondary);
}

.documentation-view .doc-content ul {
    margin: 0 0 0.5rem 0;
    padding-left: 1.25rem;
}

.documentation-view .doc-content li {
    margin-bottom: 0.15rem;
    line-height: 1.5;
    font-size: 0.875rem;
    color: var(--text-color-secondary);
}

.documentation-view .doc-content code {
    background: var(--surface-ground);
    padding: 0.1rem 0.35rem;
    border-radius: 4px;
    font-size: 0.8rem;
    color: var(--primary-color);
}

.documentation-view .doc-code {
    background: var(--p-surface-100, #f3f4f6);
    border: 1px solid var(--p-surface-200, #e5e7eb);
    border-radius: 6px;
    padding: 0.6rem 0.85rem;
    font-size: 0.85rem;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    color: var(--p-surface-800, #1f2937);
    margin-bottom: 0.75rem;
    overflow-x: auto;
}

.documentation-view .doc-hint {
    font-size: 0.8rem;
    font-style: italic;
    opacity: 0.7;
}

/* ---- Tables ---- */

.documentation-view .doc-content table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    margin-bottom: 1rem;
    font-size: 0.85rem;
    border: 1px solid var(--p-surface-300, #d1d5db);
    border-radius: 8px;
    overflow: hidden;
}

.documentation-view .doc-content thead tr {
    background: var(--p-surface-100, #f3f4f6);
}

.documentation-view .doc-content th {
    text-align: left;
    padding: 0.6rem 0.85rem;
    font-weight: 700;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--p-surface-600, #6b7280);
    border-bottom: 2px solid var(--p-surface-300, #d1d5db);
}

.documentation-view .doc-content td {
    padding: 0.55rem 0.85rem;
    color: var(--p-surface-600, #6b7280);
    border-bottom: 1px solid var(--p-surface-200, #e5e7eb);
    line-height: 1.5;
}

.documentation-view .doc-content tbody tr:last-child td {
    border-bottom: none;
}

.documentation-view .doc-content tbody tr:nth-child(even) {
    background: var(--p-surface-50, #f9fafb);
}

.documentation-view .doc-content tbody tr:hover {
    background: var(--p-surface-100, #f3f4f6);
}

.documentation-view .doc-content td:first-child {
    font-weight: 500;
    color: var(--p-surface-800, #1f2937);
}

.documentation-view .doc-content td code {
    white-space: nowrap;
}

/* ---- Dividers ---- */

.documentation-view .doc-divider {
    border: none;
    border-top: 1px solid var(--p-surface-200, #e5e7eb);
    margin: 1.5rem 0 0.5rem 0;
}

/* ---- Sections ---- */

.documentation-view section + section {
    margin-top: 3rem;
}
</style>
