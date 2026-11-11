<script setup lang="ts">
import Toast from 'primevue/toast'
</script>

<template>
    <Toast />
    <div class="app-layout">
        <main>
            <router-view />
        </main>
    </div>
</template>

<style lang="css">
.app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.app-topbar {
    display: flex;
    align-items: center;
    padding: 0 1rem;
    height: 3rem;
    background: var(--p-primary-color);
    color: var(--p-primary-contrast-color);
}

.app-topbar-icon {
    height: 1.75rem;
    width: 1.75rem;
    margin-right: 0.5rem;
    cursor: pointer;
}

/* Rendered as a <button> for keyboard access; reset the UA button styles so it
   reads as the plain topbar title. */
.app-topbar-title {
    background: none;
    border: 0;
    padding: 0;
    color: inherit;
    font: inherit;
    font-size: 1.25rem;
    font-weight: 700;
    cursor: pointer;
}

.app-layout main {
    flex: 1;
}
</style>
