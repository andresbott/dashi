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
            expect(wrapper.find('[data-testid="preview-light"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="preview-dark"]').exists()).toBe(false)
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
                expect(wrapper.find(`[data-testid="preview-${mode}"]`).exists(), mode).toBe(true)
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

        it('marks the tab of an invalid mode, and clears it once fixed', async () => {
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            const wrapper = mountView()
            await flushPromises()
            expect(wrapper.findAll('[role="tab"]')[0].attributes('aria-selected')).toBe('true')
            expect(wrapper.find('[data-testid="tab-invalid-light"]').exists()).toBe(false)
            expect(wrapper.find('[data-testid="tab-invalid-dark"]').exists()).toBe(false)

            await wrapper.find('#color-dark-text').setValue('#12')
            const marker = wrapper.find('[data-testid="tab-invalid-dark"]')
            expect(marker.exists()).toBe(true)
            expect(marker.attributes('aria-label')).toBeTruthy()
            expect(wrapper.findAll('[role="tab"]')[1].find('[data-testid="tab-invalid-dark"]').exists()).toBe(true)
            expect(wrapper.find('[data-testid="tab-invalid-light"]').exists()).toBe(false)

            await wrapper.find('#color-dark-text').setValue('#123456')
            expect(wrapper.find('[data-testid="tab-invalid-dark"]').exists()).toBe(false)
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

    describe('preview', () => {
        const gradientBg = (from: string) => ({ gradient: { direction: '90deg', stops: [from, '#ffffff'] } })
        const previewOf = (wrapper: VueWrapper, mode: string) =>
            wrapper.find(`[data-testid="section-preview-${mode}"] [data-testid="preview-${mode}"]`)

        it('has no side column and no preview toggle of its own', async () => {
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            const wrapper = mountView()
            await flushPromises()

            expect(wrapper.find('.edit-preview').exists()).toBe(false)
            expect(wrapper.find('.edit-grid').exists()).toBe(false)
            // The only SelectButtons left are the two background Base pickers.
            expect(wrapper.findAll('.p-selectbutton')).toHaveLength(2)
            expect(wrapper.findAll('[data-testid^="preview-"]')).toHaveLength(2)
        })

        it('shows the light colours and background after the other sections in light-only mode', async () => {
            theme.value!.background = { light: gradientBg('#aa0000'), dark: {} }
            const wrapper = mountView()
            await flushPromises()

            const sections = wrapper.findAll('[data-testid^="section-"]').map(s => s.attributes('data-testid'))
            expect(sections).toEqual(['section-colors-light', 'section-background-light', 'section-preview-light'])
            expect(wrapper.find('[data-testid="section-preview-light"]').text()).toContain('Preview')
            const style = previewOf(wrapper, 'light').attributes('style')
            expect(style).toContain('--dashi-bg: #ffffff')
            expect(style).toContain('#aa0000')

            await wrapper.find('#color-light-bg').setValue('#fafafa')
            expect(previewOf(wrapper, 'light').attributes('style')).toContain('--dashi-bg: #fafafa')
        })

        it('shows each tab its own mode', async () => {
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            theme.value!.background = { light: gradientBg('#aa0000'), dark: gradientBg('#00bb00') }
            const wrapper = mountView()
            await flushPromises()

            for (const mode of ['light', 'dark']) {
                const ids = wrapper.find(`[data-testid="section-colors-${mode}"]`).element.parentElement!
                    .querySelectorAll(':scope > [data-testid^="section-"]')
                expect(Array.from(ids, e => e.getAttribute('data-testid')), mode).toEqual([
                    `section-colors-${mode}`, `section-background-${mode}`, `section-preview-${mode}`,
                ])
            }
            const light = previewOf(wrapper, 'light').attributes('style')
            const dark = previewOf(wrapper, 'dark').attributes('style')
            expect(light).toContain('--dashi-bg: #ffffff')
            expect(light).toContain('#aa0000')
            expect(dark).toContain('--dashi-bg: #101010')
            expect(dark).toContain('#00bb00')
            expect(dark).not.toContain('#aa0000')
        })
    })

    describe('collapsible sections', () => {
        const SECTIONS = ['colors', 'background', 'preview'] as const
        type Section = (typeof SECTIONS)[number]

        const panel = (wrapper: VueWrapper, section: Section, mode: string) =>
            wrapper.find(`[data-testid="section-${section}-${mode}"]`)
        const toggleOf = (wrapper: VueWrapper, section: Section, mode: string) =>
            panel(wrapper, section, mode).find('button[aria-expanded]')
        const expanded = (wrapper: VueWrapper, section: Section, mode: string) =>
            toggleOf(wrapper, section, mode).attributes('aria-expanded') === 'true'
        // For a section on screen: the toggle state and the content's
        // visibility must agree.
        const isOpen = (wrapper: VueWrapper, section: Section, mode: string) => {
            const open = expanded(wrapper, section, mode)
            expect(panel(wrapper, section, mode).find('[role="region"]').isVisible()).toBe(open)
            return open
        }

        it('starts with every section open and toggles each on its own in light-only mode', async () => {
            const wrapper = mountView()
            await flushPromises()
            expect(panel(wrapper, 'colors', 'light').text()).toContain('Colours')
            expect(panel(wrapper, 'background', 'light').text()).toContain('Background')
            expect(panel(wrapper, 'preview', 'light').text()).toContain('Preview')
            for (const s of SECTIONS) expect(isOpen(wrapper, s, 'light'), s).toBe(true)

            await toggleOf(wrapper, 'colors', 'light').trigger('click')
            expect(isOpen(wrapper, 'colors', 'light')).toBe(false)
            expect(isOpen(wrapper, 'background', 'light')).toBe(true)
            expect(isOpen(wrapper, 'preview', 'light')).toBe(true)

            await toggleOf(wrapper, 'background', 'light').trigger('click')
            await toggleOf(wrapper, 'colors', 'light').trigger('click')
            expect(isOpen(wrapper, 'colors', 'light')).toBe(true)
            expect(isOpen(wrapper, 'background', 'light')).toBe(false)
            expect(isOpen(wrapper, 'preview', 'light')).toBe(true)

            await toggleOf(wrapper, 'preview', 'light').trigger('click')
            expect(isOpen(wrapper, 'preview', 'light')).toBe(false)
            expect(isOpen(wrapper, 'colors', 'light')).toBe(true)
        })

        it('gives each tab its own sections, open at first, and keeps them across tab switches', async () => {
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            const wrapper = mountView()
            await flushPromises()
            for (const s of SECTIONS) {
                expect(isOpen(wrapper, s, 'light'), s).toBe(true)
                expect(expanded(wrapper, s, 'dark'), s).toBe(true)
            }

            // Light tab: collapse Background only.
            await toggleOf(wrapper, 'background', 'light').trigger('click')
            expect(isOpen(wrapper, 'colors', 'light')).toBe(true)
            expect(isOpen(wrapper, 'background', 'light')).toBe(false)
            // The Dark tab's sections are untouched.
            expect(expanded(wrapper, 'colors', 'dark')).toBe(true)
            expect(expanded(wrapper, 'background', 'dark')).toBe(true)

            // Dark tab: collapse Colours and Preview.
            const tabs = wrapper.findAll('[role="tab"]')
            await tabs[1].trigger('click')
            await toggleOf(wrapper, 'colors', 'dark').trigger('click')
            await toggleOf(wrapper, 'preview', 'dark').trigger('click')
            expect(isOpen(wrapper, 'colors', 'dark')).toBe(false)
            expect(isOpen(wrapper, 'background', 'dark')).toBe(true)
            expect(isOpen(wrapper, 'preview', 'dark')).toBe(false)

            // Back and forth: each tab keeps its own state.
            await tabs[0].trigger('click')
            expect(isOpen(wrapper, 'colors', 'light')).toBe(true)
            expect(isOpen(wrapper, 'background', 'light')).toBe(false)
            expect(isOpen(wrapper, 'preview', 'light')).toBe(true)
            await tabs[1].trigger('click')
            expect(isOpen(wrapper, 'colors', 'dark')).toBe(false)
            expect(isOpen(wrapper, 'background', 'dark')).toBe(true)
            expect(isOpen(wrapper, 'preview', 'dark')).toBe(false)
        })

        const marker = (wrapper: VueWrapper, section: Section, mode: string) =>
            wrapper.find(`[data-testid="section-invalid-${section}-${mode}"]`)
        const markers = (wrapper: VueWrapper) =>
            wrapper.findAll('[data-testid^="section-invalid-"]').map(m => m.attributes('data-testid'))

        it('marks a collapsed Colours section holding an invalid value, and clears it once fixed', async () => {
            const wrapper = mountView()
            await flushPromises()
            await toggleOf(wrapper, 'colors', 'light').trigger('click')
            expect(isOpen(wrapper, 'colors', 'light')).toBe(false)
            expect(markers(wrapper)).toEqual([])

            await wrapper.find('#color-light-text').setValue('#12')
            expect(markers(wrapper)).toEqual(['section-invalid-colors-light'])
            const m = marker(wrapper, 'colors', 'light')
            // In the header, next to the title that stays.
            expect(panel(wrapper, 'colors', 'light').find('.p-panel-header').find('[data-testid="section-invalid-colors-light"]').exists()).toBe(true)
            expect(panel(wrapper, 'colors', 'light').find('.p-panel-header').text()).toContain('Colours')
            expect(m.classes()).toContain('tab-invalid')
            expect(m.attributes('aria-label')).toBe('Colours has invalid values')

            await wrapper.find('#color-light-text').setValue('#123456')
            expect(markers(wrapper)).toEqual([])
        })

        it('marks only the Background section of the mode with an invalid background', async () => {
            const gradient = { gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] } }
            theme.value!.darkMode = true
            theme.value!.colors.dark = { ...darkPalette }
            theme.value!.background = { light: structuredClone(gradient), dark: structuredClone(gradient) }
            const wrapper = mountView()
            await flushPromises()
            expect(markers(wrapper)).toEqual([])

            await wrapper.find('[data-testid="stop-dark-0"]').setValue('#12')
            expect(markers(wrapper)).toEqual(['section-invalid-background-dark'])
            expect(marker(wrapper, 'background', 'dark').attributes('aria-label')).toBe('Background has invalid values')
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

        it('says above the sections whether changes are being saved', async () => {
            const wrapper = mountView()
            await flushPromises()
            const status = wrapper.find('[data-testid="save-status"]')
            expect(status.text()).toBe('Changes are saved automatically.')
            // It sits with the Name and dark mode fields, before any section.
            const order = status.element.compareDocumentPosition(
                wrapper.find('[data-testid="section-colors-light"]').element,
            )
            expect(order & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()

            await wrapper.find('#color-light-text').setValue('#12')
            expect(wrapper.find('[data-testid="save-status"]').text()).toBe('Fix the highlighted fields to save your changes.')
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
            // Same tick: ThemeColorsEditor writes each role into the bound
            // palette, so neither edit overwrites the other.
            pickers[0].vm.$emit('update:modelValue', '112233') // bg
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
