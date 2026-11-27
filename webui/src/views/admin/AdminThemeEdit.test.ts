import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import ColorPicker from 'primevue/colorpicker'
import AdminThemeEdit from './AdminThemeEdit.vue'
import { emptyPalette } from '@/lib/themeEditor'
import type { Palette, Theme } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}
const darkPalette: Palette = { ...palette, bg: '#101010', text: '#eeeeee' }

const routeParams = { id: 'abc123' }
const theme = ref<Theme | undefined>(undefined)
const updateTheme = vi.fn()
const uploadAsset = vi.fn()
const toastAdd = vi.fn()

vi.mock('@/composables/useThemes', () => ({
    useGetTheme: () => ({ data: theme, isLoading: ref(false), isError: ref(false) }),
    useUpdateTheme: () => ({ updateTheme, isUpdating: ref(false) }),
    useThemeAssets: () => ({
        assets: ref(['sky.png']),
        uploadAsset,
        deleteAsset: vi.fn(),
        isUploadingAsset: ref(false),
    }),
}))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: routeParams }),
    useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd }) }))

function mountView() {
    return mount(AdminThemeEdit, {
        global: { plugins: [PrimeVue], directives: { tooltip: {} }, stubs: { RouterLink: true } },
    })
}

// Lets the autosave debounce (1s) elapse and any save settle.
async function settle() {
    await vi.advanceTimersByTimeAsync(1500)
    await flushPromises()
}

const lastPayload = (): Theme => updateTheme.mock.calls[updateTheme.mock.calls.length - 1][0].payload

const setDarkMode = async (wrapper: VueWrapper, on: boolean) => {
    await wrapper.find('[data-testid="dark-mode-toggle"] input').setValue(on)
}

const NOW = Date.UTC(2026, 10, 27)
const STAMP = NOW.toString(36)

describe('AdminThemeEdit', () => {
    beforeEach(() => {
        // PrimeVue's TabList watches its size and Select listens for orientation
        // changes; jsdom has neither API.
        vi.stubGlobal('ResizeObserver', class {
            observe() {}
            unobserve() {}
            disconnect() {}
        })
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        vi.useFakeTimers()
        vi.setSystemTime(NOW)
        vi.clearAllMocks()
        updateTheme.mockResolvedValue({})
        uploadAsset.mockResolvedValue(undefined)
        routeParams.id = 'abc123'
        theme.value = {
            id: 'abc123',
            name: 'Mine',
            darkMode: false,
            background: { light: {}, dark: {} },
            colors: { light: { ...palette }, dark: emptyPalette() },
        }
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    describe('dark mode', () => {
        it('shows no tabs and only the light fields while dark mode is off', async () => {
            const wrapper = mountView()
            await flushPromises()

            const toggle = wrapper.find<HTMLInputElement>('[data-testid="dark-mode-toggle"] input')
            expect(toggle.element.checked).toBe(false)
            expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
            expect(wrapper.find('#color-light-text').exists()).toBe(true)
            expect(wrapper.find('[data-testid="image-light-upload"]').exists()).toBe(true)
            expect(wrapper.find('#color-dark-text').exists()).toBe(false)
            expect(wrapper.find('[data-testid="image-dark-upload"]').exists()).toBe(false)
            expect(wrapper.find('[data-testid="bg-preview-dark"]').exists()).toBe(false)
        })

        it('ticking shows Light and Dark tabs, each with its own colours and background', async () => {
            const wrapper = mountView()
            await flushPromises()
            await setDarkMode(wrapper, true)

            const tabs = wrapper.findAll('[role="tab"]')
            expect(tabs.map(t => t.text())).toEqual(['Light', 'Dark'])
            for (const mode of ['light', 'dark']) {
                expect(wrapper.find(`#color-${mode}-text`).exists(), mode).toBe(true)
                expect(wrapper.find(`[data-testid="image-${mode}-upload"]`).exists(), mode).toBe(true)
                expect(wrapper.find(`[data-testid="bg-preview-${mode}"]`).exists(), mode).toBe(true)
            }
            expect(wrapper.findAll('[data-testid="bg-base"]')).toHaveLength(2)
        })

        it('opens on the Light tab and switches to Dark', async () => {
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            const wrapper = mountView()
            await flushPromises()

            const tabs = wrapper.findAll('[role="tab"]')
            expect(tabs[0].attributes('aria-selected')).toBe('true')
            await tabs[1].trigger('click')
            expect(wrapper.findAll('[role="tab"]')[1].attributes('aria-selected')).toBe('true')
        })

        it('autosaves darkMode and the dark values', async () => {
            const wrapper = mountView()
            await flushPromises()
            await setDarkMode(wrapper, true)
            await wrapper.find('#color-dark-text').setValue('#abcdef')
            await settle()

            const payload = lastPayload()
            expect(payload.darkMode).toBe(true)
            expect(payload.colors.dark.text).toBe('#abcdef')
            expect(payload.colors.light.text).toBe('#333333')
        })

        it('saves the dark background separately from the light one', async () => {
            const gradient = { gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] } }
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            theme.value!.background = { light: structuredClone(gradient), dark: structuredClone(gradient) }
            const wrapper = mountView()
            await flushPromises()

            await wrapper.find('[data-testid="stop-dark-0"]').setValue('#123456')
            await settle()
            const bg = lastPayload().background
            expect(bg.light.gradient?.stops).toEqual(['#000000', '#ffffff'])
            expect(bg.dark.gradient?.stops).toEqual(['#123456', '#ffffff'])
        })

        it('unticking hides the Dark tab and keeps the values; re-ticking restores them', async () => {
            const wrapper = mountView()
            await flushPromises()
            await setDarkMode(wrapper, true)
            await wrapper.find('#color-dark-text').setValue('#abcdef')
            await setDarkMode(wrapper, false)

            expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
            expect(wrapper.find('#color-dark-text').exists()).toBe(false)
            await settle()
            const payload = lastPayload()
            expect(payload.darkMode).toBe(false)
            expect(payload.colors.dark.text).toBe('#abcdef')

            await setDarkMode(wrapper, true)
            expect(wrapper.find<HTMLInputElement>('#color-dark-text').element.value).toBe('#abcdef')
        })

        it('first tick seeds dark from light, without the image', async () => {
            const gradient = { direction: '90deg', stops: ['#000000', '#ffffff'] }
            const image = { ref: 'asset:sky.png', fit: 'cover' as const, position: 'center' as const, repeat: 'no-repeat' as const }
            theme.value!.background = { light: { gradient, image }, dark: {} }
            const wrapper = mountView()
            await flushPromises()
            await setDarkMode(wrapper, true)
            await settle()

            const payload = lastPayload()
            expect(payload.darkMode).toBe(true)
            expect(payload.colors.dark).toEqual(palette)
            expect(payload.background.dark).toEqual({ gradient })
            expect(payload.background.light.image).toEqual(image)
            expect(wrapper.find('[data-testid="image-dark-thumb"]').exists()).toBe(false)
        })

        it('flags invalid dormant dark values while dark mode is off', async () => {
            theme.value!.colors.dark = { primary: 'red' } as Palette
            const wrapper = mountView()
            await flushPromises()

            await wrapper.find('#color-light-text').setValue('#1a2b3c')
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
            expect(wrapper.text()).toContain('Dark mode has invalid values; enable it to fix them')
        })

        it('saves a partial dark palette while dark mode is off', async () => {
            theme.value!.colors.dark = { primary: '#123456' } as Palette
            const wrapper = mountView()
            await flushPromises()

            await wrapper.find('#color-light-text').setValue('#1a2b3c')
            await settle()
            expect(updateTheme).toHaveBeenCalledTimes(1)
            expect(wrapper.text()).not.toContain('Dark mode has invalid values')
        })
    })

    describe('validity', () => {
        it('does not save while a role colour is half-typed, then saves once valid', async () => {
            const wrapper = mountView()
            await flushPromises()

            const text = wrapper.find('#color-light-text')
            await text.setValue('#1a2b')
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
            expect(toastAdd).not.toHaveBeenCalled()
            expect(text.classes()).toContain('p-invalid')

            await text.setValue('#1a2b3c')
            await settle()
            expect(updateTheme).toHaveBeenCalledTimes(1)
            expect(lastPayload().colors.light.text).toBe('#1a2b3c')
        })

        it('does not save a translucent page background', async () => {
            const wrapper = mountView()
            await flushPromises()
            await wrapper.find('#color-light-bg').setValue('#ffffff80')
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
        })

        it('does not save while a gradient stop is invalid', async () => {
            theme.value!.background = {
                light: { gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] } },
                dark: {},
            }
            const wrapper = mountView()
            await flushPromises()

            const stop = wrapper.find('[data-testid="stop-light-1"]')
            await stop.setValue('#ff')
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
            expect(stop.classes()).toContain('p-invalid')

            await stop.setValue('#ff0000')
            await settle()
            expect(updateTheme).toHaveBeenCalledTimes(1)
        })

        it('does not save an image layer with no image chosen', async () => {
            theme.value!.background = {
                light: { image: { ref: '', fit: 'cover', position: 'center', repeat: 'no-repeat' } },
                dark: {},
            }
            const wrapper = mountView()
            await flushPromises()

            await wrapper.find('#color-light-text').setValue('#123456')
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
            expect(toastAdd).not.toHaveBeenCalled()
            expect(wrapper.text()).toContain('Upload an image')
        })

        it('does not save when the theme is only opened', async () => {
            const gradient = { gradient: { direction: 'to right', stops: ['#000000', '#ffffff'] } }
            const loaded: Theme = { ...theme.value!, background: { light: gradient, dark: {} } }
            // The theme arrives after mount, as it does from the server.
            theme.value = undefined
            mountView()
            await flushPromises()
            theme.value = loaded
            await flushPromises()
            await settle()
            expect(updateTheme).not.toHaveBeenCalled()
        })

        it('drops the alpha when the picker writes the page background', async () => {
            const wrapper = mountView()
            await flushPromises()
            const bg = wrapper.find<HTMLInputElement>('#color-light-bg')
            const surface = wrapper.find<HTMLInputElement>('#color-light-surface')
            await bg.setValue('#ffffff80')
            await surface.setValue('#ffffff80')

            const pickers = wrapper.findAllComponents(ColorPicker)
            // Separate ticks, as real picker events are: ThemeColorsEditor
            // emits a whole palette built from its current prop.
            pickers[0].vm.$emit('update:modelValue', '112233') // bg
            await flushPromises()
            pickers[1].vm.$emit('update:modelValue', '445566') // surface
            await flushPromises()
            expect(bg.element.value).toBe('#112233')
            expect(surface.element.value).toBe('#44556680')
        })
    })

    describe('upload', () => {
        const chooseFile = async (wrapper: VueWrapper, name: string) => {
            const input = wrapper.find<HTMLInputElement>('[data-testid="image-upload-input"]')
            Object.defineProperty(input.element, 'files', { value: [{ name, arrayBuffer: async () => new ArrayBuffer(1) }] })
            await input.trigger('change')
            await flushPromises()
        }

        it('saves pending edits first, blocks autosave and the form while uploading, then saves the new ref', async () => {
            let finish: () => void = () => {}
            uploadAsset.mockImplementation(() => new Promise<void>(r => { finish = r }))
            const wrapper = mountView()
            await flushPromises()

            await wrapper.find('#color-light-text').setValue('#1a2b3c')
            await chooseFile(wrapper, 'sky.png')
            // The pending edit was flushed before the upload started.
            expect(updateTheme).toHaveBeenCalledTimes(1)
            expect(uploadAsset).toHaveBeenCalledTimes(1)

            // The whole editor is read-only while the file uploads.
            expect(wrapper.find('[data-testid="theme-name"]').attributes('disabled')).toBeDefined()
            expect(wrapper.find('[data-testid="dark-mode-toggle"] input').attributes('disabled')).toBeDefined()
            expect(wrapper.find('#color-light-text').attributes('disabled')).toBeDefined()
            expect(wrapper.find('[data-testid="image-light-upload"]').attributes('disabled')).toBeDefined()
            await settle()
            expect(updateTheme).toHaveBeenCalledTimes(1)

            finish()
            await flushPromises()
            expect(updateTheme).toHaveBeenCalledTimes(2)
            expect(lastPayload().background.light.image?.ref).toBe(`asset:sky-${STAMP}.png`)
            expect(wrapper.find('#color-light-text').attributes('disabled')).toBeUndefined()
        })

        it('re-enables the form after a failed upload and saves nothing', async () => {
            uploadAsset.mockRejectedValue(new Error('nope'))
            const wrapper = mountView()
            await flushPromises()
            await chooseFile(wrapper, 'sky.png')
            await settle()

            expect(updateTheme).not.toHaveBeenCalled()
            expect(wrapper.find('#color-light-text').attributes('disabled')).toBeUndefined()
        })
    })

    it.each(['default', 'ember'])('keeps the builtin %s read-only', async (id) => {
        routeParams.id = id
        theme.value = { ...theme.value!, id, name: id, builtin: true }
        const wrapper = mountView()
        await flushPromises()

        expect(wrapper.text()).toContain('cannot be edited')
        expect(wrapper.find('#color-light-text').exists()).toBe(false)
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
    })
})
