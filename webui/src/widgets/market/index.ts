import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const market: WidgetModule = {
    type: 'market',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Market',
    icon: 'ti-chart-line',
    description: 'Stock / crypto market ticker',
}

export default market
