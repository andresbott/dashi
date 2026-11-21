// Bookmark URL rule — the editor side of validateURL in
// internal/widgets/bookmark/bookmark.go. The server rejects (400) any save
// whose bookmark url fails it; internal/widgets/bookmark/testdata/url_cases.json
// is the case table both sides are tested against.
//
// Allowed: empty (not configured yet), http:// or https:// (any case, no
// leading whitespace), or a single "/" path on this server. Rejected: other
// schemes (javascript:, data:, ...), protocol-relative "//host" and "/\host",
// scheme-less text like "example.com", and ASCII control characters anywhere
// (browsers strip tabs/newlines, turning "/\t/host" into "//host").

export const BOOKMARK_URL_ERROR = 'URL must start with http://, https:// or / (a path on this server)'

const CONTROL_CHARS = /[\u0000-\u001f\u007f]/

export function isSafeBookmarkUrl(url: string): boolean {
    if (url === '') return true
    if (CONTROL_CHARS.test(url)) return false
    const lower = url.toLowerCase()
    if (lower.startsWith('http://') || lower.startsWith('https://')) return true
    return url.startsWith('/') && !url.startsWith('//') && !url.startsWith('/\\')
}

// The href to bind for a stored URL: the URL itself when it passes the rule,
// otherwise undefined so the link renders without an href (inert). Stored
// dashboards may predate the server-side check.
export function safeBookmarkHref(url: string | undefined): string | undefined {
    if (!url || !isSafeBookmarkUrl(url)) return undefined
    return url
}
