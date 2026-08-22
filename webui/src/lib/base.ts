// Runtime base path for the SPA. The editor server injects
// window.__DASHI_BASE__ (always ending in "/") so the app works under Home
// Assistant's dynamic ingress prefix. Defaults to "/" for direct/dev serving.
export const appBase = (): string => {
  const b = (window as unknown as { __DASHI_BASE__?: string }).__DASHI_BASE__
  return b && b.length > 0 ? b : '/'
}

// withBase turns an absolute app path ("/api/v0") into a URL under the runtime
// base ("/api/hassio_ingress/<t>/api/v0"). The result keeps a leading slash,
// so it is an absolute URL that ignores any <base href>.
export const withBase = (path: string): string => appBase() + path.replace(/^\//, '')
