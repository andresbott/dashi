import type { Component } from 'vue'

export interface WidgetModule {
    type: string
    component: Component
    configComponent: Component | null
    label: string
    icon: string
    description: string
    noWidgetProp?: boolean
}
