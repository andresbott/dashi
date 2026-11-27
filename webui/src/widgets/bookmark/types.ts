export interface BookmarkWidgetConfig {
    url: string
    icon: string
    title: string
    subtitle: string
    textBelow?: boolean
    /** Open the link in a new tab; off (default) opens it in place. */
    newTab?: boolean
}
