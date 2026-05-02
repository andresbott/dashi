import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const weatherCompact: WidgetModule = {
    type: 'weather-compact',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Weather (Compact)',
    icon: 'ti-cloud',
    description: 'Compact weather display',
}

export default weatherCompact
