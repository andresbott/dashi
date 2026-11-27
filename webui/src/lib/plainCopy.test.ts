import { describe, it, expect } from 'vitest'
import { reactive, isProxy } from 'vue'
import { plainCopy } from './plainCopy'

describe('plainCopy', () => {
    it('deep-copies data holding reactive proxies below the top level', () => {
        const child = reactive({ id: 'c1', config: { format: '24h' } })
        const original = { widgets: [child], tags: ['a'] }

        const copy = plainCopy(original)

        expect(copy).toEqual({ widgets: [{ id: 'c1', config: { format: '24h' } }], tags: ['a'] })
        expect(copy).not.toBe(original)
        expect(copy.widgets[0]).not.toBe(child)
        expect(isProxy(copy.widgets[0])).toBe(false)

        copy.widgets[0].config.format = '12h'
        expect(child.config.format).toBe('24h')
    })
})
