import type { InjectionKey, Ref } from 'vue'

export const DASHBOARD_THEME: InjectionKey<Ref<string>> = Symbol('dashboardTheme')
export const DASHBOARD_ID: InjectionKey<Ref<string>> = Symbol('dashboardId')
export const ACTIVE_PAGE: InjectionKey<Ref<number>> = Symbol('activePage')
export const TOTAL_PAGES: InjectionKey<Ref<number>> = Symbol('totalPages')
export const EDITING_MODE: InjectionKey<boolean> = Symbol('editing-mode')
