import { describe, it, expect } from 'vitest'
import { fromDraft, toDraft, type EditorBackground } from '@/lib/themeBackgroundDraft'
import type { ThemeBackground } from '@/types/theme'

const enums = { fit: 'cover', position: 'center', repeat: 'no-repeat' } as const

describe('themeBackgroundDraft', () => {
    it('maps an empty background to an empty draft and back', () => {
        expect(toDraft({ light: {}, dark: {} })).toEqual({})
        expect(toDraft(undefined)).toEqual({})
        expect(fromDraft({})).toEqual({ light: {}, dark: {} })
    })

    it('copies the light values into dark when the draft has no dark override', () => {
        const draft: EditorBackground = {
            gradient: { direction: '90deg', light: ['#000000', '#ffffff'] },
            image: { light: 'asset:a.png', ...enums },
        }
        expect(fromDraft(draft)).toEqual({
            light: {
                gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] },
                image: { ref: 'asset:a.png', ...enums },
            },
            dark: {
                gradient: { direction: '90deg', stops: ['#000000', '#ffffff'] },
                image: { ref: 'asset:a.png', ...enums },
            },
        })
    })

    it('applies dark stops and a dark image on top of the light copy', () => {
        const bg = fromDraft({
            gradient: { direction: '90deg', light: ['#000000', '#ffffff'], dark: ['#111111', '#222222'] },
            image: { light: 'asset:a.png', dark: 'asset:b.png', ...enums },
        })
        expect(bg.dark.gradient?.stops).toEqual(['#111111', '#222222'])
        expect(bg.dark.image?.ref).toBe('asset:b.png')
        expect(bg.light.gradient?.stops).toEqual(['#000000', '#ffffff'])
        expect(bg.light.image?.ref).toBe('asset:a.png')
    })

    it('does not alias the stop lists between modes or with the draft', () => {
        const draft: EditorBackground = { gradient: { direction: '90deg', light: ['#000000', '#ffffff'] } }
        const bg = fromDraft(draft)
        bg.dark.gradient!.stops.push('#ff0000')
        draft.gradient!.light.push('#00ff00')
        expect(bg.light.gradient?.stops).toEqual(['#000000', '#ffffff'])
    })

    it('attaches dark values to the draft only when they differ from light', () => {
        const same: ThemeBackground = {
            light: { gradient: { direction: '90deg', stops: ['#000', '#fff'] }, image: { ref: 'asset:a.png', ...enums } },
            dark: { gradient: { direction: '90deg', stops: ['#000', '#fff'] }, image: { ref: 'asset:a.png', ...enums } },
        }
        expect(toDraft(same)).toEqual({
            gradient: { direction: '90deg', light: ['#000', '#fff'] },
            image: { light: 'asset:a.png', ...enums },
        })

        const differing: ThemeBackground = {
            light: same.light,
            dark: { gradient: { direction: '90deg', stops: ['#111', '#222'] }, image: { ref: 'asset:b.png', ...enums } },
        }
        expect(toDraft(differing)).toEqual({
            gradient: { direction: '90deg', light: ['#000', '#fff'], dark: ['#111', '#222'] },
            image: { light: 'asset:a.png', dark: 'asset:b.png', ...enums },
        })
    })

    it('round-trips every draft the editor can produce', () => {
        const drafts: EditorBackground[] = [
            {},
            { gradient: { direction: '135deg', light: ['#667eea', '#764ba2'] } },
            { gradient: { direction: '135deg', light: ['#667eea', '#764ba2'], dark: ['#1e1b4b', '#0f172a'] } },
            { image: { light: 'asset:a.png', ...enums } },
            { image: { light: 'asset:a.png', dark: 'asset:b.png', fit: 'contain', position: 'top', repeat: 'repeat' } },
        ]
        for (const d of drafts) {
            expect(toDraft(fromDraft(d))).toEqual(d)
        }
    })
})
