<script setup lang="ts">
import { inject, ref } from 'vue'
import { ACTIVE_PAGE, TOTAL_PAGES } from '@/lib/injectionKeys'

const activePage = inject(ACTIVE_PAGE, ref(0))
const totalPages = inject(TOTAL_PAGES, ref(1))
</script>

<template>
    <div class="page-indicator">
        <span
            v-for="i in totalPages"
            :key="i"
            class="dot"
            :class="{ active: i - 1 === activePage }"
        />
    </div>
</template>

<style scoped>
.page-indicator {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem;
    min-height: 40px;
}

.dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 2px solid var(--p-primary-color);
    background: transparent;
    transition: background 0.2s;
}

.dot.active {
    background: var(--p-primary-color);
}
</style>
