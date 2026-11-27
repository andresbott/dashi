import type { InjectionKey, Ref } from 'vue'
import type { WidgetMove } from '@/lib/widgetMove'

// The dashboard's widget theme name. Provided only by the editor and read only
// by widget config forms (e.g. the clock font picker); the editor preview
// itself is theme-agnostic.
export const DASHBOARD_WIDGET_THEME: InjectionKey<Ref<string>> = Symbol('dashboardWidgetTheme')
export const DASHBOARD_ID: InjectionKey<Ref<string>> = Symbol('dashboardId')
export const ACTIVE_PAGE: InjectionKey<Ref<number>> = Symbol('activePage')
export const TOTAL_PAGES: InjectionKey<Ref<number>> = Symbol('totalPages')
export const EDITING_MODE: InjectionKey<boolean> = Symbol('editing-mode')
// Maps an uploaded image name to its URL. Only the widget docs provide it, to
// show sample images; the image widget falls back to the data API.
export const DATA_IMAGE_URL: InjectionKey<(name: string) => string> = Symbol('dataImageUrl')
// Commits a dragged widget's drop (row ↔ row, row ↔ stack, stack ↔ stack).
// Provided only by the editor, which applies the move to the page's rows in
// one assignment; rows and stacks call it with the move endWidgetDrag returns.
export const WIDGET_MOVE: InjectionKey<(move: WidgetMove) => void> = Symbol('widgetMove')
