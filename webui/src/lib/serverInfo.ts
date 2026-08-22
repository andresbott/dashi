import { ref } from 'vue'
import { getServerInfo, type ServerInfo } from '@/lib/api/info'

// This SPA is the admin/editor UI only — dashboards are rendered by the Go
// viewer server, which listens on its own port (and may sit behind a different
// host). The backend tells us where that is; we ask once at app init and keep
// the answer in a ref so links become absolute as soon as it arrives.
const serverInfo = ref<ServerInfo | null>(null)

export const loadServerInfo = async (): Promise<void> => {
    try {
        serverInfo.value = await getServerInfo()
    } catch {
        // Not fatal: dashboard links fall back to same-origin paths, which is
        // what an editor-only deployment serves anyway.
        serverInfo.value = null
    }
}

// publicViewerUrl is the base URL of the public viewer, or '' when the backend
// did not advertise one (viewer disabled, old backend, request failed).
export const publicViewerUrl = (): string => {
    const viewer = serverInfo.value?.viewer
    return viewer?.enabled ? viewer.url : ''
}

// dashboardViewUrl is the link to the server-rendered view of a dashboard.
export const dashboardViewUrl = (id: string): string => {
    return `${publicViewerUrl()}/${encodeURIComponent(id)}`
}

// resetServerInfo exists for tests.
export const resetServerInfo = (): void => {
    serverInfo.value = null
}
