<script setup lang="ts">
// A 3x3 position grid that sets both container alignment axes at once. Each
// cell is a toggle button drawn as a tiny diagram of where the container
// lands. One tab stop (the current cell); arrow keys move focus between
// cells and Enter/Space picks one, as for any button.
import { computed } from 'vue'
import type { Container } from '@/types/dashboard'

type Vertical = Container['verticalAlign']
type Horizontal = Container['horizontalAlign']

const vertical = defineModel<Vertical>('vertical')
const horizontal = defineModel<Horizontal>('horizontal')

const VERTICAL: Vertical[] = ['top', 'center', 'bottom']
const HORIZONTAL: Horizontal[] = ['left', 'center', 'right']
const FLEX = ['flex-start', 'center', 'flex-end']
const capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1)

// Missing or unknown values (hand-edited dashboard.json) render like the
// server does (dashboard.PlaceContainer): at the top, centred.
const current = computed(() => ({
    v: vertical.value && VERTICAL.includes(vertical.value) ? vertical.value : 'top',
    h: horizontal.value && HORIZONTAL.includes(horizontal.value) ? horizontal.value : 'center',
}))

const cells = VERTICAL.flatMap((v, row) =>
    HORIZONTAL.map((h, col) => ({
        v,
        h,
        label: v === 'center' && h === 'center' ? 'Center' : `${capitalize(v)} ${h}`,
        style: { alignItems: FLEX[row], justifyContent: FLEX[col] },
    })),
)

const currentIndex = computed(() => cells.findIndex(c => c.v === current.value.v && c.h === current.value.h))

const pick = (cell: (typeof cells)[number]) => {
    vertical.value = cell.v
    horizontal.value = cell.h
}

// Indexed by cell; a ref array filled by v-for does not guarantee order.
const buttons: (HTMLButtonElement | null)[] = []
const setButton = (index: number) => (el: unknown) => {
    buttons[index] = el as HTMLButtonElement | null
}

const MOVES: Record<string, [number, number]> = {
    ArrowUp: [-1, 0],
    ArrowDown: [1, 0],
    ArrowLeft: [0, -1],
    ArrowRight: [0, 1],
}

const onKeydown = (event: KeyboardEvent, index: number) => {
    const move = MOVES[event.key]
    if (!move) return
    event.preventDefault()
    const row = Math.min(2, Math.max(0, Math.floor(index / 3) + move[0]))
    const col = Math.min(2, Math.max(0, (index % 3) + move[1]))
    buttons[row * 3 + col]?.focus()
}
</script>

<template>
    <div class="alignment-picker">
        <div role="group" aria-label="Container alignment" class="alignment-grid">
            <button
                v-for="(cell, index) in cells"
                :key="cell.label"
                :ref="setButton(index)"
                type="button"
                class="alignment-option"
                :class="{ selected: index === currentIndex }"
                :aria-label="cell.label"
                :aria-pressed="index === currentIndex ? 'true' : 'false'"
                :tabindex="index === currentIndex ? 0 : -1"
                :title="cell.label"
                data-testid="alignment-option"
                @click="pick(cell)"
                @keydown="onKeydown($event, index)"
            >
                <span class="alignment-box" :style="cell.style" aria-hidden="true">
                    <span class="alignment-content" />
                </span>
            </button>
        </div>
        <span class="alignment-current" aria-hidden="true">{{ cells[currentIndex]?.label }}</span>
    </div>
</template>

<style scoped>
.alignment-picker {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.alignment-grid {
    display: grid;
    grid-template-columns: repeat(3, auto);
    gap: 0.25rem;
}

.alignment-option {
    display: flex;
    padding: 0.25rem;
    border: 1px solid var(--p-content-border-color);
    border-radius: 6px;
    background: var(--p-content-background);
    cursor: pointer;
    transition: border-color 0.15s, background-color 0.15s;
}

.alignment-option:hover {
    border-color: var(--p-primary-color);
}

.alignment-option:focus-visible {
    outline: 2px solid var(--p-primary-color);
    outline-offset: 2px;
}

.alignment-option.selected {
    border-color: var(--p-primary-color);
    background: var(--p-highlight-background);
}

/* The viewport: a small frame with the container drawn in its position. */
.alignment-box {
    display: flex;
    width: 2.25rem;
    height: 1.5rem;
    padding: 2px;
    border: 1px solid var(--p-text-muted-color);
    border-radius: 2px;
}

.alignment-content {
    width: 45%;
    height: 40%;
    border-radius: 1px;
    background: var(--p-text-muted-color);
}

.alignment-option.selected .alignment-content {
    background: var(--p-primary-color);
}

.alignment-option.selected .alignment-box {
    border-color: var(--p-primary-color);
}

.alignment-current {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
}
</style>
