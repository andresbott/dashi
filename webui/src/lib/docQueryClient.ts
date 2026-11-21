import { onUnmounted, provide } from 'vue'
import { QueryClient, VUE_QUERY_CLIENT } from '@tanstack/vue-query'
import type { DocQueryMock } from './widgetDocs'

// A QueryClient for the widget documentation pages. Widgets fetch through
// TanStack Query; this client answers every query whose key starts with a
// mocked prefix (['weather', ...], ['market', ...]) from the canned payloads,
// and never refetches them, so previews need no network and look the same
// every time. Queries without a mock (widget themes, icon lookups) run their
// real queryFn against the local API.
export function createDocQueryClient(mocks: Record<string, DocQueryMock>): QueryClient {
    const client = new QueryClient({
        defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } },
    })
    const defaults = client.defaultQueryOptions.bind(client)
    // Per-query options (queryFn, refetchInterval) override client defaults,
    // so the mock has to be applied after they are merged, on every observer.
    client.defaultQueryOptions = ((options) => {
        const resolved = defaults(options)
        const mock = mocks[String(resolved.queryKey?.[0])]
        if (!mock) return resolved
        return {
            ...resolved,
            queryFn: ({ queryKey }: { queryKey: readonly unknown[] }) => {
                const data = mock(queryKey)
                if (data === undefined) throw new Error(`no doc fixture for ${JSON.stringify(queryKey)}`)
                return data
            },
            staleTime: Infinity,
            refetchInterval: false,
        }
    }) as typeof client.defaultQueryOptions
    return client
}

// Provides a mock client to the calling component's descendants (previews and
// settings forms), replacing the app client for that subtree only.
export function provideDocQueryClient(mocks: Record<string, DocQueryMock>): QueryClient {
    const client = createDocQueryClient(mocks)
    provide(VUE_QUERY_CLIENT, client)
    onUnmounted(() => client.clear())
    return client
}
