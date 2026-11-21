<script setup lang="ts">
import { getWidgetsByCategory } from '@/lib/widgetRegistry'
import { allDocMocks, getWidgetDocs } from '@/lib/widgetDocs'
import { provideDocQueryClient } from '@/lib/docQueryClient'
import WidgetPreview from '@/components/docs/WidgetPreview.vue'

// Widget documentation index: one section per category, each a grid of cards
// showing the widget rendered from sample data. Each card opens the widget's
// own page (DocWidget.vue).
provideDocQueryClient(allDocMocks())

const groups = getWidgetsByCategory()
const previewConfig = (type: string) => getWidgetDocs(type)?.preview.config
</script>

<template>
    <section id="widgets">
        <h2>Widgets</h2>
        <p>Widgets are the building blocks of a dashboard. Each widget occupies one or more columns in a 12-column grid row. Widgets can fetch live data from external APIs or render purely client-side. Pick a widget to see it in action and read about its settings.</p>

        <section
            v-for="group in groups"
            :key="group.id"
            class="doc-widget-category"
            :aria-labelledby="`widgets-${group.id}`"
            :data-category="group.id"
        >
            <h3 :id="`widgets-${group.id}`">{{ group.label }}</h3>
            <ul class="doc-widget-grid">
                <li v-for="w in group.widgets" :key="w.type">
                    <router-link
                        :to="{ name: 'doc-widget', params: { type: w.type } }"
                        class="doc-widget-card"
                        :data-type="w.type"
                    >
                        <div class="doc-widget-card-stage" inert aria-hidden="true">
                            <WidgetPreview :type="w.type" :config="previewConfig(w.type)" bare />
                        </div>
                        <span class="doc-widget-card-name">
                            <i :class="`ti ${w.icon}`" aria-hidden="true" />
                            {{ w.label }}
                        </span>
                    </router-link>
                </li>
            </ul>
        </section>
    </section>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.doc-widget-category {
    margin-top: 2rem;
}

.doc-content .doc-widget-category h3 {
    margin: 0 0 0.75rem 0;
}

.doc-widget-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1rem;

    @include bp.below-lg {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    @include bp.below-sm {
        grid-template-columns: minmax(0, 1fr);
    }

    li {
        margin: 0;
    }
}

.doc-widget-card {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
    overflow: hidden;
    color: inherit;
    text-decoration: none;
    transition: border-color 0.15s, box-shadow 0.15s;

    &:hover {
        border-color: var(--p-primary-color);
        box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
    }

    &:focus-visible {
        outline: 2px solid var(--p-primary-color);
        outline-offset: 2px;
    }
}

/* A clipped, non-interactive rendering of the widget. Short widgets sit in
   the middle; tall ones show their top and fade out at the bottom. */
.doc-widget-card-stage {
    position: relative;
    height: 9rem;
    overflow: hidden;
    padding: 0.5rem 0.75rem;
    display: flex;
    flex-direction: column;
    justify-content: safe center;
    pointer-events: none;
    background: var(--p-surface-50);
    border-bottom: 1px solid var(--p-surface-200);

    &::after {
        content: '';
        position: absolute;
        inset: auto 0 0 0;
        height: 1.5rem;
        background: linear-gradient(transparent, var(--p-surface-50));
    }
}

.doc-widget-card-name {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.6rem 0.85rem;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--p-text-color);

    i {
        color: var(--p-primary-color);
        font-size: 1.1rem;
    }
}
</style>
