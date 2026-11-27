import { describe, it, expect } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AlignmentPicker from './AlignmentPicker.vue'
import type { Container } from '@/types/dashboard'

// Plain strings: a hand-edited dashboard.json can hold anything.
const mountPicker = (vertical?: string, horizontal?: string) =>
    mount(AlignmentPicker, {
        props: {
            vertical: vertical as Container['verticalAlign'] | undefined,
            horizontal: horizontal as Container['horizontalAlign'] | undefined,
            labelledBy: 'alignment-heading',
        },
        attachTo: document.body,
    })

const options = (wrapper: ReturnType<typeof mountPicker>) => wrapper.findAll('[data-testid="alignment-option"]')
const checked = (wrapper: ReturnType<typeof mountPicker>) =>
    options(wrapper).filter(o => o.attributes('aria-checked') === 'true')

describe('AlignmentPicker', () => {
    it('shows a 3x3 radio group, named by the visible heading, in reading order', () => {
        const wrapper = mountPicker('top', 'left')
        const group = wrapper.find('[role="radiogroup"]')
        expect(group.attributes('aria-labelledby')).toBe('alignment-heading')
        expect(group.attributes('aria-label')).toBeUndefined()
        const buttons = options(wrapper)
        expect(buttons.map(b => b.element.tagName)).toEqual(Array(9).fill('BUTTON'))
        expect(buttons.every(b => b.attributes('role') === 'radio')).toBe(true)
        expect(buttons.every(b => b.attributes('aria-pressed') === undefined)).toBe(true)
        expect(buttons.every(b => b.attributes('type') === 'button')).toBe(true)
        expect(buttons.map(b => b.attributes('aria-label'))).toEqual([
            'Top left', 'Top center', 'Top right',
            'Center left', 'Center', 'Center right',
            'Bottom left', 'Bottom center', 'Bottom right',
        ])
    })

    it('marks only the stored combination as checked', () => {
        const wrapper = mountPicker('bottom', 'right')
        expect(options(wrapper).map(o => o.attributes('aria-checked')).filter(v => v === 'false')).toHaveLength(8)
        expect(checked(wrapper).map(o => o.attributes('aria-label'))).toEqual(['Bottom right'])
    })

    it('shows the server defaults (top, centered) for missing or unknown values', () => {
        expect(checked(mountPicker(undefined, undefined)).map(o => o.attributes('aria-label'))).toEqual(['Top center'])
        expect(checked(mountPicker('middle', 'sideways')).map(o => o.attributes('aria-label'))).toEqual(['Top center'])
    })

    it('writes both axes when a cell is picked', async () => {
        const wrapper = mountPicker('top', 'left')
        await options(wrapper)[5].trigger('click') // Center right
        expect(wrapper.emitted('update:vertical')).toEqual([['center']])
        expect(wrapper.emitted('update:horizontal')).toEqual([['right']])
    })

    it('draws each cell as a diagram placing the box in its position', () => {
        const wrapper = mountPicker('top', 'left')
        const style = (i: number) => (options(wrapper)[i].find('.alignment-box').element as HTMLElement).style
        expect(style(0).justifyContent).toBe('flex-start')
        expect(style(0).alignItems).toBe('flex-start')
        expect(style(4).justifyContent).toBe('center')
        expect(style(4).alignItems).toBe('center')
        expect(style(8).justifyContent).toBe('flex-end')
        expect(style(8).alignItems).toBe('flex-end')
    })

    it('keeps one tab stop on the current cell and moves focus with the arrow keys', async () => {
        const wrapper = mountPicker('center', 'center')
        const buttons = options(wrapper)
        expect(buttons.map(b => b.attributes('tabindex'))).toEqual(['-1', '-1', '-1', '-1', '0', '-1', '-1', '-1', '-1'])

        const focused = () => document.activeElement?.getAttribute('aria-label')
        await buttons[4].trigger('keydown', { key: 'ArrowRight' })
        expect(focused()).toBe('Center right')
        await buttons[5].trigger('keydown', { key: 'ArrowUp' })
        expect(focused()).toBe('Top right')
        // Edges stop rather than wrap.
        await buttons[2].trigger('keydown', { key: 'ArrowRight' })
        expect(focused()).toBe('Top right')
        await buttons[2].trigger('keydown', { key: 'ArrowDown' })
        await buttons[5].trigger('keydown', { key: 'ArrowLeft' })
        expect(focused()).toBe('Center')
        // Non-arrow keys do nothing.
        await buttons[4].trigger('keydown', { key: 'a' })
        expect(focused()).toBe('Center')
        wrapper.unmount()
    })

    it('selects the cell an arrow key moves to and makes it the focused tab stop', async () => {
        const wrapper = mountPicker('top', 'center')
        await options(wrapper)[1].trigger('keydown', { key: 'ArrowRight' })
        expect(wrapper.emitted('update:horizontal')).toEqual([['right']])
        expect(wrapper.emitted('update:vertical')).toBeUndefined() // unchanged axis

        await wrapper.setProps({ vertical: 'top', horizontal: 'right' })
        await flushPromises()
        const buttons = options(wrapper)
        expect(document.activeElement?.getAttribute('aria-label')).toBe('Top right')
        expect(checked(wrapper).map(o => o.attributes('aria-label'))).toEqual(['Top right'])
        expect(buttons.map(b => b.attributes('tabindex'))).toEqual(['-1', '-1', '0', '-1', '-1', '-1', '-1', '-1', '-1'])
        wrapper.unmount()
    })
})
