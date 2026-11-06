import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const transport: WidgetModule = {
    type: 'transport',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Departures',
    icon: 'ti-bus',
    description: 'Public transport departures',
}

export default transport
