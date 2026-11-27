import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import ColorPicker from 'primevue/colorpicker'
import AdminThemeEdit from './AdminThemeEdit.vue'
import type { Palette, Theme } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}

const routeParams = { id: 'abc123' }
const theme = ref<Theme | undefined>(undefined)
const updateTheme = vi.fn()
const toastAdd = vi.fn()

vi.mock('@/composables/useThemes', () => ({
    useGetTheme: () => ({ data: theme, isLoading: ref(false), isError: ref(false) }),
    useUpdateTheme: () => ({ updateTheme, isUpdating: ref(false) }),
    useThemeAssets: () => ({
        assets: ref(['sky.png']),
        uploadAsset: vi.fn(),
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

describe('AdminThemeEdit', () => {
    beforeEach(() => {
        // PrimeVue Select listens for orientation changes; jsdom has no matchMedia.
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        vi.useFakeTimers()
        vi.clearAllMocks()
        updateTheme.mockResolvedValue({})
        routeParams.id = 'abc123'
        theme.value = {
            id: 'abc123',
            name: 'Mine',
            darkMode: true,
            background: { light: {}, dark: {} },
            colors: { light: { ...palette }, dark: { ...palette } },
        }
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    it('does not save while a role colour is half-typed, then saves once valid', async () => {
        const wrapper = mountView()
        await flushPromises()

        const text = wrapper.find('#color-text')
        await text.setValue('#1a2b')
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
        expect(toastAdd).not.toHaveBeenCalled()
        expect(text.classes()).toContain('p-invalid')

        await text.setValue('#1a2b3c')
        await settle()
        expect(updateTheme).toHaveBeenCalledTimes(1)
        expect(updateTheme.mock.calls[0][0].payload.colors.light.text).toBe('#1a2b3c')
    })

    it('does not save a translucent page background', async () => {
        const wrapper = mountView()
        await flushPromises()
        await wrapper.find('#color-bg').setValue('#ffffff80')
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
    })

    it('does not save while a gradient stop is invalid', async () => {
        const gradient = { gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] } }
        theme.value!.background = { light: gradient, dark: gradient }
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

        await wrapper.find('#color-text').setValue('#123456')
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
        expect(toastAdd).not.toHaveBeenCalled()
        expect(wrapper.text()).toContain('Pick an image')
    })

    it('saves a background image picked from the theme\'s own images', async () => {
        const wrapper = mountView()
        await flushPromises()

        await wrapper.find('[data-testid="image-light-asset:sky.png"]').trigger('click')
        await settle()
        expect(updateTheme).toHaveBeenCalledTimes(1)
        const image = { ref: 'asset:sky.png', fit: 'cover', position: 'center', repeat: 'no-repeat' }
        expect(updateTheme.mock.calls[0][0].payload.background).toEqual({ light: { image }, dark: { image } })
    })

    it('saves different dark stops to background.dark.gradient.stops', async () => {
        const gradient = { gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] } }
        theme.value!.background = { light: gradient, dark: gradient }
        const wrapper = mountView()
        await flushPromises()

        await wrapper.find('[data-testid="gradient-dark-toggle"] input').setValue(true)
        await wrapper.find('[data-testid="stop-dark-0"]').setValue('#123456')
        await settle()
        expect(updateTheme).toHaveBeenCalledTimes(1)
        const bg = updateTheme.mock.calls[0][0].payload.background
        expect(bg.light.gradient.stops).toEqual(['#000000', '#ffffff'])
        expect(bg.dark.gradient.stops).toEqual(['#123456', '#ffffff'])
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

    it('does not save a malformed dark role while dark mode is off', async () => {
        theme.value!.darkMode = false
        theme.value!.colors.dark = { primary: 'red' } as Palette
        const wrapper = mountView()
        await flushPromises()

        await wrapper.find('#color-text').setValue('#1a2b3c')
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
    })

    it('saves a partial dark palette while dark mode is off', async () => {
        theme.value!.darkMode = false
        theme.value!.colors.dark = { primary: '#123456' } as Palette
        const wrapper = mountView()
        await flushPromises()

        await wrapper.find('#color-text').setValue('#1a2b3c')
        await settle()
        expect(updateTheme).toHaveBeenCalledTimes(1)
    })

    it('drops the alpha when the picker writes the page background', async () => {
        const wrapper = mountView()
        await flushPromises()
        const bg = wrapper.find<HTMLInputElement>('#color-bg')
        const surface = wrapper.find<HTMLInputElement>('#color-surface')
        await bg.setValue('#ffffff80')
        await surface.setValue('#ffffff80')

        const pickers = wrapper.findAllComponents(ColorPicker)
        pickers[0].vm.$emit('update:modelValue', '112233') // bg
        pickers[1].vm.$emit('update:modelValue', '445566') // surface
        await flushPromises()
        expect(bg.element.value).toBe('#112233')
        expect(surface.element.value).toBe('#44556680')
    })

    it.each(['default', 'ember'])('keeps the builtin %s read-only', async (id) => {
        routeParams.id = id
        theme.value = { ...theme.value!, id, name: id, builtin: true }
        const wrapper = mountView()
        await flushPromises()

        expect(wrapper.text()).toContain('cannot be edited')
        expect(wrapper.find('#color-text').exists()).toBe(false)
        await settle()
        expect(updateTheme).not.toHaveBeenCalled()
    })
})
