import axios from 'axios'
import { withBase } from '@/lib/base'

export const apiClient = axios.create({
    baseURL: withBase('/api/v0'),
    headers: {
        'Content-Type': 'application/json'
    }
})

// apiErrorMessage returns the server's error text from a failed API call, or
// undefined when there is none. Handlers answer {"error": "..."}, but the
// server's error middleware wraps that whole body as the string value of its
// own {"error": ...}, so the message may sit one JSON level down.
export function apiErrorMessage(err: unknown): string | undefined {
    const outer = (err as { response?: { data?: { error?: unknown } } })?.response?.data?.error
    if (typeof outer !== 'string' || outer === '') return undefined
    try {
        const inner = (JSON.parse(outer) as { error?: unknown })?.error
        if (typeof inner === 'string' && inner !== '') return inner
    } catch {
        // not JSON: the outer string is the message
    }
    return outer
}
