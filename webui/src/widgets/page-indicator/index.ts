import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const pageIndicator: WidgetModule = {
    type: 'page-indicator',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: null,
    label: 'Page Indicator',
    icon: 'ti-circles',
    description: 'Shows dots for each page',
    category: 'layout',
    noWidgetProp: true,
}

export default pageIndicator
