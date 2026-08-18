import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const search: WidgetModule = {
    type: 'search',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Search',
    icon: 'ti-search',
    description: 'Search engine input',
}

export default search
