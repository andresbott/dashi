<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()

// Sections are grouped by what they hold: Content is user data that widgets
// display (notes, images), Theming is how dashboards look. Background images
// have no section: each theme holds its own, uploaded in the theme editor.
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
            { key: 'admin-themes',        label: 'Themes',        icon: 'ti ti-palette' },
            { key: 'admin-widget-themes', label: 'Widget themes', icon: 'ti ti-typography' },
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

// Child routes that have no nav entry of their own highlight their parent, so
// the sidebar does not go blank while you are editing.
const navParents: Record<string, string> = {
    'admin-theme-edit': 'admin-themes',
    'doc-widget': 'doc-widgets',
}

const activeKey = computed(() => {
    const name = (route.name as string) ?? ''
    return navParents[name] ?? name
})
const isDocRoute = computed(() => activeKey.value.startsWith('doc-'))

// Below the md breakpoint the sidebar is an off-canvas drawer opened from the
// topbar's menu button (CSS decides when; on wider screens navOpen has no
// visible effect). Any navigation closes it, including links inside a panel.
const navOpen = ref(false)
const toggleNav = () => {
    navOpen.value = !navOpen.value
}
const closeNav = () => {
    navOpen.value = false
}
const navigate = (name: string) => {
    closeNav()
    router.push({ name })
}
watch(() => route.fullPath, closeNav)

const onKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && navOpen.value) closeNav()
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
    <header class="app-topbar">
        <button
            type="button"
            class="admin-nav-toggle"
            :aria-expanded="navOpen"
            aria-controls="admin-nav"
            :aria-label="navOpen ? 'Close menu' : 'Open menu'"
            @click="toggleNav"
        >
            <i :class="navOpen ? 'ti ti-x' : 'ti ti-menu-2'" />
        </button>
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <button type="button" class="app-topbar-title" @click="navigate('admin-dashboards')">Dashi</button>
    </header>
    <div class="admin-view">
        <div class="admin-layout">
            <div v-if="navOpen" class="admin-nav-backdrop" @click="closeNav" />
            <nav id="admin-nav" class="admin-nav" :class="{ open: navOpen }">
                <div
                    v-for="group in groups"
                    :key="group.label"
                    class="admin-nav-group"
                >
                    <div class="admin-nav-group-label">{{ group.label }}</div>
                    <button
                        v-for="item in group.items"
                        :key="item.key"
                        type="button"
                        class="admin-nav-item"
                        :class="{ active: activeKey === item.key }"
                        :aria-current="activeKey === item.key ? 'page' : undefined"
                        @click="navigate(item.key)"
                    >
                        <i :class="item.icon" />
                        <span>{{ item.label }}</span>
                    </button>
                </div>
            </nav>
            <div class="admin-panel" :class="{ 'doc-content': isDocRoute }">
                <router-view />
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.admin-view {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem 1rem;

    @include bp.below-md {
        padding: 1rem 0.75rem;
    }
}

.admin-layout {
    display: flex;
    gap: 2rem;
}

/* Only shown below md, where the sidebar is a drawer. */
.admin-nav-toggle {
    display: none;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    margin: 0 0.25rem 0 -0.5rem;
    background: none;
    border: 0;
    border-radius: 6px;
    color: inherit;
    font-size: 1.4rem;
    cursor: pointer;

    &:focus-visible {
        outline: 2px solid currentColor;
        outline-offset: -2px;
    }

    @include bp.below-md {
        display: inline-flex;
    }
}

.admin-nav-backdrop {
    display: none;

    @include bp.below-md {
        display: block;
        position: fixed;
        inset: 3rem 0 0 0;
        z-index: 1000;
        background: rgba(0, 0, 0, 0.4);
    }
}

.admin-nav {
    position: sticky;
    top: 1rem;
    align-self: flex-start;
    min-width: 200px;
    display: flex;
    flex-direction: column;
    gap: 1rem;

    /* Off-canvas drawer under the topbar (3rem high, see App.vue). Hidden
       with visibility as well so its links leave the tab order when shut. */
    @include bp.below-md {
        position: fixed;
        top: 3rem;
        bottom: 0;
        left: 0;
        /* flex-start would shrink a fixed flex child to its content height */
        align-self: stretch;
        z-index: 1001;
        width: min(280px, 85vw);
        min-width: 0;
        padding: 1rem 0.75rem;
        overflow-y: auto;
        background: var(--p-content-background);
        box-shadow: 2px 0 12px rgba(0, 0, 0, 0.15);
        transform: translateX(-100%);
        visibility: hidden;
        transition: transform 0.2s ease, visibility 0.2s;

        &.open {
            transform: none;
            visibility: visible;
        }
    }
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
    font: inherit;
    font-size: 0.95rem;
    text-align: left;
    width: 100%;
    background: none;
    border: 0;
    color: var(--p-text-muted-color);
    transition: all 0.15s;
    white-space: nowrap;

    &:focus-visible {
        outline: 2px solid var(--p-primary-color);
        outline-offset: -2px;
    }
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
