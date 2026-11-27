export interface BookmarkGroupWidgetConfig {
    bookmarks: BookmarkGroupItem[]
    /** Shared by every bookmark: icon on top, text below. */
    textBelow?: boolean
    /** Shared by every bookmark: open links in a new tab. */
    newTab?: boolean
    /** A wrapping row (default) or a column. */
    direction?: BookmarkGroupDirection
    /**
     * Horizontal only: bookmarks per row before the next row starts, in
     * equal-width columns. 0 or absent fits as many as the width allows.
     */
    columns?: number
    /** Where the bookmarks come from: `bookmarks` above (default) or a content file. */
    source?: BookmarkGroupSource
    /** With source "file": the bookmarks file name, e.g. "links.yaml". */
    file?: string
}

export type BookmarkGroupSource = 'dashboard' | 'file'

/** Highest `columns` value; the server rejects anything above it. */
export const MAX_GROUP_COLUMNS = 12

export type BookmarkGroupDirection = 'horizontal' | 'vertical'

/** One bookmark of a group; same fields as a single bookmark minus the shared options. */
export interface BookmarkGroupItem {
    url: string
    icon: string
    title: string
    subtitle: string
}
