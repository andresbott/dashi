import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import PrimeVue from 'primevue/config'
import ColorPicker from 'primevue/colorpicker'
import ThemeBackgroundEditor from './ThemeBackgroundEditor.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { ModeBackground } from '@/types/theme'

const assets = ref<string[] | undefined>([])
const uploadAsset = vi.fn()
const toastAdd = vi.fn()

vi.mock('@/composables/useThemes', () => ({
    useThemeAssets: () => ({ assets, uploadAsset, isUploadingAsset: ref(false) }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd }) }))

interface MountOptions {
    mode?: 'light' | 'dark'
    disabled?: boolean
    flush?: () => Promise<void>
    onUploading?: (v: boolean) => void
}

function mountEditor(background: ModeBackground, opts: MountOptions = {}) {
    const bg = reactive(background)
    const wrapper = mount(ThemeBackgroundEditor, {
        props: {
            background: bg,
            themeId: 'abc123',
            mode: opts.mode ?? 'light',
            disabled: opts.disabled ?? false,
            flush: opts.flush ?? (async () => {}),
            onUploading: opts.onUploading,
        },
        global: { plugins: [PrimeVue], stubs: { RouterLink: true } },
    })
    return { wrapper, bg }
}

// jsdom's File has no arrayBuffer(); the component only needs name + bytes.
const fakeFile = (name: string) => ({ name, arrayBuffer: async () => new ArrayBuffer(1) })

async function chooseFile(wrapper: ReturnType<typeof mountEditor>['wrapper'], name: string) {
    const input = wrapper.find<HTMLInputElement>('[data-testid="image-upload-input"]')
    Object.defineProperty(input.element, 'files', { value: [fakeFile(name)], configurable: true })
    await input.trigger('change')
    await flushPromises()
}

const gradient = (): ModeBackground => ({
    gradient: { direction: '90deg', stops: ['#000000', '#ff0000', '#ffffff'] },
})

const image = (ref: string): ModeBackground => ({
    image: { ref, fit: 'contain', position: 'top', repeat: 'repeat' },
})

// 2026-11-27T00:00:00Z, so uploaded names carry a known stamp.
const NOW = Date.UTC(2026, 10, 27)
const STAMP = NOW.toString(36)

describe('ThemeBackgroundEditor', () => {
    beforeEach(() => {
        // PrimeVue Select listens for orientation changes; jsdom has no matchMedia.
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        vi.useFakeTimers({ toFake: ['Date'] })
        vi.setSystemTime(NOW)
        vi.clearAllMocks()
        assets.value = ['own.png', 'sky.jpg']
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    describe('gradient', () => {
        it('switching the base to a gradient adds a default two-stop gradient, and back removes it', async () => {
            const { wrapper, bg } = mountEditor({})
            const buttons = wrapper.findAll('[data-testid="bg-base"] button')
            await buttons[1].trigger('click')
            expect(bg.gradient?.stops).toHaveLength(2)
            expect(bg.gradient?.direction).toBe('135deg')

            await buttons[0].trigger('click')
            expect(bg.gradient).toBeUndefined()
        })

        it('scopes the stop controls to its mode', () => {
            const { wrapper } = mountEditor(gradient(), { mode: 'dark' })
            expect(wrapper.find('[data-testid="stop-dark-0"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="stop-dark-add"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="stop-light-0"]').exists()).toBe(false)
        })

        it('has no dark toggles of its own', () => {
            const { wrapper } = mountEditor({ ...gradient(), ...image('asset:own.png') })
            expect(wrapper.find('[data-testid="gradient-dark-toggle"]').exists()).toBe(false)
            expect(wrapper.find('[data-testid="image-dark-toggle"]').exists()).toBe(false)
        })

        it('adds a stop repeating the last colour', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-add"]').trigger('click')
            expect(bg.gradient?.stops).toEqual(['#000000', '#ff0000', '#ffffff', '#ffffff'])
        })

        it('removes a stop but never below two', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-1-remove"]').trigger('click')
            expect(bg.gradient?.stops).toEqual(['#000000', '#ffffff'])
            expect(wrapper.find('[data-testid="stop-light-0-remove"]').attributes('disabled')).toBeDefined()
        })

        it('reorders stops', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-0-down"]').trigger('click')
            expect(bg.gradient?.stops).toEqual(['#ff0000', '#000000', '#ffffff'])
            await wrapper.find('[data-testid="stop-light-2-up"]').trigger('click')
            expect(bg.gradient?.stops).toEqual(['#ff0000', '#ffffff', '#000000'])
            expect(wrapper.find('[data-testid="stop-light-0-up"]').attributes('disabled')).toBeDefined()
        })

        it('edits a stop as text and flags an invalid one', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            const input = wrapper.find('[data-testid="stop-light-1"]')
            await input.setValue('#12')
            expect(bg.gradient?.stops[1]).toBe('#12')
            expect(input.classes()).toContain('p-invalid')
        })

        it('keeps a stop alpha when the picker writes it', async () => {
            const { wrapper, bg } = mountEditor({
                gradient: { direction: '90deg', stops: ['#00000080', '#ffffff'] },
            })
            wrapper.findAllComponents(ColorPicker)[0].vm.$emit('update:modelValue', '112233')
            await flushPromises()
            expect(bg.gradient?.stops[0]).toBe('#11223380')
        })

        it('shows a keyword direction as its angle without rewriting it', () => {
            const { wrapper, bg } = mountEditor({
                gradient: { direction: 'to bottom right', stops: ['#000000', '#ffffff'] },
            })
            expect(wrapper.find('[data-testid="bg-angle"]').text()).toBe('135°')
            expect(bg.gradient?.direction).toBe('to bottom right')
        })
    })

    describe('image', () => {
        it('shows the current image as a thumbnail with Replace and Remove', () => {
            const { wrapper } = mountEditor(image('asset:sky.jpg'), { mode: 'dark' })
            const thumb = wrapper.find('[data-testid="image-dark-thumb"]')
            expect(thumb.attributes('src')).toContain('/api/v0/themes/abc123/assets/sky.jpg')
            expect(wrapper.find('[data-testid="image-dark-upload"]').text()).toContain('Replace')
            expect(wrapper.find('[data-testid="image-dark-remove"]').exists()).toBe(true)
        })

        it('offers only Upload when there is no image', () => {
            const { wrapper } = mountEditor({})
            expect(wrapper.find('[data-testid="image-light-thumb"]').exists()).toBe(false)
            expect(wrapper.find('[data-testid="image-light-upload"]').text()).toContain('Upload')
            expect(wrapper.find('[data-testid="image-light-remove"]').exists()).toBe(false)
        })

        it('uploads under a unique name and sets it with default settings', async () => {
            uploadAsset.mockResolvedValue(undefined)
            const { wrapper, bg } = mountEditor({})
            await chooseFile(wrapper, 'new.webp')

            const name = `new-${STAMP}.webp`
            expect(uploadAsset).toHaveBeenCalledWith(expect.objectContaining({ name }))
            expect(bg.image).toEqual({ ref: `asset:${name}`, fit: 'cover', position: 'center', repeat: 'no-repeat' })
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'success' }))
        })

        // The parent's autosave is gated while uploading: it must be open again
        // when the ref lands, or the edit that references the file is dropped.
        it('flushes, blocks autosave, uploads, reopens it, sets the ref, then flushes again', async () => {
            const calls: string[] = []
            let bgRef: ModeBackground | null = null
            const current = () => bgRef?.image?.ref ?? 'none'
            const flush = vi.fn(async () => {
                calls.push(`flush:${current()}`)
            })
            uploadAsset.mockImplementation(async () => {
                calls.push('upload')
            })
            const { wrapper, bg } = mountEditor({}, {
                flush,
                onUploading: (v: boolean) => calls.push(`uploading:${v}:${current()}`),
            })
            bgRef = bg
            await chooseFile(wrapper, 'a.png')

            expect(calls).toEqual([
                'flush:none',
                'uploading:true:none',
                'upload',
                'uploading:false:none',
                `flush:asset:a-${STAMP}.png`,
            ])
        })

        it('reports a failed upload and sets nothing', async () => {
            uploadAsset.mockRejectedValue(new Error('nope'))
            const flush = vi.fn(async () => {})
            const { wrapper, bg } = mountEditor({}, { flush })
            await chooseFile(wrapper, 'bad.png')

            expect(uploadAsset).toHaveBeenCalledTimes(1)
            expect(bg.image).toBeUndefined()
            expect(flush).toHaveBeenCalledTimes(1)
            expect((wrapper.emitted('uploading') ?? []).map(e => e[0])).toEqual([true, false])
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error' }))
        })

        it('Replace swaps the ref and keeps fit, position and repeat', async () => {
            uploadAsset.mockResolvedValue(undefined)
            const { wrapper, bg } = mountEditor(image('asset:sky.jpg'))
            await chooseFile(wrapper, 'sky.jpg')
            expect(bg.image).toEqual({ ref: `asset:sky-${STAMP}.jpg`, fit: 'contain', position: 'top', repeat: 'repeat' })
        })

        it('Remove asks first, then clears only the image', async () => {
            const { wrapper, bg } = mountEditor({ ...gradient(), ...image('asset:sky.jpg') })
            await wrapper.find('[data-testid="image-light-remove"]').trigger('click')
            const dialog = wrapper.findComponent(ConfirmDialog)
            expect(dialog.props('visible')).toBe(true)
            expect(bg.image?.ref).toBe('asset:sky.jpg')

            dialog.vm.$emit('confirm')
            await flushPromises()
            expect(bg.image).toBeUndefined()
            expect(bg.gradient?.stops).toHaveLength(3)
        })

        it('calls out an image file that no longer exists', () => {
            const { wrapper } = mountEditor(image('asset:gone.png'))
            expect(wrapper.text()).toContain('“gone.png” no longer exists')
        })

        it('calls out a leftover reference to the removed shared pool', () => {
            const { wrapper } = mountEditor(image('shared:old.png'))
            expect(wrapper.text()).toContain('“old.png” no longer exists')
        })

        it('asks for an upload when the image layer has no file', () => {
            const { wrapper } = mountEditor(image(''))
            expect(wrapper.text()).toContain('Upload an image, or remove the image layer.')
        })
    })

    it('renders no preview of its own (the editor has a Preview section)', () => {
        const { wrapper } = mountEditor(gradient())
        expect(wrapper.find('[data-testid^="bg-preview"]').exists()).toBe(false)
    })

    describe('read-only', () => {
        it('disables every control and ignores clicks', async () => {
            const { wrapper, bg } = mountEditor(
                { ...gradient(), ...image('asset:sky.jpg') },
                { disabled: true },
            )
            const before = JSON.parse(JSON.stringify(bg))

            for (const sel of [
                '[data-testid="stop-light-add"]',
                '[data-testid="stop-light-1-remove"]',
                '[data-testid="stop-light-1-up"]',
                '[data-testid="image-light-upload"]',
                '[data-testid="image-light-remove"]',
            ]) {
                const el = wrapper.find(sel)
                expect(el.attributes('disabled'), sel).toBeDefined()
                await el.trigger('click')
            }
            expect(wrapper.findComponent(ConfirmDialog).props('visible')).toBe(false)
            expect(wrapper.find('[data-testid="stop-light-0"]').attributes('disabled')).toBeDefined()
            expect(wrapper.findAll('[data-testid="bg-base"] button').every(b => b.attributes('disabled') !== undefined)).toBe(true)
            expect(JSON.parse(JSON.stringify(bg))).toEqual(before)
        })
    })
})
