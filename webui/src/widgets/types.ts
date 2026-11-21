import type { Component } from 'vue'
import type { WidgetCategory } from './categories'

export interface WidgetModule {
    type: string
    component: Component
    configComponent: Component | null
    label: string
    icon: string
    description: string
    // Group the widget is listed under (see ./categories.ts).
    category: WidgetCategory
    noWidgetProp?: boolean
    // Width of the Widget Settings dialog: 'wide' (60rem) for forms laid out
    // in columns; omitted keeps the default narrow dialog (28rem).
    settingsSize?: 'wide'
}
