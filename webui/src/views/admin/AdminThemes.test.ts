import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import AdminThemes from './AdminThemes.vue'
import type { Palette, ThemeMeta } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}
const meta = (over: Partial<ThemeMeta>): ThemeMeta => ({
    id: 'x', name: 'X', darkMode: true, usedBy: 0, previewCss: '', colors: { light: palette, dark: palette }, ...over,
})

const push = vi.fn()
const copyTheme = vi.fn()
const createTheme = vi.fn()
const themes = ref<ThemeMeta[]>([])

vi.mock('@/composables/useThemes', () => ({
    useListThemes: () => ({
        themes,
        isLoading: ref(false),
        createTheme,
        isCreating: ref(false),
        deleteTheme: vi.fn(),
        copyTheme,
        isCopying: ref(false),
    }),
}))
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: vi.fn() }) }))

function mountView() {
    return mount(AdminThemes, {
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
        },
    })
}

function row(wrapper: ReturnType<typeof mountView>, index: number) {
    return wrapper.findAll('tbody tr')[index]
}

describe('AdminThemes', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        themes.value = [
            meta({ id: 'default', name: 'Default', builtin: true }),
            meta({ id: 'abc123', name: 'Mine', usedBy: 2 }),
        ]
    })

    it('offers only Copy on the builtin row', () => {
        const wrapper = mountView()
        const builtin = row(wrapper, 0)
        expect(builtin.find('button[aria-label="Copy"]').exists()).toBe(true)
        expect(builtin.find('button[aria-label="Edit"]').exists()).toBe(false)
        expect(builtin.find('button[aria-label="Delete"]').exists()).toBe(false)

        const own = row(wrapper, 1)
        expect(own.find('button[aria-label="Copy"]').exists()).toBe(true)
        expect(own.find('button[aria-label="Edit"]').exists()).toBe(true)
        expect(own.find('button[aria-label="Delete"]').exists()).toBe(true)
    })

    it('creates a new theme with the darkMode flag the server requires', async () => {
        createTheme.mockResolvedValue({ id: 'new123', name: 'New theme' })
        const wrapper = mountView()

        await wrapper.find('.admin-section-header button').trigger('click')
        await flushPromises()

        expect(createTheme).toHaveBeenCalledWith(expect.objectContaining({ name: 'New theme', darkMode: true }))
        expect(push).toHaveBeenCalledWith({ name: 'admin-theme-edit', params: { id: 'new123' } })
    })

    it('copies a theme and opens the copy in the editor', async () => {
        copyTheme.mockResolvedValue({ id: 'def456', name: 'Copy of Default' })
        const wrapper = mountView()

        await row(wrapper, 0).find('button[aria-label="Copy"]').trigger('click')
        await flushPromises()

        expect(copyTheme).toHaveBeenCalledWith('default')
        expect(push).toHaveBeenCalledWith({ name: 'admin-theme-edit', params: { id: 'def456' } })
    })
})
