import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import PrimeVue from 'primevue/config'
import ColorPicker from 'primevue/colorpicker'
import ThemeBackgroundEditor from './ThemeBackgroundEditor.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { EditorBackground } from '@/lib/themeBackgroundDraft'
import type { Palette } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}
const darkPalette: Palette = { ...palette, bg: '#101010' }

const assets = ref<string[] | undefined>([])
const uploadAsset = vi.fn()
const deleteAsset = vi.fn()
const toastAdd = vi.fn()

vi.mock('@/composables/useThemes', () => ({
    useThemeAssets: () => ({ assets, uploadAsset, deleteAsset, isUploadingAsset: ref(false) }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd }) }))

function mountEditor(background: EditorBackground, disabled = false) {
    const bg = reactive(background)
    const wrapper = mount(ThemeBackgroundEditor, {
        props: {
            background: bg,
            themeId: 'abc123',
            colors: { light: { ...palette }, dark: { ...darkPalette } },
            disabled,
        },
        global: { plugins: [PrimeVue], stubs: { RouterLink: true } },
    })
    return { wrapper, bg }
}

// jsdom's File has no arrayBuffer(); the component only needs name + bytes.
const fakeFile = (name: string) => ({ name, arrayBuffer: async () => new ArrayBuffer(1) })

const gradient = (): EditorBackground => ({
    gradient: { direction: '90deg', light: ['#000000', '#ff0000', '#ffffff'] },
})

describe('ThemeBackgroundEditor', () => {
    beforeEach(() => {
        // PrimeVue Select listens for orientation changes; jsdom has no matchMedia.
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        vi.clearAllMocks()
        assets.value = ['own.png', 'sky.jpg', 'notes.txt', 'anim.gif']
    })

    afterEach(() => {
        vi.unstubAllGlobals()
    })

    describe('gradient', () => {
        it('switching the base to a gradient adds a default two-stop gradient, and back removes it', async () => {
            const { wrapper, bg } = mountEditor({})
            const buttons = wrapper.findAll('[data-testid="bg-base"] button')
            await buttons[1].trigger('click')
            expect(bg.gradient?.light).toHaveLength(2)
            expect(bg.gradient?.direction).toBe('135deg')

            await buttons[0].trigger('click')
            expect(bg.gradient).toBeUndefined()
        })

        it('adds a stop repeating the last colour', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-add"]').trigger('click')
            expect(bg.gradient?.light).toEqual(['#000000', '#ff0000', '#ffffff', '#ffffff'])
        })

        it('removes a stop but never below two', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-1-remove"]').trigger('click')
            expect(bg.gradient?.light).toEqual(['#000000', '#ffffff'])
            expect(wrapper.find('[data-testid="stop-light-0-remove"]').attributes('disabled')).toBeDefined()
        })

        it('reorders stops', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-0-down"]').trigger('click')
            expect(bg.gradient?.light).toEqual(['#ff0000', '#000000', '#ffffff'])
            await wrapper.find('[data-testid="stop-light-2-up"]').trigger('click')
            expect(bg.gradient?.light).toEqual(['#ff0000', '#ffffff', '#000000'])
            expect(wrapper.find('[data-testid="stop-light-0-up"]').attributes('disabled')).toBeDefined()
        })

        it('edits a stop as text and flags an invalid one', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            const input = wrapper.find('[data-testid="stop-light-1"]')
            await input.setValue('#12')
            expect(bg.gradient?.light[1]).toBe('#12')
            expect(input.classes()).toContain('p-invalid')
        })

        it('keeps a stop alpha when the picker writes it', async () => {
            const { wrapper, bg } = mountEditor({
                gradient: { direction: '90deg', light: ['#00000080', '#ffffff'] },
            })
            wrapper.findAllComponents(ColorPicker)[0].vm.$emit('update:modelValue', '112233')
            await flushPromises()
            expect(bg.gradient?.light[0]).toBe('#11223380')
        })

        it('toggles separate dark stops, seeded from the light ones', async () => {
            const { wrapper, bg } = mountEditor(gradient())
            await wrapper.find('[data-testid="gradient-dark-toggle"] input').setValue(true)
            expect(bg.gradient?.dark).toEqual(['#000000', '#ff0000', '#ffffff'])
            await wrapper.find('[data-testid="stop-dark-add"]').trigger('click')
            expect(bg.gradient?.dark).toHaveLength(4)
            expect(bg.gradient?.light).toHaveLength(3)

            await wrapper.find('[data-testid="gradient-dark-toggle"] input').setValue(false)
            expect(bg.gradient && 'dark' in bg.gradient).toBe(false)
        })

        it('shows a keyword direction as its angle without rewriting it', () => {
            const { wrapper, bg } = mountEditor({
                gradient: { direction: 'to bottom right', light: ['#000000', '#ffffff'] },
            })
            expect(wrapper.find('[data-testid="bg-angle"]').text()).toBe('135°')
            expect(bg.gradient?.direction).toBe('to bottom right')
        })
    })

    describe('image', () => {
        it('offers only this theme\'s own images, image files only', () => {
            const { wrapper } = mountEditor({})
            expect(wrapper.find('[data-testid="image-light-asset:own.png"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="image-light-asset:sky.jpg"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="image-light-asset:notes.txt"]').exists()).toBe(false)
            expect(wrapper.find('[data-testid="image-light-asset:anim.gif"]').exists()).toBe(false)
            const img = wrapper.find('[data-testid="image-light-asset:sky.jpg"] img')
            expect(img.attributes('src')).toContain('/api/v0/themes/abc123/assets/sky.jpg')
            expect(wrapper.html()).not.toContain('/api/v0/data/backgrounds')
        })

        it('prompts for an upload when the theme has no images', () => {
            assets.value = []
            const { wrapper } = mountEditor({})
            expect(wrapper.text()).toContain('No images in this theme yet. Upload one.')
            expect(wrapper.find('[data-testid="image-upload-light"]').exists()).toBe(true)
        })

        it('picking an image creates the image layer with defaults, clearing removes it', async () => {
            const { wrapper, bg } = mountEditor({})
            await wrapper.find('[data-testid="image-light-asset:sky.jpg"]').trigger('click')
            expect(bg.image).toEqual({ light: 'asset:sky.jpg', fit: 'cover', position: 'center', repeat: 'no-repeat' })
            expect(wrapper.find('[data-testid="image-light-asset:sky.jpg"]').classes()).toContain('selected')

            await wrapper.find('[data-testid="image-light-asset:own.png"]').trigger('click')
            expect(bg.image?.light).toBe('asset:own.png')

            await wrapper.find('[data-testid="image-clear"]').trigger('click')
            expect(bg.image).toBeUndefined()
        })

        it('picks a separate dark image', async () => {
            const { wrapper, bg } = mountEditor({
                image: { light: 'asset:sky.jpg', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            })
            await wrapper.find('[data-testid="image-dark-toggle"] input').setValue(true)
            expect(bg.image?.dark).toBe('asset:sky.jpg')
            await wrapper.find('[data-testid="image-dark-asset:own.png"]').trigger('click')
            expect(bg.image?.dark).toBe('asset:own.png')
            expect(bg.image?.light).toBe('asset:sky.jpg')
        })

        it('uploads into the theme and selects the upload', async () => {
            uploadAsset.mockResolvedValue(undefined)
            const { wrapper, bg } = mountEditor({})
            const input = wrapper.find<HTMLInputElement>('[data-testid="image-upload-input"]')
            Object.defineProperty(input.element, 'files', { value: [fakeFile('new.webp')] })
            await input.trigger('change')
            await flushPromises()

            expect(uploadAsset).toHaveBeenCalledWith(expect.objectContaining({ name: 'new.webp' }))
            expect(bg.image?.light).toBe('asset:new.webp')
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'success' }))
        })

        it('reports a failed upload and selects nothing', async () => {
            uploadAsset.mockRejectedValue(new Error('nope'))
            const { wrapper, bg } = mountEditor({})
            const input = wrapper.find<HTMLInputElement>('[data-testid="image-upload-input"]')
            Object.defineProperty(input.element, 'files', { value: [fakeFile('bad.png')] })
            await input.trigger('change')
            await flushPromises()

            expect(uploadAsset).toHaveBeenCalledTimes(1)
            expect(bg.image).toBeUndefined()
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error' }))
        })

        it('calls out an image that no longer exists', () => {
            const { wrapper } = mountEditor({
                image: { light: 'asset:gone.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            })
            expect(wrapper.text()).toContain('“gone.png” no longer exists')
        })

        it('calls out a leftover reference to the removed shared pool', () => {
            const { wrapper } = mountEditor({
                image: { light: 'shared:old.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            })
            expect(wrapper.text()).toContain('“old.png” no longer exists')
        })

        it('deletes an image from the theme after confirmation', async () => {
            deleteAsset.mockResolvedValue(undefined)
            const { wrapper, bg } = mountEditor({})
            await wrapper.find('[data-testid="image-delete-light-asset:own.png"]').trigger('click')
            const dialog = wrapper.findComponent(ConfirmDialog)
            expect(dialog.props('name')).toBe('own.png')
            expect(dialog.props('message')).toBe('Delete this image from the theme?')
            expect(deleteAsset).not.toHaveBeenCalled()

            dialog.vm.$emit('confirm')
            await flushPromises()
            expect(deleteAsset).toHaveBeenCalledWith('own.png')
            expect(bg.image).toBeUndefined()
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'success' }))
        })

        it('warns when the image to delete is used by the background', async () => {
            const { wrapper } = mountEditor({
                image: { light: 'asset:sky.jpg', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            })
            await wrapper.find('[data-testid="image-delete-light-asset:sky.jpg"]').trigger('click')
            expect(wrapper.findComponent(ConfirmDialog).props('message')).toContain('used by the background')
        })

        it('reports a failed delete', async () => {
            deleteAsset.mockRejectedValue(new Error('nope'))
            const { wrapper } = mountEditor({})
            await wrapper.find('[data-testid="image-delete-light-asset:own.png"]').trigger('click')
            wrapper.findComponent(ConfirmDialog).vm.$emit('confirm')
            await flushPromises()
            expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error' }))
        })
    })

    describe('preview', () => {
        it('renders the background per colour mode the way the viewer does', () => {
            const { wrapper } = mountEditor({
                gradient: { direction: '90deg', light: ['#000000', '#ffffff'], dark: ['#111111', '#222222'] },
            })
            const light = wrapper.find('[data-testid="bg-preview-light"]').attributes('style')
            const dark = wrapper.find('[data-testid="bg-preview-dark"]').attributes('style')
            expect(light).toContain('--dashi-page-bg: linear-gradient(90deg,#000000,#ffffff)')
            expect(light).toContain('--dashi-bg: #ffffff')
            expect(dark).toContain('--dashi-page-bg: linear-gradient(90deg,#111111,#222222)')
            expect(dark).toContain('--dashi-bg: #101010')
        })

        it('updates as the background is edited', async () => {
            const { wrapper } = mountEditor(gradient())
            await wrapper.find('[data-testid="stop-light-0"]').setValue('#abcdef')
            const light = wrapper.find('[data-testid="bg-preview-light"]').attributes('style')
            expect(light).toContain('linear-gradient(90deg,#abcdef,#ff0000,#ffffff)')
        })
    })

    describe('read-only', () => {
        it('disables every control and ignores clicks', async () => {
            const { wrapper, bg } = mountEditor(
                {
                    gradient: { direction: '90deg', light: ['#000000', '#ff0000', '#ffffff'] },
                    image: { light: 'asset:sky.jpg', fit: 'cover', position: 'center', repeat: 'no-repeat' },
                },
                true,
            )
            // A read-only theme offers no delete at all.
            expect(wrapper.find('[data-testid="image-delete-light-asset:own.png"]').exists()).toBe(false)
            const before = JSON.parse(JSON.stringify(bg))

            for (const sel of [
                '[data-testid="stop-light-add"]',
                '[data-testid="stop-light-1-remove"]',
                '[data-testid="stop-light-1-up"]',
                '[data-testid="image-clear"]',
                '[data-testid="image-upload-light"]',
                '[data-testid="image-light-asset:own.png"]',
            ]) {
                const el = wrapper.find(sel)
                expect(el.attributes('disabled'), sel).toBeDefined()
                await el.trigger('click')
            }
            expect(wrapper.find('[data-testid="stop-light-0"]').attributes('disabled')).toBeDefined()
            expect(wrapper.findAll('[data-testid="bg-base"] button').every(b => b.attributes('disabled') !== undefined)).toBe(true)
            expect(JSON.parse(JSON.stringify(bg))).toEqual(before)
        })
    })
})
