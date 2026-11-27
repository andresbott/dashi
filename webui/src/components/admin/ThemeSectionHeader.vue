<script setup lang="ts">
// The title of a theme editor section Panel (Colours, Background), for its
// #header slot, plus a red alert icon while the section is invalid: a
// collapsed Panel still points to the invalid field it hides.
//
// titleId and titleClass are the slot's id and class: the Panel labels its
// content region by that id, so the title must carry it. Keep the Panel's
// header prop too: PrimeVue uses it as the toggle button's aria-label.
defineProps<{
    title: string
    invalid: boolean
    markerTestid: string
    titleId: string
    titleClass: string
}>()
</script>

<template>
    <span class="section-header">
        <span :id="titleId" :class="titleClass">{{ title }}</span>
        <i
            v-if="invalid"
            class="ti ti-alert-circle tab-invalid"
            role="img"
            :aria-label="`${title} has invalid values`"
            :data-testid="markerTestid"
        />
    </span>
</template>

<style scoped lang="scss">
/* Keeps the marker next to the title, not pushed to the middle of the
   header by its space-between layout. */
.section-header {
    display: inline-flex;
    align-items: center;
}

/* Same look as the tab marker in AdminThemeEdit.vue. */
.tab-invalid {
    margin-left: 0.35rem;
    color: var(--p-red-500);
}
</style>
