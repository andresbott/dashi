import { ref, computed, watch, onMounted, onBeforeUnmount, type Ref, type ComputedRef } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useGetDashboard, useUpdateDashboard } from '@/composables/useDashboards'
import { useAutosave, type AutosaveStatus } from '@/composables/useAutosave'
import { apiErrorMessage } from '@/lib/api/client'
import { plainCopy } from '@/lib/plainCopy'
import type { Dashboard } from '@/types/dashboard'

export interface UseDashboardDraftResult {
    // The editor's local working copy; null until the server dashboard arrives.
    draft: Ref<Dashboard | null>
    isLoading: Ref<boolean>
    isError: Ref<boolean>
    saveStatus: Ref<AutosaveStatus>
    isSaving: ComputedRef<boolean>
    // True while a change has not yet reached the server (debounce pending,
    // save in flight, or last save failed).
    hasUnsaved: ComputedRef<boolean>
    flush: () => Promise<void>
}

// Owns the editor's local working copy of a dashboard: seeds it as a deep copy
// of the server data, re-seeds it when the route id changes (so one dashboard's
// copy is never written to another), autosaves it, warns before the tab is
// closed with unsaved changes, and reports save failures via toast.
export function useDashboardDraft(id: () => string): UseDashboardDraftResult {
    const toast = useToast()
    const idRef = computed(id)

    const { data: serverDashboard, isLoading, isError } = useGetDashboard(id)
    const { updateDashboard } = useUpdateDashboard()

    const draft = ref<Dashboard | null>(null)
    const dirty = ref(false)

    const { status: saveStatus, error: saveError, isSaving, flush } = useAutosave<Dashboard>({
        source: () => draft.value,
        // The id travels with the draft, so a save never targets the route's
        // current id if the route moved on since the change was made.
        save: (d) => updateDashboard({ id: d.id, payload: d }),
    })

    // On an id change the old copy is persisted first and only then dropped:
    // flush() re-reads the source after any in-flight save, so the draft must
    // still hold the old dashboard at that point or a pending edit is lost.
    watch(idRef, async () => {
        await flush()
        draft.value = null
        dirty.value = false
    })

    // Seed when the server data for the *current* id is available and no copy
    // is held (initially, or once the previous id's copy has been dropped).
    watch([serverDashboard, draft], ([val]) => {
        if (val && !draft.value && val.id === idRef.value) {
            draft.value = plainCopy(val)
        }
    }, { immediate: true })

    // Sync so hasUnsaved is true the moment a change is made, mirroring the
    // autosave watcher. The initial population (null -> copy) is not a change.
    watch(draft, (val, oldVal) => {
        if (val == null || oldVal == null) return
        dirty.value = true
    }, { deep: true, flush: 'sync' })

    watch(saveStatus, (s) => {
        if (s === 'saved') dirty.value = false
        if (s === 'error') {
            // A 400 names the cause, e.g. which widget's config the server
            // rejected; anything else keeps the generic message.
            const reason = apiErrorMessage(saveError.value)
            const detail = reason ? `Failed to save dashboard: ${reason}` : 'Failed to save dashboard'
            toast.add({ severity: 'error', summary: 'Error', detail, life: 8000 })
        }
    })

    const hasUnsaved = computed(() => dirty.value || isSaving.value)

    const onBeforeUnload = (event: BeforeUnloadEvent) => {
        if (!hasUnsaved.value) return
        event.preventDefault()
        event.returnValue = true
    }
    onMounted(() => window.addEventListener('beforeunload', onBeforeUnload))
    onBeforeUnmount(() => window.removeEventListener('beforeunload', onBeforeUnload))

    return { draft, isLoading, isError, saveStatus, isSaving, hasUnsaved, flush }
}
