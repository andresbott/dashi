import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import Select from 'primevue/select'
import ToggleSwitch from 'primevue/toggleswitch'
import DashboardSettingsView from './DashboardSettingsView.vue'
import AlignmentPicker from '@/components/dashboards/AlignmentPicker.vue'
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
const editorUrl = ref('')
vi.mock('@/lib/serverInfo', () => ({ publicEditorUrl: () => editorUrl.value }))

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

    describe('edit button toggle', () => {
        const mountGeneral = async () => {
            const wrapper = mount(DashboardSettingsView, {
                global: { plugins: [PrimeVue], directives: { tooltip: {} } },
            })
            await flushPromises()
            return wrapper
        }
        const toggle = (wrapper: Awaited<ReturnType<typeof mountGeneral>>) =>
            wrapper.findAllComponents(ToggleSwitch).find(t => t.attributes('data-testid') === 'edit-button-toggle')
        const urlHint = '[data-testid="edit-button-url-hint"]'

        beforeEach(() => {
            editorUrl.value = 'https://editor.local'
        })

        it('is off by default for interactive dashboards', async () => {
            const wrapper = await mountGeneral()
            expect(toggle(wrapper)!.props('modelValue')).toBe(false)
            expect(wrapper.find('label[for]').exists()).toBe(true)
            expect(wrapper.text()).toContain('Show Edit button')
        })

        it('is shown for dashboards without a stored type', async () => {
            delete (draft.value as Partial<Dashboard>).type
            const wrapper = await mountGeneral()
            expect(toggle(wrapper)).toBeDefined()
        })

        it('is hidden for image dashboards and keeps the stored value', async () => {
            draft.value!.type = 'image'
            draft.value!.showEditButton = true
            const wrapper = await mountGeneral()
            expect(toggle(wrapper)).toBeUndefined()
            expect(draft.value!.showEditButton).toBe(true)
        })

        it('stores the flag when turned on and removes it when turned off', async () => {
            const wrapper = await mountGeneral()
            toggle(wrapper)!.vm.$emit('update:modelValue', true)
            await flushPromises()
            expect(draft.value!.showEditButton).toBe(true)
            expect(toggle(wrapper)!.props('modelValue')).toBe(true)

            toggle(wrapper)!.vm.$emit('update:modelValue', false)
            await flushPromises()
            expect('showEditButton' in draft.value!).toBe(false)
        })

        it('hints at the editor public URL when none is advertised', async () => {
            editorUrl.value = ''
            const wrapper = await mountGeneral()
            expect(toggle(wrapper)!.props('disabled')).toBeFalsy()
            const hint = wrapper.find(urlHint)
            expect(hint.exists()).toBe(true)
            expect(hint.text()).toContain('server.editor.publicUrl')
            expect(hint.text()).toContain('editor_public_url')
        })

        it('shows no hint when the editor public URL is set', async () => {
            const wrapper = await mountGeneral()
            expect(toggle(wrapper)).toBeDefined()
            expect(wrapper.find(urlHint).exists()).toBe(false)
        })
    })

    describe('max width', () => {
        const maxWidthSelect = (wrapper: Awaited<ReturnType<typeof mountAppearance>>) =>
            wrapper.findAllComponents(Select).find(s => s.attributes('data-testid') === 'max-width-select')!

        it('offers common widths in an editable dropdown', async () => {
            draft.value!.container.maxWidth = '1600px'
            const wrapper = await mountAppearance()
            const select = maxWidthSelect(wrapper)
            expect(select.props('editable')).toBe(true)
            expect(select.props('modelValue')).toBe('1600px')
            expect(select.props('options')).toEqual(['100%', '1920px', '1600px', '1440px', '1280px', '1024px', '800px'])
        })

        it('stores a picked preset', async () => {
            const wrapper = await mountAppearance()
            maxWidthSelect(wrapper).vm.$emit('update:modelValue', '1280px')
            await flushPromises()
            expect(draft.value!.container.maxWidth).toBe('1280px')
        })

        it('stores a typed custom value', async () => {
            const wrapper = await mountAppearance()
            const input = maxWidthSelect(wrapper).find('input')
            await input.setValue('75%')
            await flushPromises()
            expect(draft.value!.container.maxWidth).toBe('75%')
        })

        it.each(['1200px', '75%'])('shows a stored non-preset value %s in the input', async value => {
            draft.value!.container.maxWidth = value
            const wrapper = await mountAppearance()
            const input = maxWidthSelect(wrapper).find('input').element as HTMLInputElement
            expect(input.value).toBe(value)
        })
    })

    describe('alignment', () => {
        it('replaces the two dropdowns with one picker writing both axes', async () => {
            draft.value!.container.verticalAlign = 'bottom'
            draft.value!.container.horizontalAlign = 'left'
            const wrapper = await mountAppearance()
            // Only theme, widget theme, colour mode and max width remain Selects.
            expect(selects(wrapper)).toHaveLength(4)

            const picker = wrapper.findComponent(AlignmentPicker)
            expect(picker.props('vertical')).toBe('bottom')
            expect(picker.props('horizontal')).toBe('left')

            await wrapper.find('[aria-label="Top right"]').trigger('click')
            expect(draft.value!.container.verticalAlign).toBe('top')
            expect(draft.value!.container.horizontalAlign).toBe('right')
        })

        it('names the picker by its visible Alignment heading', async () => {
            const wrapper = await mountAppearance()
            const heading = wrapper.findAll('span').find(s => s.text() === 'Alignment')!
            expect(heading.attributes('id')).toBeTruthy()
            expect(wrapper.find('[role="radiogroup"]').attributes('aria-labelledby')).toBe(heading.attributes('id'))
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
