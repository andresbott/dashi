import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const weather: WidgetModule = {
    type: 'weather',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Weather',
    icon: 'ti-sun',
    description: 'Current conditions and forecast',
}

export default weather
