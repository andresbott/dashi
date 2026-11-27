<script setup lang="ts">
// The theme editor's Preview section for ONE colour mode: a sample card and
// the accent colours over the mode's page background. The style comes from
// previewStyle, the same logic the server uses, so it shows what the
// dashboard will render without a save or a round-trip.
import { computed } from 'vue'

import { previewStyle } from '@/lib/themeCss'
import type { PaletteRole, Theme } from '@/types/theme'

const props = defineProps<{
    theme: Theme
    mode: 'light' | 'dark'
}>()

const style = computed(() => previewStyle(props.theme, props.mode))

const chips: { role: PaletteRole; label: string }[] = [
    { role: 'primary', label: 'Primary' },
    { role: 'secondary', label: 'Secondary' },
    { role: 'success', label: 'Success' },
    { role: 'warning', label: 'Warning' },
    { role: 'danger', label: 'Danger' },
]
</script>

<template>
    <div class="preview-box" :style="style" :data-testid="`preview-${mode}`">
        <div class="preview-card">
            <div class="preview-title">Card title</div>
            <div class="preview-muted">Muted secondary text</div>
            <div class="preview-chips">
                <span
                    v-for="chip in chips"
                    :key="chip.role"
                    class="preview-chip"
                    :class="{ 'preview-chip-primary': chip.role === 'primary' }"
                    :style="{ background: `var(--dashi-${chip.role})` }"
                >
                    {{ chip.label }}
                </span>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.preview-box {
    height: 220px;
    border-radius: 6px;
    border: 1px solid var(--p-content-border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
}

/* A dashboard card is not page-wide: keep the sample at card size. */
.preview-card {
    width: 100%;
    max-width: 320px;
    background: var(--dashi-surface);
    border: 1px solid var(--dashi-border);
    border-radius: 6px;
    padding: 0.75rem;
    color: var(--dashi-text);
}

.preview-title {
    font-weight: 600;
}

.preview-muted {
    color: var(--dashi-text-muted);
    font-size: 0.85rem;
    margin-bottom: 0.5rem;
}

.preview-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
}

.preview-chip {
    font-size: 0.7rem;
    padding: 0.1rem 0.4rem;
    border-radius: 999px;
    color: #ffffff;
}

.preview-chip-primary {
    color: var(--dashi-on-primary);
}
</style>
