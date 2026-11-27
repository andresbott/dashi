export interface BookmarkGroupWidgetConfig {
    bookmarks: BookmarkGroupItem[]
    /** Shared by every bookmark: icon on top, text below. */
    textBelow?: boolean
    /** Shared by every bookmark: open links in a new tab. */
    newTab?: boolean
    /** A wrapping row (default) or a column. */
    direction?: BookmarkGroupDirection
}

export type BookmarkGroupDirection = 'horizontal' | 'vertical'

/** One bookmark of a group; same fields as a single bookmark minus the shared options. */
export interface BookmarkGroupItem {
    url: string
    icon: string
    title: string
    subtitle: string
}
