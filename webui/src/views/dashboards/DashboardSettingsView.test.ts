import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import Select from 'primevue/select'
import DashboardSettingsView from './DashboardSettingsView.vue'
import type { Dashboard } from '@/types/dashboard'
import type { Palette, ThemeMeta } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}
const meta = (id: string, name: string): ThemeMeta => ({
    id, name, usedBy: 0, previewCss: '', colors: { light: palette, dark: palette },
})

const draft = ref<Dashboard | null>(null)
const themes = ref<ThemeMeta[] | undefined>(undefined)
const widgetThemes = ref<{ name: string }[] | undefined>(undefined)

vi.mock('@/composables/useDashboardDraft', () => ({
    useDashboardDraft: () => ({
        draft,
        isLoading: ref(false),
        isError: ref(false),
        saveStatus: ref('idle'),
        hasUnsaved: ref(false),
        flush: vi.fn(),
    }),
}))
vi.mock('@/composables/useDashboards', () => ({
    useDashboardAuth: () => ({
        auth: ref(null),
        isLoadingAuth: ref(false),
        setAuth: vi.fn(),
        isSettingAuth: ref(false),
        deleteAuth: vi.fn(),
        isDeletingAuth: ref(false),
    }),
}))
vi.mock('@/composables/useThemes', () => ({ useListThemes: () => ({ themes }) }))
vi.mock('@/composables/useWidgetThemes', () => ({ useWidgetThemes: () => ({ data: widgetThemes }) }))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: { id: 'd1' } }),
    useRouter: () => ({ push: vi.fn(), back: vi.fn(), options: { history: { state: {} } } }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: vi.fn() }) }))

async function mountAppearance() {
    const wrapper = mount(DashboardSettingsView, {
        global: { plugins: [PrimeVue], directives: { tooltip: {} } },
    })
    const nav = wrapper.findAll('.settings-nav-item').find(n => n.text().includes('Appearance'))
    await nav!.trigger('click')
    await flushPromises()
    return wrapper
}

// The Appearance section's Selects, in order: theme, widget theme, ...
const selects = (wrapper: Awaited<ReturnType<typeof mountAppearance>>) => wrapper.findAllComponents(Select)

describe('DashboardSettingsView', () => {
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        themes.value = undefined
        widgetThemes.value = undefined
        draft.value = {
            id: 'd1', name: 'D', icon: '', type: 'interactive',
            container: { maxWidth: '', verticalAlign: 'top', horizontalAlign: 'center' },
            themeId: 'gone12', widgetTheme: 'vanished', pages: [],
        } as Dashboard
    })

    afterEach(() => {
        vi.unstubAllGlobals()
    })

    it('shows Default for a dangling theme and widget theme once the lists load', async () => {
        const wrapper = await mountAppearance()
        // Before the lists load the stored values are shown, so nothing flashes.
        expect(selects(wrapper)[0].props('modelValue')).toBe('gone12')
        expect(selects(wrapper)[1].props('modelValue')).toBe('vanished')

        themes.value = [meta('default', 'Default'), meta('abc123', 'Mine')]
        widgetThemes.value = [{ name: 'default' }, { name: 'retro' }]
        await flushPromises()

        expect(selects(wrapper)[0].props('modelValue')).toBe('default')
        expect(selects(wrapper)[1].props('modelValue')).toBe('default')
        // Displaying the fallback does not rewrite the stored references.
        expect(draft.value!.themeId).toBe('gone12')
        expect(draft.value!.widgetTheme).toBe('vanished')
    })

    describe('dither select', () => {
        const mountGeneral = async () => {
            const wrapper = mount(DashboardSettingsView, {
                global: { plugins: [PrimeVue], directives: { tooltip: {} } },
            })
            await flushPromises()
            return wrapper
        }
        const ditherSelect = (wrapper: Awaited<ReturnType<typeof mountGeneral>>) =>
            wrapper.findAllComponents(Select).find(s => s.attributes('data-testid') === 'dither-select')

        it('is hidden for interactive dashboards', async () => {
            const wrapper = await mountGeneral()
            expect(ditherSelect(wrapper)).toBeUndefined()
        })

        it('shows Plain when the field is absent and stores a picked mode', async () => {
            draft.value!.type = 'image'
            const wrapper = await mountGeneral()
            const select = ditherSelect(wrapper)!
            expect(select.props('modelValue')).toBe('plain')
            expect(wrapper.find('[data-testid="dither-hint"]').text()).toContain('Floyd-Steinberg')

            select.vm.$emit('update:modelValue', 'grayscale')
            await flushPromises()
            expect(draft.value!.dither).toBe('grayscale')
            expect(select.props('modelValue')).toBe('grayscale')
            expect(wrapper.find('[data-testid="dither-hint"]').text()).toContain('black-and-white')
        })

        it('removes the field when Plain is picked again', async () => {
            draft.value!.type = 'image'
            draft.value!.dither = 'auto'
            const wrapper = await mountGeneral()
            const select = ditherSelect(wrapper)!
            expect(select.props('modelValue')).toBe('auto')

            select.vm.$emit('update:modelValue', 'plain')
            await flushPromises()
            expect('dither' in draft.value!).toBe(false)
        })

        it('offers plain, auto and every server preset', async () => {
            draft.value!.type = 'image'
            const wrapper = await mountGeneral()
            const values = (ditherSelect(wrapper)!.props('options') as { value: string }[]).map(o => o.value)
            expect(values).toEqual(['plain', 'auto', 'balanced', 'dynamic', 'vivid', 'soft', 'grayscale'])
        })
    })

    it('shows a known theme and widget theme as stored', async () => {
        draft.value!.themeId = 'abc123'
        draft.value!.widgetTheme = 'retro'
        themes.value = [meta('default', 'Default'), meta('abc123', 'Mine')]
        widgetThemes.value = [{ name: 'default' }, { name: 'retro' }]
        const wrapper = await mountAppearance()
        expect(selects(wrapper)[0].props('modelValue')).toBe('abc123')
        expect(selects(wrapper)[1].props('modelValue')).toBe('retro')
    })
})
