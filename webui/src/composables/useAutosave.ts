import { ref, computed, watch, onBeforeUnmount, type Ref, type ComputedRef } from 'vue'

export type AutosaveStatus = 'idle' | 'saving' | 'saved' | 'error'

export interface UseAutosaveOptions<T> {
    // Getter for the reactive value to persist. Return null/undefined while the
    // value has not loaded yet — nothing is saved until it becomes non-null.
    source: () => T | null | undefined
    // Persists the value; typically wraps a mutation's mutateAsync.
    save: (value: T) => Promise<unknown>
    // Debounce window in ms between the last change and the save. Default 1000.
    delay?: number
    // Optional gate; when it returns false, changes are ignored.
    enabled?: () => boolean
}

export interface UseAutosaveResult {
    status: Ref<AutosaveStatus>
    isSaving: ComputedRef<boolean>
    error: Ref<unknown>
    // Cancels the pending debounce and persists immediately if there are
    // unsaved changes. Awaitable — used on navigation/unmount.
    flush: () => Promise<void>
}

// Persists a reactive value automatically a short debounce after it changes,
// so an editor needs no explicit Save button. Saves never overlap: a change
// made while a save is in flight is coalesced into a single follow-up save.
export function useAutosave<T>(options: UseAutosaveOptions<T>): UseAutosaveResult {
    const { source, save, enabled } = options
    const delay = options.delay ?? 1000

    const status = ref<AutosaveStatus>('idle')
    const error = ref<unknown>(null)
    const isSaving = computed(() => status.value === 'saving')

    let timer: ReturnType<typeof setTimeout> | null = null
    let dirty = false // unsaved changes are waiting to be persisted
    let running = false // a save loop is currently active

    const clearTimer = () => {
        if (timer !== null) {
            clearTimeout(timer)
            timer = null
        }
    }

    // Drains all pending changes. Re-reads the source on each iteration so a
    // change that lands mid-save is persisted by a follow-up pass rather than lost.
    const performSave = async () => {
        if (running) return
        running = true
        try {
            while (dirty) {
                dirty = false
                const value = source()
                if (value == null) break
                status.value = 'saving'
                error.value = null
                await save(value)
                status.value = 'saved'
            }
        } catch (e) {
            error.value = e
            status.value = 'error'
        } finally {
            running = false
        }
    }

    // Synchronous so a programmatic change schedules the save immediately and
    // flush() called right after a change sees the pending state.
    watch(
        source,
        (val, oldVal) => {
            if (val == null) return // nothing to persist yet
            if (oldVal == null) return // initial population (server -> local copy)
            if (enabled && !enabled()) return
            dirty = true
            clearTimer()
            timer = setTimeout(() => {
                timer = null
                void performSave()
            }, delay)
        },
        { deep: true, flush: 'sync' },
    )

    const flush = async () => {
        clearTimer()
        if (dirty) await performSave()
    }

    onBeforeUnmount(() => {
        void flush()
    })

    return { status, error, isSaving, flush }
}
