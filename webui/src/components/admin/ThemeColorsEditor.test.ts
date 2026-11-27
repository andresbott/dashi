import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import ColorPicker from 'primevue/colorpicker'
import InputText from 'primevue/inputtext'
import ThemeColorsEditor from './ThemeColorsEditor.vue'
import { PALETTE_ROLES } from '@/lib/themeCss'
import type { Palette } from '@/types/theme'

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}

// Mounts the editor as a parent would with v-model:palette: every emitted
// palette is fed back as the new prop.
function mountEditor(p: Palette, opts: { mode?: 'light' | 'dark'; disabled?: boolean } = {}) {
    const wrapper = mount(ThemeColorsEditor, {
        props: {
            mode: opts.mode ?? 'light',
            palette: p,
            disabled: opts.disabled ?? false,
            'onUpdate:palette': (v: Palette) => wrapper.setProps({ palette: v }),
        },
        global: { plugins: [PrimeVue] },
    })
    return wrapper
}

const lastPalette = (wrapper: ReturnType<typeof mountEditor>): Palette => {
    const events = wrapper.emitted('update:palette') as Palette[][]
    return events[events.length - 1][0]
}

describe('ThemeColorsEditor', () => {
    it('renders one labelled field per role, with ids scoped to the mode', () => {
        const wrapper = mountEditor({ ...palette }, { mode: 'dark' })
        for (const { role } of PALETTE_ROLES) {
            const input = wrapper.find<HTMLInputElement>(`#color-dark-${role}`)
            expect(input.exists()).toBe(true)
            expect(input.element.value).toBe(palette[role])
            expect(wrapper.find(`label[for="color-dark-${role}"]`).exists()).toBe(true)
        }
        expect(wrapper.find('#color-light-bg').exists()).toBe(false)
        expect(wrapper.find('label[for="color-dark-textMuted"]').text()).toBe('Muted text')
    })

    it('emits a new palette when a field is typed, leaving the prop untouched', async () => {
        const original = { ...palette }
        const wrapper = mountEditor(original)
        await wrapper.find('#color-light-text').setValue('#1a2b3c')
        expect(lastPalette(wrapper)).toEqual({ ...palette, text: '#1a2b3c' })
        expect(original.text).toBe('#333333')
    })

    it('highlights an invalid value, and an alpha page background', async () => {
        const wrapper = mountEditor({ ...palette })
        const text = wrapper.findAllComponents(InputText).find(c => c.attributes('id') === 'color-light-text')!
        expect(text.props('invalid')).toBe(false)

        await wrapper.find('#color-light-text').setValue('#1a2b')
        expect(text.props('invalid')).toBe(true)

        await wrapper.find('#color-light-bg').setValue('#ffffff80')
        const bg = wrapper.findAllComponents(InputText).find(c => c.attributes('id') === 'color-light-bg')!
        expect(bg.props('invalid')).toBe(true)

        await wrapper.find('#color-light-surface').setValue('#ffffff80')
        const surface = wrapper.findAllComponents(InputText).find(c => c.attributes('id') === 'color-light-surface')!
        expect(surface.props('invalid')).toBe(false)
    })

    it('shows the picker a six-digit value for #rgb, #rrggbbaa and unset roles', () => {
        const wrapper = mountEditor({ ...palette, bg: '#abc', surface: '#11223380', border: '' })
        const pickers = wrapper.findAllComponents(ColorPicker)
        expect(pickers).toHaveLength(PALETTE_ROLES.length)
        expect(pickers[0].props('modelValue')).toBe('aabbcc') // bg
        expect(pickers[1].props('modelValue')).toBe('112233') // surface
        expect(pickers[2].props('modelValue')).toBe('000000') // border
    })

    it('keeps a role alpha when picking, but writes the page background opaque', async () => {
        const wrapper = mountEditor({ ...palette, bg: '#ffffff', surface: '#f5f5f780' })
        const pickers = wrapper.findAllComponents(ColorPicker)

        pickers[1].vm.$emit('update:modelValue', '445566') // surface
        await wrapper.vm.$nextTick()
        expect(lastPalette(wrapper).surface).toBe('#44556680')

        pickers[0].vm.$emit('update:modelValue', '112233') // bg
        await wrapper.vm.$nextTick()
        expect(lastPalette(wrapper).bg).toBe('#112233')
        expect(lastPalette(wrapper).surface).toBe('#44556680')
    })

    it('ignores an undefined picker value', async () => {
        const wrapper = mountEditor({ ...palette })
        wrapper.findAllComponents(ColorPicker)[0].vm.$emit('update:modelValue', undefined)
        await wrapper.vm.$nextTick()
        expect(wrapper.emitted('update:palette')).toBeUndefined()
    })

    it('disables every control and emits nothing when disabled', async () => {
        const wrapper = mountEditor({ ...palette }, { disabled: true })
        expect(wrapper.find<HTMLInputElement>('#color-light-text').element.disabled).toBe(true)
        for (const picker of wrapper.findAllComponents(ColorPicker)) {
            expect(picker.props('disabled')).toBe(true)
        }
        wrapper.findAllComponents(ColorPicker)[0].vm.$emit('update:modelValue', '112233')
        await wrapper.vm.$nextTick()
        expect(wrapper.emitted('update:palette')).toBeUndefined()
    })
})
