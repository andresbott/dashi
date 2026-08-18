<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()

const groups = [
    {
        label: 'Dashboards',
        items: [
            { key: 'admin-dashboards', label: 'Dashboards', icon: 'ti ti-layout-dashboard' },
        ],
    },
    {
        label: 'Theming',
        items: [
            { key: 'admin-backgrounds', label: 'Backgrounds', icon: 'ti ti-wallpaper' },
            { key: 'admin-themes',      label: 'Themes',      icon: 'ti ti-palette' },
        ],
    },
    {
        label: 'Content',
        items: [
            { key: 'admin-notes',  label: 'Notes',  icon: 'ti ti-notes' },
            { key: 'admin-images', label: 'Images', icon: 'ti ti-photo' },
        ],
    },
    {
        label: 'Documentation',
        items: [
            { key: 'doc-dashboards', label: 'Dashboards', icon: 'ti ti-layout-dashboard' },
            { key: 'doc-widgets',    label: 'Widgets',    icon: 'ti ti-apps' },
            { key: 'doc-theming',    label: 'Theming',    icon: 'ti ti-palette' },
        ],
    },
]

const activeKey = computed(() => (route.name as string) ?? '')
const isDocRoute = computed(() => activeKey.value.startsWith('doc-'))
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/admin')">Dashi</span>
    </header>
    <div class="admin-view">
        <div class="admin-layout">
            <nav class="admin-nav">
                <div
                    v-for="group in groups"
                    :key="group.label"
                    class="admin-nav-group"
                >
                    <div class="admin-nav-group-label">{{ group.label }}</div>
                    <div
                        v-for="item in group.items"
                        :key="item.key"
                        class="admin-nav-item"
                        :class="{ active: activeKey === item.key }"
                        @click="router.push({ name: item.key })"
                    >
                        <i :class="item.icon" />
                        <span>{{ item.label }}</span>
                    </div>
                </div>
            </nav>
            <div class="admin-panel" :class="{ 'doc-content': isDocRoute }">
                <router-view />
            </div>
        </div>
    </div>
</template>

<style scoped>
.admin-view {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}

.admin-layout {
    display: flex;
    gap: 2rem;
}

.admin-nav {
    position: sticky;
    top: 1rem;
    align-self: flex-start;
    min-width: 200px;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.admin-nav-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.admin-nav-group-label {
    padding: 0 0.75rem;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--p-text-muted-color);
    opacity: 0.7;
}

.admin-nav-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.95rem;
    color: var(--p-text-muted-color);
    transition: all 0.15s;
    white-space: nowrap;
}

.admin-nav-item:hover {
    background-color: var(--p-surface-100);
    color: var(--p-text-color);
}

.admin-nav-item.active {
    background-color: var(--p-primary-50);
    color: var(--p-primary-color);
    font-weight: 600;
}

.admin-nav-item i {
    font-size: 1.1rem;
    width: 1.25rem;
    text-align: center;
}

.admin-panel {
    flex: 1;
    min-width: 0;
}
</style>
