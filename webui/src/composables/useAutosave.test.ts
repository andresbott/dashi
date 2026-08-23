import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, defineComponent, type Ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { useAutosave, type UseAutosaveResult } from './useAutosave'

// Mounts a component whose setup wires useAutosave to a reactive source, and
// hands back the source ref (to mutate), the composable result, and the wrapper
// (to unmount). A component context is required so onBeforeUnmount fires.
function setupAutosave<T>(opts: {
    initial: T | null
    save: (v: T) => Promise<unknown>
    delay?: number
    enabled?: () => boolean
}) {
    const source = ref<T | null>(opts.initial) as Ref<T | null>
    let result!: UseAutosaveResult
    const wrapper = mount(
        defineComponent({
            setup() {
                result = useAutosave<T>({
                    source: () => source.value,
                    save: opts.save,
                    delay: opts.delay,
                    enabled: opts.enabled,
                })
                return () => null
            },
        }),
    )
    return { source, result, wrapper }
}

describe('useAutosave', () => {
    beforeEach(() => vi.useFakeTimers())
    afterEach(() => {
        vi.useRealTimers()
        vi.restoreAllMocks()
    })

    it('starts idle', () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { result } = setupAutosave({ initial: { n: 0 }, save })
        expect(result.status.value).toBe('idle')
        expect(result.isSaving.value).toBe(false)
    })

    it('does not save when the source is first populated', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source } = setupAutosave<{ name: string }>({ initial: null, save })

        // Simulates the server response being cloned into the local editing copy.
        source.value = { name: 'A' }
        await vi.advanceTimersByTimeAsync(5000)

        expect(save).not.toHaveBeenCalled()
    })

    it('saves once after a change, after the debounce delay', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source, result } = setupAutosave<{ name: string }>({
            initial: { name: 'A' },
            save,
            delay: 1000,
        })

        source.value = { name: 'B' }
        await vi.advanceTimersByTimeAsync(500)
        expect(save).not.toHaveBeenCalled()

        await vi.advanceTimersByTimeAsync(500)
        expect(save).toHaveBeenCalledTimes(1)
        expect(save).toHaveBeenCalledWith({ name: 'B' })
        expect(result.status.value).toBe('saved')
    })

    it('detects deep in-place mutations of the source', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source } = setupAutosave<{ page: { title: string } }>({
            initial: { page: { title: 'a' } },
            save,
            delay: 1000,
        })

        source.value!.page.title = 'b'
        await vi.advanceTimersByTimeAsync(1000)

        expect(save).toHaveBeenCalledTimes(1)
    })

    it('coalesces rapid successive changes into a single save', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 1000 })

        source.value = { n: 1 }
        await vi.advanceTimersByTimeAsync(400)
        source.value = { n: 2 }
        await vi.advanceTimersByTimeAsync(400)
        source.value = { n: 3 }
        await vi.advanceTimersByTimeAsync(1000)

        expect(save).toHaveBeenCalledTimes(1)
        expect(save).toHaveBeenCalledWith({ n: 3 })
    })

    it('reports saving while in flight, then saved', async () => {
        let resolve!: () => void
        const save = vi.fn().mockReturnValue(new Promise<void>((r) => (resolve = r)))
        const { source, result } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 1000 })

        source.value = { n: 1 }
        await vi.advanceTimersByTimeAsync(1000)
        expect(save).toHaveBeenCalledTimes(1)
        expect(result.status.value).toBe('saving')
        expect(result.isSaving.value).toBe(true)

        resolve()
        await flushPromises()
        expect(result.status.value).toBe('saved')
        expect(result.isSaving.value).toBe(false)
    })

    it('reports error when a save fails', async () => {
        const err = new Error('boom')
        const save = vi.fn().mockRejectedValue(err)
        const { source, result } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 1000 })

        source.value = { n: 1 }
        await vi.advanceTimersByTimeAsync(1000)
        await flushPromises()

        expect(result.status.value).toBe('error')
        expect(result.error.value).toBe(err)
    })

    it('runs a follow-up save for changes made while a save is in flight', async () => {
        const resolvers: Array<() => void> = []
        const save = vi.fn().mockImplementation(() => new Promise<void>((r) => resolvers.push(r)))
        const { source } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 1000 })

        source.value!.n = 1
        await vi.advanceTimersByTimeAsync(1000) // save #1 dispatched, still pending
        expect(save).toHaveBeenCalledTimes(1)

        source.value!.n = 2
        await vi.advanceTimersByTimeAsync(1000) // debounce fires but a save is in flight
        expect(save).toHaveBeenCalledTimes(1)

        resolvers[0]() // save #1 completes -> queued change is flushed
        await flushPromises()
        expect(save).toHaveBeenCalledTimes(2)
        expect(save).toHaveBeenLastCalledWith({ n: 2 })

        resolvers[1]()
        await flushPromises()
    })

    it('flush() saves pending changes immediately without waiting for the debounce', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source, result } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 5000 })

        source.value = { n: 1 }
        await result.flush()

        expect(save).toHaveBeenCalledTimes(1)
        expect(save).toHaveBeenCalledWith({ n: 1 })

        // The pending debounce must have been cancelled, so no second save fires.
        await vi.advanceTimersByTimeAsync(5000)
        expect(save).toHaveBeenCalledTimes(1)
    })

    it('saves pending changes when the component unmounts', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source, wrapper } = setupAutosave<{ n: number }>({ initial: { n: 0 }, save, delay: 5000 })

        source.value = { n: 1 }
        wrapper.unmount()
        await flushPromises()

        expect(save).toHaveBeenCalledTimes(1)
    })

    it('does not save while disabled', async () => {
        const save = vi.fn().mockResolvedValue(undefined)
        const { source } = setupAutosave<{ n: number }>({
            initial: { n: 0 },
            save,
            delay: 1000,
            enabled: () => false,
        })

        source.value = { n: 1 }
        await vi.advanceTimersByTimeAsync(2000)

        expect(save).not.toHaveBeenCalled()
    })
})
