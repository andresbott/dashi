import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const battery: WidgetModule = {
    type: 'battery',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: null,
    label: 'Battery',
    icon: 'ti-battery-2',
    description: 'Battery status from query parameter',
}

export default battery
